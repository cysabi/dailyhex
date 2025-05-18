package main

import (
	"context"
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"
	"net"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/charmbracelet/log"
	"github.com/charmbracelet/ssh"
	"github.com/charmbracelet/wish"
	"github.com/charmbracelet/wish/activeterm"
	"github.com/charmbracelet/wish/bubbletea"
	"github.com/charmbracelet/wish/logging"
	"github.com/creack/pty"
	"github.com/muesli/termenv"
	"github.com/tidwall/buntdb"
	gosh "golang.org/x/crypto/ssh"
)

const (
	host = "0.0.0.0"
	port = "22"
)

type state struct {
	db            *buntdb.DB
	day           int64
	dayPage       int64
	secret        string
	playerid      string
	height        int
	width         int
	showCountdown bool
	gameState     GameState
	screen        Screen
	styles        Styles
}

type GameState string

const (
	Idle    GameState = "8"
	Invalid GameState = "9"
	Win     GameState = "10"
)

type Screen string

const (
	TitleScreen Screen = "back to title"
	PlayScreen  Screen = "play today!"
	BoardScreen Screen = "see leaderboard"
	HelpScreen  Screen = "help"
)

func main() {
	if err := os.MkdirAll("store", 0744); err != nil {
		log.Fatal(err)
	}
	if _, err := os.Stat("store/color_secret"); os.IsNotExist(err) {
		secret := make([]byte, 32)
		rand.Read(secret)
		if err := os.WriteFile("store/color_secret", secret, 0744); err != nil {
			log.Fatal(err)
		}
	}

	db, err := buntdb.Open("store/data.db")
	if err != nil {
		log.Fatal(err)
	}

	s, err := wish.NewServer(
		wish.WithAddress(net.JoinHostPort(host, port)),
		wish.WithHostKeyPath("store/ssh_id_ed25519"),
		wish.WithPublicKeyAuth(func(ctx ssh.Context, key ssh.PublicKey) bool {
			return true
		}),
		wish.WithPasswordAuth(func(ctx ssh.Context, password string) bool {
			return true
		}),
		wish.WithMiddleware(
			bubbletea.Middleware(func(s ssh.Session) (tea.Model, []tea.ProgramOption) {
				clientOutput := outputFromSession(s)
				pty, _, _ := s.Pty()
				ren := bubbletea.MakeRenderer(s)
				ren.SetOutput(clientOutput)

				if s.PublicKey() == nil {
					wish.Println(s, ren.NewStyle().Foreground(lipgloss.Color("5")).BorderForeground(lipgloss.Color("5")).Border(lipgloss.OuterHalfBlockBorder(), false, false, false, true).PaddingLeft(2).Render(
						lipgloss.JoinVertical(0,
							" ",
							"welcome!! to play, first make a public key with",
							" ",
							ren.NewStyle().Bold(true).Render("ssh-keygen -t ed25519"),
							" ",
							"i just need this to keep track of your identity",
							"have fun!",
							" ")))
					s.Exit(1)
					return nil, nil
				}

				day := day()
				secret := secret(day)
				playerId := string(gosh.MarshalAuthorizedKey(s.PublicKey()))

				state := state{
					db:            db,
					day:           day,
					secret:        secret,
					playerid:      playerId,
					height:        pty.Window.Height,
					width:         pty.Window.Width,
					showCountdown: false,
					gameState:     Idle,
					screen:        TitleScreen,
					styles:        Styles{}.New(ren, secret),
				}
				if state.GetDone() {
					state.gameState = Win
				}

				m := Model{state: &state}.New()
				return m, []tea.ProgramOption{tea.WithAltScreen()}
			}),
			activeterm.Middleware(),
			logging.Middleware(),
		),
	)
	if err != nil {
		log.Error("Could not start server", "error", err)
	}

	done := make(chan os.Signal, 1)
	signal.Notify(done, os.Interrupt, syscall.SIGINT, syscall.SIGTERM)
	log.Info("Starting SSH server", "host", host, "port", port)
	go func() {
		if err = s.ListenAndServe(); err != nil && !errors.Is(err, ssh.ErrServerClosed) {
			log.Error("Could not start server", "error", err)
			done <- nil
		}
	}()

	<-done
	log.Info("Stopping SSH server")
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer func() { cancel() }()
	if err := s.Shutdown(ctx); err != nil && !errors.Is(err, ssh.ErrServerClosed) {
		log.Error("Could not stop server", "error", err)
	}
}

// Bridge Wish and Termenv so we can query for a user's terminal capabilities.
type sshOutput struct {
	ssh.Session
	tty *os.File
}

func (s *sshOutput) Write(p []byte) (int, error) {
	return s.Session.Write(p)
}

func (s *sshOutput) Read(p []byte) (int, error) {
	return s.Session.Read(p)
}

func (s *sshOutput) Fd() uintptr {
	return s.tty.Fd()
}

type sshEnviron struct {
	environ []string
}

func (s *sshEnviron) Getenv(key string) string {
	for _, v := range s.environ {
		if strings.HasPrefix(v, key+"=") {
			return v[len(key)+1:]
		}
	}
	return ""
}

func (s *sshEnviron) Environ() []string {
	return s.environ
}

func outputFromSession(sess ssh.Session) *termenv.Output {
	sshPty, _, _ := sess.Pty()
	_, tty, err := pty.Open()
	if err != nil {
		log.Fatal(err)
	}
	o := &sshOutput{
		Session: sess,
		tty:     tty,
	}
	environ := sess.Environ()
	environ = append(environ, fmt.Sprintf("TERM=%s", sshPty.Term))
	e := &sshEnviron{environ: environ}
	// We need to use unsafe mode here because the ssh session is not running
	// locally and we already know that the session is a TTY.
	return termenv.NewOutput(o, termenv.WithUnsafe(), termenv.WithEnvironment(e))
}

func day() int64 {
	loc, _ := time.LoadLocation("America/New_York")
	now := time.Now().In(loc)
	shifted := now.Add(-11 * time.Hour)
	midnight := time.Date(shifted.Year(), shifted.Month(), shifted.Day(), 0, 0, 0, 0, loc)
	return midnight.Unix() / 86400
}

func secret(day int64) string {
	file, err := os.ReadFile("store/color_secret")
	if err != nil {
		log.Fatal(err)
	}
	hash := sha256.Sum256(fmt.Append(file, day))
	return hex.EncodeToString(hash[:3])
}

package main

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"
	"net"
	"os"
	"os/signal"
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
	Idle    GameState = "0"
	Invalid GameState = "9"
	Win     GameState = "10"
)

type Screen string

const (
	TitleScreen Screen = "back to title"
	PlayScreen  Screen = "play today!"
	BoardScreen Screen = "see leaderboard"
)

func main() {
	_ = os.Mkdir("store", 0744)
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
			(func() wish.Middleware {

				newProg := func(m tea.Model, opts ...tea.ProgramOption) *tea.Program {
					p := tea.NewProgram(m, opts...)
					return p
				}
				teaHandler := func(s ssh.Session) *tea.Program {
					pty, _, _ := s.Pty()
					ren := bubbletea.MakeRenderer(s)

					if s.PublicKey() == nil {
						wish.Println(s, ren.NewStyle().Foreground(lipgloss.Color("05")).BorderForeground(lipgloss.Color("05")).Border(lipgloss.OuterHalfBlockBorder(), false, false, false, true).PaddingLeft(2).Render(
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
						return nil
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

					return newProg(m, append(bubbletea.MakeOptions(s), tea.WithAltScreen())...)
				}
				return bubbletea.MiddlewareWithProgramHandler(teaHandler, termenv.TrueColor)
			})(),
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

func day() int64 {
	loc, _ := time.LoadLocation("America/New_York")
	now := time.Now().In(loc)
	shifted := now.Add(-11 * time.Hour)
	midnight := time.Date(shifted.Year(), shifted.Month(), shifted.Day(), 0, 0, 0, 0, loc)
	return midnight.Unix() / 86400
}

func secret(day int64) string {
	file, err := os.ReadFile("store/ssh_id_ed25519")
	if err != nil {
		log.Fatal(err)
	}
	hash := sha256.Sum256(fmt.Append(file, day))
	return hex.EncodeToString(hash[:3])
}

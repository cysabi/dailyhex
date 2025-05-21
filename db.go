package main

import (
	"fmt"
	"strings"

	"github.com/muesli/termenv"
	"github.com/tidwall/buntdb"
)

// board
func (s state) GetBoard() [][]string {
	var out [][]string

	key := fmt.Sprint(s.day+s.dayPage) + ":" + "*" + ":" + "moves"
	s.db.View(func(tx *buntdb.Tx) error {
		return tx.AscendKeys(key, func(key, moves string) bool {
			pid := strings.Split(key, ":")[1]

			name, _ := tx.Get(fmt.Sprint(pid) + ":" + "name")
			done, _ := tx.Get(fmt.Sprint(s.day+s.dayPage) + ":" + pid + ":" + "done")

			out = append(out, []string{name, moves, done})
			return true
		})
	})

	return out
}

// moves
func (s state) GetMoves() []string {
	var out string
	key := fmt.Sprint(s.day) + ":" + fmt.Sprint(s.playerid) + ":" + "moves"
	s.db.View(func(tx *buntdb.Tx) error {
		val, _ := tx.Get(key)
		out = val
		return nil
	})

	if len(out) == 0 {
		return make([]string, 0)
	}
	return strings.Split(out, ",")
}

func (s state) AppendMove(move string) error {
	key := fmt.Sprint(s.day) + ":" + fmt.Sprint(s.playerid) + ":" + "moves"
	return s.db.Update(func(tx *buntdb.Tx) error {
		val, _ := tx.Get(key)

		newVal := val + "," + move
		if len(val) == 0 {
			newVal = move
		}

		_, _, err := tx.Set(key, newVal, nil)
		return err
	})
}

// done
func (s state) GetDone() bool {
	var out string
	key := fmt.Sprint(s.day) + ":" + fmt.Sprint(s.playerid) + ":" + "done"
	s.db.View(func(tx *buntdb.Tx) error {
		val, _ := tx.Get(key)
		out = val
		return nil
	})

	return out == "true"
}

func (s state) SetDone(done bool) error {
	key := fmt.Sprint(s.day) + ":" + fmt.Sprint(s.playerid) + ":" + "done"
	return s.db.Update(func(tx *buntdb.Tx) error {
		_, _, err := tx.Set(key, fmt.Sprint(done), nil)
		return err
	})
}

// name
func (s state) GetName() string {
	var out string
	key := fmt.Sprint(s.playerid) + ":" + "name"
	s.db.View(func(tx *buntdb.Tx) error {
		val, _ := tx.Get(key)
		out = val
		return nil
	})

	return out
}

func (s state) SetName(name string) error {
	key := fmt.Sprint(s.playerid) + ":" + "name"
	return s.db.Update(func(tx *buntdb.Tx) error {
		_, _, err := tx.Set(key, name, nil)
		return err
	})
}

// color profile
func (s state) GetForceProfile() bool {
	var out string
	key := fmt.Sprint(s.playerid) + ":" + "force"
	s.db.View(func(tx *buntdb.Tx) error {
		val, _ := tx.Get(key)
		out = val
		return nil
	})

	return out == "true"
}

func (s *state) ToggleForceProfile() error {
	key := fmt.Sprint(s.playerid) + ":" + "force"
	return s.db.Update(func(tx *buntdb.Tx) error {
		val, _ := tx.Get(key)

		newVal := !(val == "true")

		if newVal {
			s.ren.SetColorProfile(termenv.TrueColor)
		} else {
			s.ren.SetColorProfile(s.sessProfile)
		}

		_, _, err := tx.Set(key, fmt.Sprint(newVal), nil)
		return err
	})
}

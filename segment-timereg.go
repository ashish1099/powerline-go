package main

import (
	"encoding/json"
	"fmt"
	"math"
	"os"
	"strings"
	"time"

	pwl "github.com/justjanne/powerline-go/powerline"
)

type timeregStatus struct {
	Active *timeregState  `json:"active"`
	Paused []timeregState `json:"paused"`
}

type timeregState struct {
	Issue     int             `json:"issue"`
	Repo      string          `json:"repo"`
	StartedAt time.Time       `json:"started_at"`
	Breaks    []timeregBreak  `json:"breaks,omitempty"`
	Pending   *timeregPending `json:"pending_break,omitempty"`
}

type timeregBreak struct {
	Start time.Time `json:"start"`
	End   time.Time `json:"end"`
}

type timeregPending struct {
	IdleSince time.Time `json:"idle_since"`
}

func segmentTimereg(p *powerline) []pwl.Segment {
	path := timeregStatePath()
	data, err := os.ReadFile(path)
	if err != nil {
		return []pwl.Segment{}
	}

	var state timeregStatus
	if err := json.Unmarshal(data, &state); err != nil {
		return []pwl.Segment{}
	}

	activeState := state.Active
	if activeState == nil {
		return []pwl.Segment{}
	}

	elapsed := timeregElapsed(activeState)
	repo := strings.TrimPrefix(activeState.Repo, "gitea-")
	content := fmt.Sprintf("#%d %s | %s", activeState.Issue, repo, timeregFormatDuration(elapsed))

	bg := p.theme.TimeregBg
	if activeState.Pending != nil {
		bg = p.theme.TimeregIdleBg
	}

	return []pwl.Segment{{
		Name:       "timereg",
		Content:    content,
		Foreground: p.theme.TimeregFg,
		Background: bg,
	}}
}

func timeregStatePath() string {
	if dir := os.Getenv("XDG_STATE_HOME"); dir != "" {
		return dir + "/timereg/tracking.json"
	}
	home, _ := os.UserHomeDir()
	return home + "/.local/state/timereg/tracking.json"
}

func timeregElapsed(s *timeregState) int {
	wall := time.Since(s.StartedAt).Minutes()
	breakMin := 0.0
	for _, b := range s.Breaks {
		breakMin += b.End.Sub(b.Start).Minutes()
	}
	return int(math.Round(wall - breakMin))
}

func timeregFormatDuration(minutes int) string {
	if minutes < 0 {
		minutes = 0
	}
	h := minutes / 60
	m := minutes % 60
	if h > 0 {
		return fmt.Sprintf("%dh%02dm", h, m)
	}
	return fmt.Sprintf("%dm", m)
}

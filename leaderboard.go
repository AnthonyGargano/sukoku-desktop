package main

import (
	"encoding/json"
	"os"
	"path/filepath"
	"sort"
	"time"
)

type Score struct {
	Seconds    int       `json:"seconds"`
	Errors     int       `json:"errors"`
	FinishedAt time.Time `json:"finished_at"`
}

type Leaderboard map[string][]Score // "Easy", "Medium", "Hard"

func lbPath() string {
	exe, err := os.Executable()
	if err == nil {
		return filepath.Join(filepath.Dir(exe), "sudoku_scores.json")
	}
	return "sudoku_scores.json"
}

func loadLB() Leaderboard {
	var lb Leaderboard
	data, err := os.ReadFile(lbPath())
	if err != nil {
		return Leaderboard{"Easy": {}, "Medium": {}, "Hard": {}}
	}
	if json.Unmarshal(data, &lb) != nil {
		return Leaderboard{"Easy": {}, "Medium": {}, "Hard": {}}
	}
	if lb["Easy"] == nil {
		lb["Easy"] = []Score{}
	}
	if lb["Medium"] == nil {
		lb["Medium"] = []Score{}
	}
	if lb["Hard"] == nil {
		lb["Hard"] = []Score{}
	}
	return lb
}

func saveLB(lb Leaderboard) {
	b, err := json.MarshalIndent(lb, "", "  ")
	if err != nil {
		return
	}
	_ = os.WriteFile(lbPath(), b, 0600)
}

func addScore(lb Leaderboard, diff string, s Score) {
	arr := append(lb[diff], s)
	sort.Slice(arr, func(i, j int) bool {
		if arr[i].Seconds == arr[j].Seconds {
			return arr[i].Errors < arr[j].Errors
		}
		return arr[i].Seconds < arr[j].Seconds
	})
	if len(arr) > 5 {
		arr = arr[:5]
	}
	lb[diff] = arr
}

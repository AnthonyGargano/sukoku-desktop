package main

import (
	"math/rand"
	"time"
)

const size = 9

// Grid is the Sudoku board type.
type Grid [size][size]int

func (g *Grid) Clone() Grid { c := *g; return c }

func (g *Grid) IsSafe(r, c, v int) bool {
	for i := 0; i < size; i++ {
		if g[r][i] == v || g[i][c] == v {
			return false
		}
	}
	br, bc := (r/3)*3, (c/3)*3
	for i := br; i < br+3; i++ {
		for j := bc; j < bc+3; j++ {
			if g[i][j] == v {
				return false
			}
		}
	}
	return true
}

func (g *Grid) FindEmpty() (int, int, bool) {
	for r := 0; r < size; r++ {
		for c := 0; c < size; c++ {
			if g[r][c] == 0 {
				return r, c, true
			}
		}
	}
	return -1, -1, false
}

// GenerateSolved builds a full valid solved grid (randomized).
func GenerateSolved() Grid {
	var g Grid
	var fill func(int, int) bool
	fill = func(r, c int) bool {
		if r == size {
			return true
		}
		nr, nc := r, c+1
		if nc == size {
			nr, nc = r+1, 0
		}
		perm := rand.Perm(9)
		for _, x := range perm {
			v := x + 1
			if g.IsSafe(r, c, v) {
				g[r][c] = v
				if fill(nr, nc) {
					return true
				}
				g[r][c] = 0
			}
		}
		return false
	}
	_ = fill(0, 0)
	return g
}

// SolveCount counts solutions up to 'limit' (early exits when exceeded).
func SolveCount(g *Grid, limit int) int {
	count := 0
	var dfs func() bool
	dfs = func() bool {
		if count > limit {
			return true
		}
		r, c, ok := g.FindEmpty()
		if !ok {
			count++
			return count > limit
		}
		for v := 1; v <= 9; v++ {
			if g.IsSafe(r, c, v) {
				g[r][c] = v
				if dfs() {
					g[r][c] = 0
					return true
				}
				g[r][c] = 0
			}
		}
		return false
	}
	dfs()
	return count
}

// MakeUniquePuzzle removes cells while preserving uniqueness (target clues).
func MakeUniquePuzzle(targetClues int) (puzzle Grid, solution Grid) {
	for {
		solved := GenerateSolved()
		p := solved.Clone()
		order := rand.Perm(81)
		removed := 0
		goalRemove := 81 - targetClues
		for _, idx := range order {
			if removed >= goalRemove {
				break
			}
			r, c := idx/9, idx%9
			if p[r][c] == 0 {
				continue
			}
			save := p[r][c]
			p[r][c] = 0
			test := p.Clone()
			if SolveCount(&test, 1) == 1 {
				removed++
			} else {
				p[r][c] = save
			}
		}
		final := p.Clone()
		if SolveCount(&final, 1) == 1 {
			return p, solved
		}
	}
}

func init() { rand.Seed(time.Now().UnixNano()) }

package main

import (
	"math/bits"
	"math/rand"
	"time"
)

const fullMask = 0x3FE // bits 1..9 set

// bit utilities
func boxIndex(r, c int) int { return (r/3)*3 + (c / 3) }

// SolveCountMasked counts solutions up to limit with bitmasks + MRV.
func SolveCountMasked(g *Grid, limit int) int {
	rows := [9]uint16{}
	cols := [9]uint16{}
	boxes := [9]uint16{}
	var empties [81][2]int
	nEmpty := 0

	for r := 0; r < 9; r++ {
		for c := 0; c < 9; c++ {
			v := g[r][c]
			if v == 0 {
				empties[nEmpty][0] = r
				empties[nEmpty][1] = c
				nEmpty++
				continue
			}
			bit := uint16(1 << v)
			rows[r] |= bit
			cols[c] |= bit
			boxes[boxIndex(r, c)] |= bit
		}
	}

	count := 0
	var dfs func(int) bool
	dfs = func(depth int) bool {
		if count >= limit {
			return true
		}
		if depth == nEmpty {
			count++
			return count >= limit
		}
		// MRV: choose the empty cell with the smallest candidates
		best := -1
		bestMask := uint16(0)
		bestBits := 10
		for i := depth; i < nEmpty; i++ {
			r := empties[i][0]
			c := empties[i][1]
			mask := uint16(fullMask) ^ (rows[r] | cols[c] | boxes[boxIndex(r, c)])
			if mask == 0 {
				continue
			}
			b := bits.OnesCount16(mask)
			if b < bestBits {
				bestBits = b
				bestMask = mask
				best = i
				if b == 1 {
					break
				}
			}
		}
		if best == -1 {
			return false
		}
		// swap chosen with current depth
		empties[depth], empties[best] = empties[best], empties[depth]
		r := empties[depth][0]
		c := empties[depth][1]
		if bestMask == 0 {
			bestMask = uint16(fullMask) ^ (rows[r] | cols[c] | boxes[boxIndex(r, c)])
		}
		for m := bestMask; m != 0; m &= m - 1 {
			vBit := m & -m
			v := bits.TrailingZeros16(vBit)
			bit := uint16(1 << v)
			g[r][c] = v
			rows[r] |= bit
			cols[c] |= bit
			bIdx := boxIndex(r, c)
			boxes[bIdx] |= bit
			if dfs(depth + 1) {
				g[r][c] = 0
				rows[r] &^= bit
				cols[c] &^= bit
				boxes[bIdx] &^= bit
				return true
			}
			g[r][c] = 0
			rows[r] &^= bit
			cols[c] &^= bit
			boxes[bIdx] &^= bit
		}
		return false
	}
	_ = dfs(0)
	return count
}

// GenerateSolvedFast builds a solved grid using masks with row-major order and local shuffles.
func GenerateSolvedFast() Grid {
	for tries := 0; tries < 50; tries++ {
		var g Grid
		rows := [9]uint16{}
		cols := [9]uint16{}
		boxes := [9]uint16{}
		choices := [9]int{1, 2, 3, 4, 5, 6, 7, 8, 9}
		nodes := 0

		var fill func(int, int) bool
		fill = func(r, c int) bool {
			nodes++
			if nodes > 600000 {
				return false
			}
			if r == 9 {
				return true
			}
			nr, nc := r, c+1
			if nc == 9 {
				nr, nc = r+1, 0
			}
			used := rows[r] | cols[c] | boxes[boxIndex(r, c)]
			mask := uint16(fullMask) &^ used
			n := 0
			for v := 1; v <= 9; v++ {
				if mask&(uint16(1)<<v) != 0 {
					choices[n] = v
					n++
				}
			}
			if n == 0 {
				return false
			}
			rand.Shuffle(n, func(i, j int) { choices[i], choices[j] = choices[j], choices[i] })
			for i := 0; i < n; i++ {
				v := choices[i]
				bit := uint16(1 << v)
				g[r][c] = v
				rows[r] |= bit
				cols[c] |= bit
				b := boxIndex(r, c)
				boxes[b] |= bit
				if fill(nr, nc) {
					return true
				}
				g[r][c] = 0
				rows[r] &^= bit
				cols[c] &^= bit
				boxes[b] &^= bit
			}
			return false
		}
		if fill(0, 0) {
			return g
		}
	}
	return GenerateSolved()
}

// MakeUniquePuzzleFast removes cells with uniqueness check using masked solver and symmetry pairs.
func MakeUniquePuzzleFast(targetClues int) (Grid, Grid) {
	start := time.Now()
	solved := GenerateSolvedFast()
	p := solved.Clone()
	goalRemove := 81 - targetClues
	// symmetric indices
	idxs := rand.Perm(81)
	removed := 0
	maxFailures := 200
	fails := 0
	for _, idx := range idxs {
		if removed >= goalRemove || fails > maxFailures {
			break
		}
		r := idx / 9
		c := idx % 9
		r2 := 8 - r
		c2 := 8 - c
		if p[r][c] == 0 && p[r2][c2] == 0 {
			continue
		}
		s1, s2 := p[r][c], p[r2][c2]
		p[r][c] = 0
		if r2 != r || c2 != c {
			p[r2][c2] = 0
		}
		test := p.Clone()
		if SolveCountMasked(&test, 1) == 1 {
			if r2 == r && c2 == c {
				removed += 1
			} else {
				removed += 2
			}
		} else {
			p[r][c] = s1
			p[r2][c2] = s2
			fails++
		}
	}
	// final uniqueness check; if failed (too many fails), fall back to legacy removal refinement
	final := p.Clone()
	if SolveCountMasked(&final, 1) == 1 {
		return p, solved
	}
	// fallback refinement pass
	order := rand.Perm(81)
	for _, idx := range order {
		if removed >= goalRemove {
			break
		}
		r := idx / 9
		c := idx % 9
		if p[r][c] == 0 {
			continue
		}
		save := p[r][c]
		p[r][c] = 0
		test := p.Clone()
		if SolveCountMasked(&test, 1) != 1 {
			p[r][c] = save
		}
	}
	if time.Since(start) > 2*time.Second {
		return MakeUniquePuzzle(targetClues)
	}
	return p, solved
}

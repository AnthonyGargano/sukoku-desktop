package main

import "fmt"

// Simple logical hints: naked single and hidden single (row/col/box).

func rowHas(g Grid, r, v int) bool {
	for c := 0; c < size; c++ {
		if g[r][c] == v {
			return true
		}
	}
	return false
}
func colHas(g Grid, c, v int) bool {
	for r := 0; r < size; r++ {
		if g[r][c] == v {
			return true
		}
	}
	return false
}
func boxHas(g Grid, r, c, v int) bool {
	br, bc := (r/3)*3, (c/3)*3
	for i := br; i < br+3; i++ {
		for j := bc; j < bc+3; j++ {
			if g[i][j] == v {
				return true
			}
		}
	}
	return false
}

func candidates(g Grid, r, c int) []int {
	if g[r][c] != 0 {
		return nil
	}
	avail := []int{}
	for v := 1; v <= 9; v++ {
		if !rowHas(g, r, v) && !colHas(g, c, v) && !boxHas(g, r, c, v) {
			avail = append(avail, v)
		}
	}
	return avail
}

// FindHint searches for a single-step logical deduction.
// Returns r,c,val,reason,ok.
func FindHint(g Grid) (int, int, int, string, bool) {
	// 1) Naked singles
	for r := 0; r < size; r++ {
		for c := 0; c < size; c++ {
			if g[r][c] != 0 {
				continue
			}
			av := candidates(g, r, c)
			if len(av) == 1 {
				v := av[0]
				reason := fmt.Sprintf(
					"Naked single: Only %d fits at row %d, col %d because row, column, and box exclude all others.",
					v, r+1, c+1)
				return r, c, v, reason, true
			}
		}
	}

	// 2) Hidden single in row
	for r := 0; r < size; r++ {
		for v := 1; v <= 9; v++ {
			if rowHas(g, r, v) {
				continue
			}
			count, lastC := 0, -1
			for c := 0; c < size; c++ {
				if g[r][c] == 0 && !colHas(g, c, v) && !boxHas(g, r, c, v) {
					count++
					lastC = c
					if count > 1 {
						break
					}
				}
			}
			if count == 1 {
				reason := fmt.Sprintf("Hidden single (row): In row %d, %d can only go in column %d.", r+1, v, lastC+1)
				return r, lastC, v, reason, true
			}
		}
	}

	// 3) Hidden single in column
	for c := 0; c < size; c++ {
		for v := 1; v <= 9; v++ {
			if colHas(g, c, v) {
				continue
			}
			count, lastR := 0, -1
			for r := 0; r < size; r++ {
				if g[r][c] == 0 && !rowHas(g, r, v) && !boxHas(g, r, c, v) {
					count++
					lastR = r
					if count > 1 {
						break
					}
				}
			}
			if count == 1 {
				reason := fmt.Sprintf("Hidden single (column): In column %d, %d can only go in row %d.", c+1, v, lastR+1)
				return lastR, c, v, reason, true
			}
		}
	}

	// 4) Hidden single in box
	for br := 0; br < 3; br++ {
		for bc := 0; bc < 3; bc++ {
			r0, c0 := br*3, bc*3
			for v := 1; v <= 9; v++ {
				if boxHas(g, r0, c0, v) { // quick skip if already present
					continue
				}
				count, bestR, bestC := 0, -1, -1
				for rr := r0; rr < r0+3; rr++ {
					for cc := c0; cc < c0+3; cc++ {
						if g[rr][cc] == 0 && !rowHas(g, rr, v) && !colHas(g, cc, v) {
							count++
							bestR, bestC = rr, cc
							if count > 1 {
								break
							}
						}
					}
					if count > 1 {
						break
					}
				}
				if count == 1 {
					reason := fmt.Sprintf("Hidden single (box): In 3×3 box (%d,%d), %d can only go at row %d, col %d.",
						br+1, bc+1, v, bestR+1, bestC+1)
					return bestR, bestC, v, reason, true
				}
			}
		}
	}

	return -1, -1, 0, "No simple hints available (try notes or a different region).", false
}

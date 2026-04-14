package main

import "testing"

func TestGenerateSolved_IsValid(t *testing.T) {
	g := GenerateSolved()
	if SolveCount(&g, 1) != 1 {
		t.Fatalf("legacy solved grid should be valid and unique")
	}

	g2 := GenerateSolvedFast()
	// A fully-solved grid has exactly 0 empties, so SolveCountMasked should return 1
	// (it won't need to search at all). Use limit=2 for real uniqueness check.
	if SolveCountMasked(&g2, 2) != 1 {
		t.Fatalf("fast solved grid should be valid and unique")
	}
}

func TestMakeUniquePuzzle_Uniqueness(t *testing.T) {
	for _, clues := range []int{40, 32, 25} {
		p, sol := MakeUniquePuzzleFast(clues)
		g := p.Clone()
		// limit=2: if exactly 1 solution exists the solver won't find a 2nd, so count==1.
		if SolveCountMasked(&g, 2) != 1 {
			t.Fatalf("fast puzzle not unique at clues=%d", clues)
		}
		// All filled cells in the puzzle must match the returned solution.
		for r := 0; r < 9; r++ {
			for c := 0; c < 9; c++ {
				if p[r][c] != 0 && p[r][c] != sol[r][c] {
					t.Fatalf("solution mismatch at (%d,%d)", r, c)
				}
			}
		}
		// The stored solution must itself be the unique solution to the puzzle.
		solClone := sol
		for r := 0; r < 9; r++ {
			for c := 0; c < 9; c++ {
				if p[r][c] != 0 {
					solClone[r][c] = p[r][c]
				}
			}
		}
		if SolveCountMasked(&solClone, 2) != 1 {
			t.Fatalf("stored solution is not itself uniquely solved at clues=%d", clues)
		}
	}
}

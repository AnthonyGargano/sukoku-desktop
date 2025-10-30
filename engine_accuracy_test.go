package main

import "testing"

func TestGenerateSolved_IsValid(t *testing.T) {
	g := GenerateSolved()
	if SolveCount(&g, 1) != 1 {
		t.Fatalf("legacy solved grid should be valid and unique")
	}

	g2 := GenerateSolvedFast()
	if SolveCountMasked(&g2, 1) != 1 {
		t.Fatalf("fast solved grid should be valid and unique")
	}
}

func TestMakeUniquePuzzle_Uniqueness(t *testing.T) {
	for _, clues := range []int{40, 32, 25} {
		p, sol := MakeUniquePuzzleFast(clues)
		g := p.Clone()
		if SolveCountMasked(&g, 1) != 1 {
			t.Fatalf("fast puzzle not unique at clues=%d", clues)
		}
		// solution must satisfy puzzle
		for r := 0; r < 9; r++ {
			for c := 0; c < 9; c++ {
				if p[r][c] != 0 && p[r][c] != sol[r][c] {
					t.Fatalf("solution mismatch at (%d,%d)", r, c)
				}
			}
		}
	}
}

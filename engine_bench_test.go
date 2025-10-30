package main

import (
	"testing"
)

func BenchmarkGenerateSolved_Legacy(b *testing.B) {
	for i := 0; i < b.N; i++ {
		_ = GenerateSolved()
	}
}

func BenchmarkGenerateSolved_Fast(b *testing.B) {
	for i := 0; i < b.N; i++ {
		_ = GenerateSolvedFast()
	}
}

func BenchmarkSolveCount_Unique_Legacy(b *testing.B) {
	p, _ := MakeUniquePuzzle(32)
	for i := 0; i < b.N; i++ {
		g := p.Clone()
		_ = SolveCount(&g, 1)
	}
}

func BenchmarkSolveCount_Unique_Masked(b *testing.B) {
	p, _ := MakeUniquePuzzleFast(32)
	for i := 0; i < b.N; i++ {
		g := p.Clone()
		_ = SolveCountMasked(&g, 1)
	}
}



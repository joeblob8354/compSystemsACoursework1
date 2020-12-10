package main

import (
        "testing"
        "uk.ac.bris.cs/gameoflife/gol"
        )
var p gol.Params

func benchmarkParallel(p gol.Params, b *testing.B) {
    p.ImageHeight = 16
    p.ImageWidth = 16
    p.Turns = 100
    for n := 0; n < b.N; n++ {
        events := make(chan gol.Event)
        gol.Run(p, events, nil)
        var turn int
        for turn != p.Turns {
            for event := range events {
                switch e := event.(type) {
        	    case gol.FinalTurnComplete:
        	        turn = e.CompletedTurns
        	    }
            }
        }
    }
}

func BenchmarkParallel1(b *testing.B) {
    p.Threads = 1
    benchmarkParallel(p, b)
}
/*
func BenchmarkParallel2(b *testing.B) {
    p.Threads = 2
    benchmarkParallel(p, b)
}

func BenchmarkParallel3(b *testing.B) {
    p.Threads = 3
    benchmarkParallel(p, b)
}

func BenchmarkParallel4(b *testing.B) {
    p.Threads = 4
    benchmarkParallel(p, b)
}

func BenchmarkParallel5(b *testing.B) {
    p.Threads = 5
    benchmarkParallel(p, b)
}

func BenchmarkParallel6(b *testing.B) {
    p.Threads = 6
    benchmarkParallel(p, b)
}

func BenchmarkParallel7(b *testing.B) {
    p.Threads = 7
    benchmarkParallel(p, b)
}

func BenchmarkParallel8(b *testing.B) {
    p.Threads = 8
    benchmarkParallel(p, b)
}

func BenchmarkParallel9(b *testing.B) {
    p.Threads = 9
    benchmarkParallel(p, b)
}

func BenchmarkParallel10(b *testing.B) {
    p.Threads = 10
    benchmarkParallel(p, b)
}

func BenchmarkParallel11(b *testing.B) {
    p.Threads = 11
    benchmarkParallel(p, b)
}

func BenchmarkParallel12(b *testing.B) {
    p.Threads = 12
    benchmarkParallel(p, b)
}

func BenchmarkParallel13(b *testing.B) {
    p.Threads = 13
    benchmarkParallel(p, b)
}

func BenchmarkParallel14(b *testing.B) {
    p.Threads = 14
    benchmarkParallel(p, b)
}

func BenchmarkParallel28(b *testing.B) {
    p.Threads = 15
    benchmarkParallel(p, b)
}

func BenchmarkParallel30(b *testing.B) {
    p.Threads = 16
    benchmarkParallel(p, b)
}*/

package main

import (
        "testing"
        "uk.ac.bris.cs/gameoflife/gol"
        )
var p gol.Params

func benchmarkDistributed(p gol.Params, b *testing.B) {
    p.ImageHeight = 128
    p.ImageWidth = 128
    p.Turns = 100
    for n := 0; n < b.N; n++ {
        events := make(chan gol.Event)
        gol.Run(p, events, nil)
        var turn int
        for turn != p.Turns {
            for event := range events {
                switch e := event.(type) {
        	    case gol.ImageOutputComplete:
        	        turn = e.CompletedTurns
        	    }
            }
        }
    }
}

func BenchmarkDistributed1(b *testing.B) {
    p.Threads = 4
    benchmarkDistributed(p, b)
}

func BenchmarkDistributed2(b *testing.B) {
    p.Threads = 7
    benchmarkDistributed(p, b)
}

func BenchmarkDistributed3(b *testing.B) {
    p.Threads = 8
    benchmarkDistributed(p, b)
}/*

func BenchmarkDistributed4(b *testing.B) {
    p.Threads = 4
    benchmarkDistributed(p, b)
}

func BenchmarkDistributed5(b *testing.B) {
    p.Threads = 5
    benchmarkDistributed(p, b)
}

func BenchmarkDistributed6(b *testing.B) {
    p.Threads = 6
    benchmarkDistributed(p, b)
}

func BenchmarkDistributed7(b *testing.B) {
    p.Threads = 7
    benchmarkDistributed(p, b)
}

func BenchmarkDistributed8(b *testing.B) {
    p.Threads = 8
    benchmarkDistributed(p, b)
}

func BenchmarkDistributed9(b *testing.B) {
    p.Threads = 9
    benchmarkDistributed(p, b)
}

func BenchmarkDistributed10(b *testing.B) {
    p.Threads = 10
    benchmarkDistributed(p, b)
}

func BenchmarkDistributed11(b *testing.B) {
    p.Threads = 11
    benchmarkDistributed(p, b)
}

func BenchmarkDistributed12(b *testing.B) {
    p.Threads = 12
    benchmarkDistributed(p, b)
}

func BenchmarkDistributed13(b *testing.B) {
    p.Threads = 13
    benchmarkDistributed(p, b)
}

func BenchmarkDistributed14(b *testing.B) {
    p.Threads = 14
    benchmarkDistributed(p, b)
}

func BenchmarkDistributed28(b *testing.B) {
    p.Threads = 15
    benchmarkDistributed(p, b)
}

func BenchmarkDistributed30(b *testing.B) {
    p.Threads = 16
    benchmarkDistributed(p, b)
}*/

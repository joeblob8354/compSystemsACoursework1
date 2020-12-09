package main

import (
        "testing"
        "uk.ac.bris.cs/gameoflife/gol"
        "strconv"
        )
var p gol.Params

func benchmarkParallel(benchName string, p gol.Params, b *testing.B) {
    for n := 0; n < b.N; n++ {
        gol.Run(p, nil, nil)
    }
}

func BenchmarkParallel(b *testing.B) {
    p.ImageHeight = 16
    p.ImageWidth = 16
    p.Turns = 1000
    for p.Threads = 1; p.Threads <= 16; p.Threads++ {
        benchName := strconv.Itoa(p.ImageHeight) + "x" + strconv.Itoa(p.ImageWidth) + "x" + strconv.Itoa(p.Turns) + "-" + strconv.Itoa(p.Threads)
        benchmarkParallel(benchName, p, b)
    }
}
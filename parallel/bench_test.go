package main

import (
        "testing"
        "uk.ac.bris.cs/gameoflife/gol"
        "strconv"
        )
var p gol.Params

func benchmarkParallel(p gol.Params, b *testing.B) {
    for n := 0; n <= 10; n++ {
        gol.Run(p, nil, nil)
    }
}

func BenchmarkParallel(b *testing.B) {
    p.ImageHeight = 512
    p.ImageWidth = 512
    p.Turns = 1000
    p.Threads = 1
    benchmarkParallel(p, b)

}
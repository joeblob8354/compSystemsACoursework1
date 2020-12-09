package main

import (
        "testing"
        "uk.ac.bris.cs/gameoflife/gol"
        )
var p gol.Params

func benchmarkParallel(p gol.Params, b *testing.B) {
    for n := 0; n < 10; n++ {
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

func BenchmarkParallel(b *testing.B) {
    p.ImageHeight = 512
    p.ImageWidth = 512
    p.Turns = 1000
    p.Threads = 1
    benchmarkParallel(p, b)

}
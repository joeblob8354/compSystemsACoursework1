package main

import (
	"flag"
	//"fmt"
	"runtime"
	"uk.ac.bris.cs/gameoflife/gol"
	//"uk.ac.bris.cs/gameoflife/sdl"
	"net/rpc"
	"net"
)

type Data struct {
    TheParams gol.Params
    World     [][]byte
    Turn      int
}

type Engine struct {}

var globalTurn = 0

var globalWorld [][]byte

// Run starts the processing of Game of Life. It should initialise channels and goroutines.
func (e *Engine) Run(data Data, reply *[][]byte) error {

    if data.World == nil {
        data.World = globalWorld
    }
    globalWorld = gol.CalculateNextState(data.TheParams, data.World)
    globalTurn = data.Turn
    *reply = globalWorld
    if globalTurn == data.TheParams.Turns - 1 {
        globalTurn = 0
        globalWorld = nil
    }
    return nil
}

func (e *Engine) CheckTurnNumber(x int, turnReply *int) error {

    *turnReply = globalTurn
    return nil
}

func (e *Engine) GetWorld(x int, worldReply *[][]byte) error {

    *worldReply = globalWorld
    return nil
}

// main is the function called when starting Game of Life with 'go run .'
func main() {
	runtime.LockOSThread()

	//keyPresses := make(chan rune, 10)

    //Listen for incoming client connections
    var pAddr = flag.String("port","8030","Port to listen on")
    flag.Parse()
    rpc.Register(&Engine{})
    listener, _ := net.Listen("tcp", ":"+*pAddr)
    defer listener.Close()
    rpc.Accept(listener)
}
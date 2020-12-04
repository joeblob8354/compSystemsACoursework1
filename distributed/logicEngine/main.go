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
}

type Engine struct {}

// Run starts the processing of Game of Life. It should initialise channels and goroutines.
func (e *Engine) Run(data Data, reply *[][]byte) error {


	newWorld := gol.Distributor(data.TheParams, data.World)
    *reply = newWorld

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
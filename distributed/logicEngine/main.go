package main

import (
	"flag"
	//"fmt"
	"runtime"
	"uk.ac.bris.cs/gameoflife/gol"
	//"uk.ac.bris.cs/gameoflife/sdl"
	"net/rpc"
	"net"
	"log"
)

type Data struct {
    TheParams gol.Params
    World     [][]byte
    Turn      int
}

type WorkerData struct {
    TheParams   gol.Params
    World       [][]byte
    StartHeight int
    EndHeight   int
}

type Engine struct {}

var globalTurn = 0

var globalWorld [][]byte

//add addresses of aws nodes here
var nodeAddresses = [2]string{"54.208.137.161:8030", "3.93.7.41:8030"}

// Run starts the processing of Game of Life. It should initialise channels and goroutines.
func (e *Engine) RunMaster(data Data, reply *[][]byte) error {

    if data.World == nil {
        data.World = globalWorld
    }

    numberOfNodes := data.TheParams.Threads

    if numberOfNodes == 1 {
        globalWorld = gol.CalculateNextState(data.TheParams, 0, data.TheParams.ImageHeight, data.World)
    } else {
        heightOfSection := data.TheParams.ImageHeight/numberOfNodes

        var workerData WorkerData
        workerData.TheParams = data.TheParams
        workerData.World = data.World
        workerData.StartHeight = 0
        workerData.EndHeight = heightOfSection

        nodes := []*rpc.Client{}
        var err error
        replies := []*[][]byte{}
        for n := 0; n < numberOfNodes; n++ {
            nodes[n], err = rpc.Dial("tcp", nodeAddresses[n])
            if err != nil {
                log.Fatal("Failed to connect to node ", n, " ", err)
            }
            nodes[n].Call("Engine.RunWorker", workerData, &replies[n])
            workerData.StartHeight = workerData.StartHeight + heightOfSection
            workerData.EndHeight = workerData.EndHeight + heightOfSection
        }

        for node := 0; node < numberOfNodes; node++ {
            part := *replies[node]
            globalWorld = append(globalWorld, part...)
        }
    }

    globalTurn = data.Turn
    *reply = globalWorld
    if globalTurn == data.TheParams.Turns - 1 {
        globalTurn = 0
        globalWorld = nil
    }
    return nil
}

func (e *Engine)RunWorker (data WorkerData, reply *[][]byte) error {

    *reply = gol.CalculateNextState(data.TheParams, data.StartHeight, data.EndHeight, data.World)
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
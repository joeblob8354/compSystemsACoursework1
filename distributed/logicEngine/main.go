package main

import (
	"flag"
	"fmt"
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

var globalParams gol.Params

//add addresses of aws nodes here
var nodeAddresses = [2]string{"54.208.137.161:8030", "3.93.7.41:8030"}

// Run starts the processing of Game of Life. It should initialise channels and goroutines.
func (e *Engine) RunMaster(data Data, reply *[][]byte) error {

    globalParams = data.TheParams

    if data.World == nil {
        data.World = globalWorld
    }

    numberOfNodes := data.TheParams.Threads
    //fmt.Println(numberOfNodes)

    if numberOfNodes == 1 {
        globalWorld = gol.CalculateNextState(data.TheParams, 0, data.TheParams.ImageHeight, data.World)
    } else {
        heightOfSection := data.TheParams.ImageHeight/numberOfNodes

        var workerData WorkerData
        workerData.TheParams = data.TheParams
        workerData.World = data.World
        workerData.StartHeight = 0
        workerData.EndHeight = heightOfSection

        /*nodes := []*rpc.Client{}
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
        }*/

        node0, err := rpc.Dial("tcp", nodeAddresses[0])
        if err != nil {
            log.Fatal("Failed to connect to node 0", err)
        }

        node1, err1 := rpc.Dial("tcp", nodeAddresses[1])
        if err1 != nil {
            log.Fatal("Failed to connect to node 1", err1)
        }

        workerData.StartHeight = 0
        workerData.EndHeight = heightOfSection
        var workerReply0 [][]byte
        node0.Call("Engine.RunWorker", workerData, &workerReply0)

        workerData.StartHeight = heightOfSection
        workerData.EndHeight = heightOfSection + heightOfSection
         var workerReply1 [][]byte
        node1.Call("Engine.RunWorker", workerData, &workerReply1)

        /*for node := 0; node < numberOfNodes; node++ {
            part := *replies[node]
            globalWorld = append(globalWorld, part...)
        }*/

        globalWorld = nil

        workerReplies := [][][]byte{}
        workerReplies[0] = workerReply0
        workerReplies[1] = workerReply1

        for node := 0; node < numberOfNodes; node++ {
    	    part := workerReplies[node]
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

    fmt.Println("Params reset")
    *turnReply = globalTurn
    return nil
}

func (e *Engine) GetWorld(x int, worldReply *[][]byte) error {

    *worldReply = globalWorld
    return nil
}

//checks if the params of the connected controller match those of the previous controller
func (e *Engine) CheckParams(p gol.Params, reply *bool) error {

    if p == globalParams {
        *reply = true
    } else {
        *reply = false
    }
    return nil
}

func (e *Engine) ResetGlobals(x int, reply *bool) error {

    globalTurn, globalWorld = 0, nil
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
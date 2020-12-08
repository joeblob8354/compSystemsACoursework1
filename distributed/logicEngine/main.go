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
    if numberOfNodes > len(nodeAddresses) {
        fmt.Println("Not enough nodes available! Using", len(nodeAddresses), "nodes instead.")
        numberOfNodes = len(nodeAddresses)
    }

    numberOfNodes = 1

    if numberOfNodes == 1 {
        globalWorld = gol.CalculateNextState(data.TheParams, 0, data.TheParams.ImageHeight, data.World)
    } else {
        heightOfSection := data.TheParams.ImageHeight/numberOfNodes

        var workerData WorkerData
        workerData.TheParams = data.TheParams
        workerData.World = data.World

        workerReplies := [][][]byte{}
        listOfNodes := []*rpc.Client{}

        for numberOfWorkers := 0; numberOfWorkers < numberOfNodes; numberOfWorkers++ {
            var reply [][]byte
            workerReplies = append(workerReplies, reply)
            var client *rpc.Client
            listOfNodes = append(listOfNodes, client)
        }

        for node := 0; node < numberOfNodes; node++ {
            var err error
            listOfNodes[node], err = rpc.Dial("tcp", nodeAddresses[node])
            if err != nil {
                log.Fatal("Failed to connect to node ", node, " ", err)
            }
            workerData.StartHeight = 0 + node*heightOfSection
            workerData.EndHeight = heightOfSection + node*heightOfSection
            listOfNodes[node].Call("Engine.RunWorker", workerData, &workerReplies[node])
        }

        globalWorld = nil

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
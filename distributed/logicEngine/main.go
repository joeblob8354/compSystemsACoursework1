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

var globalParams gol.Params

//add addresses of aws nodes here
var nodeAddresses = [8]string{"54.208.137.161:8030", "3.93.7.41:8030",
                              "54.80.215.196:8030", "54.210.236.107:8030",
                              "34.229.232.22:8030", "54.81.217.163:8030",
                              "18.207.161.206:8030", "35.168.13.14:8030"}

// Run starts the processing of Game of Life. It should initialise channels and goroutines.
func (e *Engine) RunMaster(data Data, reply *[][]byte) error {

    //sets globalParams to the params provided by the client
    globalParams = data.TheParams

    //if the world data provided by the client is nil then set world data to world stored by globalWorld
    if data.World == nil {
        data.World = globalWorld
    }

    //sets numberOfNodes to number requested by controller
    numberOfNodes := data.TheParams.Threads
    //if number of available nodes is less than requested number, set numberOfNodes to max that are available
    if numberOfNodes > len(nodeAddresses) {
        numberOfNodes = len(nodeAddresses)
    }

    //if numberOfNodes requested is 1, bypass worker nodes and compute the world state on master node
    if numberOfNodes == 1 {
        globalWorld = gol.CalculateNextState(data.TheParams, 0, data.TheParams.ImageHeight, data.World)

    // if numberOfNodes requested is > 1 then split the world between however many worker nodes
    } else {
        //split into equal sized sections based on numberOfNodes requested
        heightOfSection := data.TheParams.ImageHeight/numberOfNodes

        //create a new variable to store provide/ store data to and from workers
        var workerData WorkerData
        //set worker params
        workerData.TheParams = data.TheParams
        //give workers the world
        workerData.World = data.World

        //create a slice to store the updated worlds sent back by the workers
        workerReplies := [][][]byte{}
        //create a slice to store worker connections
        listOfNodes := []*rpc.Client{}

        //for each worker, add an entry in the replies and connections slices
        for numberOfWorkers := 0; numberOfWorkers < numberOfNodes; numberOfWorkers++ {
            var reply [][]byte
            workerReplies = append(workerReplies, reply)
            var client *rpc.Client
            listOfNodes = append(listOfNodes, client)
        }

        //creates a slice of channels that will be used to receive incoming completed sections from workers.
    	chanSlice := make([]chan [][]byte, numberOfNodes)
    	for i := range chanSlice {
    		chanSlice[i] = make(chan [][]byte)
    	}

        //for each worker, connect to the aws node and store the connection pointer in the listOfNodes slice or throw an error.
        for node := 0; node < numberOfNodes - 1; node++ {
            var err error
            listOfNodes[node], err = rpc.Dial("tcp", nodeAddresses[node])
            if err != nil {
                log.Fatal("Failed to connect to node ", node, " ", err)
            }
            //give each worker a section of the world to work on and call the RunWorker method for each worker using .Call for each node in the listOfNodes.
            workerData.StartHeight = 0 + node*heightOfSection
            workerData.EndHeight = heightOfSection + node*heightOfSection
            //store the replied world data in the worker replies slice.
            go call(node, workerData, listOfNodes, workerReplies, chanSlice[node])
        }
        //this last worker .Call is to prevent odd numbers of workers causing issues
        var err error
        listOfNodes[numberOfNodes - 1], err = rpc.Dial("tcp", nodeAddresses[numberOfNodes - 1])
        if err != nil {
            log.Fatal("Failed to connect to node ", numberOfNodes - 1, " ", err)
        }
        workerData.StartHeight = (numberOfNodes - 1)*heightOfSection
        workerData.EndHeight = data.TheParams.ImageHeight
        go call((numberOfNodes - 1), workerData, listOfNodes, workerReplies, chanSlice[(numberOfNodes - 1)])

        //reset globalWorld state
        globalWorld = nil

        //stick the worker parts together into one final world state
        for node := 0; node < numberOfNodes; node++ {
    	    part := <-chanSlice[node]
    		globalWorld = append(globalWorld, part...)
    	}
    }

    //update the global turn
    globalTurn = data.Turn
    //set the reply pointer to the finished updated world state
    *reply = globalWorld

    //reset the global variables if its the last turn.=
    if globalTurn == data.TheParams.Turns - 1 {
        globalTurn = 0
        globalWorld = nil
    }
    return nil
}

func call(node int, workerData WorkerData, listOfNodes []*rpc.Client, workerReplies [][][]byte, out chan<- [][]byte) {

    listOfNodes[node].Call("Engine.RunWorker", workerData, &workerReplies[node])
    out <- workerReplies[node]
}

//calculates the next state of a world given a world state and y-coordinates to work on
func (e *Engine)RunWorker (data WorkerData, reply *[][]byte) error {

    *reply = gol.CalculateNextState(data.TheParams, data.StartHeight, data.EndHeight, data.World)
    return nil
}

//checks the turn the server was working on before it quit its last operation
func (e *Engine) CheckTurnNumber(x int, turnReply *int) error {

    *turnReply = globalTurn
    return nil
}

//get the world from the global world variable and sends in back to the client as a reply
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

//resets the global variables on the master node
func (e *Engine) ResetGlobals(x int, reply *bool) error {

    globalTurn, globalWorld = 0, nil
    return nil
}

//gets how many aws node addresses are available for use and sends this info to the client
func (e* Engine) GetAvailableNodes(x int, reply *int) error {

    *reply = len(nodeAddresses)
    return nil
}

// main is the function called when starting Game of Life with 'go run .'
func main() {
	runtime.LockOSThread()

    //Listen for incoming client connections
    var pAddr = flag.String("port","8030","Port to listen on")
    flag.Parse()
    rpc.Register(&Engine{})
    listener, _ := net.Listen("tcp", ":"+*pAddr)
    defer listener.Close()
    rpc.Accept(listener)
}
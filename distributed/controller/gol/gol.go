package gol

import (
    "uk.ac.bris.cs/gameoflife/util"
    "net/rpc"
    "log"
    "strconv"
    "fmt"
    "time"
    "os"
)

// Params provides the details of how to run the Game of Life and which image to load.
type Params struct {
	Turns         int
	Threads       int
	ImageWidth    int
	ImageHeight   int
	ServerAddress string
}

type Data struct {
    TheParams Params
    World     [][]byte
    Turn      int
}

type distributorChannels struct {
	events     chan<- Event
	ioCommand  chan<- ioCommand
	ioIdle     <-chan bool
	ioFilename chan<- string
	ioInput    <-chan uint8
	ioOutput   chan<- uint8
}

func engine(p Params, d distributorChannels, k <-chan rune) {

    //Creates a 2D slice to store the world.
    newWorld := make([][]byte, p.ImageHeight)
    for i := 0; i < p.ImageHeight; i++ {
        newWorld[i] = make([]byte, p.ImageWidth)
    }

    //sends command to io so it executes the readPgmImage() function
    d.ioCommand <- 1
    //sends created filename to readPgmImage()
    d.ioFilename <- (strconv.Itoa(p.ImageHeight) + "x" + strconv.Itoa(p.ImageWidth))

    //reads in bytes 1 at a time from the ioInput channel and populates the world
    for currRow := 0; currRow < p.ImageHeight; currRow++ {
        for currColumn := 0; currColumn < p.ImageWidth; currColumn++ {
            newWorld[currRow][currColumn] = <- d.ioInput
        }
    }

    //connect to server or return an error
    serverAddress := p.ServerAddress + ":8030"
    client, err := rpc.Dial("tcp", serverAddress)

    if err != nil {
        log.Fatal("Failed to connect to ", err)
    }

    //check to see if the params stored by the logicEngine match those given by the controller
    var paramsReply bool
    client.Call("Engine.CheckParams", p, &paramsReply)

    //create var of type Data to store necessary data to send to logic engine.
    var data Data

    var turn int
    var x int
    var turnReply int

    //get the turn stored by the logicEngine
    client.Call("Engine.CheckTurnNumber", x, &turnReply)
    turn = turnReply

    //if the params from the logic engine match those given by the controller...
    if paramsReply == true {
        //...and the logic engine has an unfinished board, then continue to process the unfinished board.
        if turn != 0 {
           fmt.Println("Unfinished board found with matching parameters, continuing processing unfinished board...")
        }

        //... or if the params match but no unfinished board is stored then set the world data to world state provided by the client
        if turn == 0 {
            data.World = newWorld
        } else {
            turn++
        }

        worldReply := newWorld
        //if the world state is still nil at this point (not provided by client) get the unfinished board world state from the logic engine.
        if data.World == nil {
            client.Call("Engine.GetWorld", x, &worldReply)
            //...and update the world data.
            data.World = worldReply
        }
    } else { //if the params given by the client don't match those stored on the engine...
        //... and there is an unfinished board stored on the engine. Then warn the controller of this...
        if turn != 0{
            fmt.Println("Warning: Unfinished board found with differing parameters, starting processing of new board with new parameters...")
        }
        var boolReply bool
        var z int
        //... and reset global variables of on the engine to prepare it for processing a new board with differing params.
        client.Call("Engine.ResetGlobals", z, &boolReply)
        turn = 0
        data.World = newWorld
    }

    //set data params and turn
    data.TheParams = p
    data.Turn = turn

    var y int
    var availableNodes int
    //check how many nodes are available for use on the engine
    client.Call("Engine.GetAvailableNodes", y, &availableNodes)

    //adjusts the parameters if less than number of nodes requested are available
    if data.TheParams.Threads > availableNodes {
        fmt.Println("Not enough nodes available! Using", availableNodes, "nodes instead.")
        data.TheParams.Threads = availableNodes
    }

    ///create a reply variable to receive the updated world from the logic engine
    var reply [][]byte

    //create a new ticker and start it.
    tk := time.NewTicker(time.Second*2)
    cellCount := len(calculateAliveCells(data.TheParams, data.World))
    go ticker(tk, &cellCount, &turn, d, p)

    //change state to executing
    d.events <- StateChange{CompletedTurns: turn, NewState: Executing}

    //For each turn, call the Run method on the server and send it the world and other data for processing.
    for turn = turn; turn < p.Turns; turn++ {
        data.Turn = turn
        cellCount = len(calculateAliveCells(data.TheParams, data.World))
        client.Call("Engine.RunMaster", data, &reply)
        data.World = reply
        var key rune

        //listening for incoming key-presses without blocking
        select {
            case key = <- k:
                //if s is pressed output a pgm img of the current world state and the corresponding turn.
                if key == 's' {
                    outputPgmFile(d, p, data.World, turn)

                //if p is pressed, change state to paused, stop the ticker, and wait for p to be pressed again before continuing.
                } else if key == 'p' {
                    d.events <- StateChange{CompletedTurns: turn, NewState: Paused}
                    tk.Stop()
                    key = <- k
                    for key != 'p' {
                        key = <- k
                    }
                    //change state to executing and restart the ticker
                    d.events <- StateChange{CompletedTurns: turn, NewState: Executing}
                    tk = time.NewTicker(time.Second*2)
                    go ticker(tk, &cellCount, &turn, d, p)
                // if q is pressed, change state to quitting and exit.
                } else if key == 'q' {
                    d.events <- StateChange{CompletedTurns: turn, NewState: Quitting}
                    os.Exit(0)
                //if k is pressed close all elements of system, starting with worker nodes, then master node
                } else if key == 'k' {
                    tk.Stop()
                    var x, reply int
                    client.Call("Engine.QuitAll", x, &reply)
                    outputPgmFile(d, p, data.World, turn)
                    // Make sure that the Io has finished any output before exiting.
                    d.ioCommand <- ioCheckIdle
                    <-d.ioIdle
                    d.events <- StateChange{CompletedTurns: turn, NewState: Quitting}
                    os.Exit(0)
                }
            //otherwise, do nothing and continue to next turn
            default:
        }
    }

    //stops ticker
    tk.Stop()

    //send array of alive cells for testing
    aliveCells := calculateAliveCells(p, data.World)
    d.events <- FinalTurnComplete{CompletedTurns: p.Turns, Alive: aliveCells}

    //outputs pgm file
    outputPgmFile(d, p, data.World, turn)

    // Make sure that the Io has finished any output before exiting.
 	d.ioCommand <- ioCheckIdle
 	<-d.ioIdle

    //close events channel
    close(d.events)
}

//returns a array of alive cells in the current state
func calculateAliveCells(p Params, world [][]byte) []util.Cell {
	var aliveCells []util.Cell
	for currRow := 0; currRow < p.ImageHeight; currRow++ {
	    for currColumn := 0; currColumn < p.ImageWidth; currColumn++ {
	        if world[currRow][currColumn] == 255 {
	            aliveCells = append(aliveCells, util.Cell{X: currColumn, Y: currRow})
	        }
	    }
	}
	return aliveCells
}

//ticker function that loops every 2 seconds and sends AliveCellsCount events
func ticker(tk *time.Ticker, cellCount *int, turn *int, d distributorChannels, p Params) {
    for range tk.C{
        theTurn := *turn
        theCount := *cellCount
        if theTurn == 0 {
            theCount = 0
        }
        d.events <- AliveCellsCount{CompletedTurns: theTurn, CellsCount: theCount}
    }
}

//Outputs a program file of the world state.
func outputPgmFile (d distributorChannels, p Params, world [][]byte, turn int) {

    //send command to io to let make it execute the writePgmImage() function.
    d.ioCommand <- 0
    //send the filename to the writePgmImage() function.
    d.ioFilename <- (strconv.Itoa(p.ImageHeight) + "x" + strconv.Itoa(p.ImageWidth) + "x" + strconv.Itoa(turn))

    //Scan across the updated world and send bytes 1 at a time to the writePgmImage() function via the ioOutput channel.
    for currRow := 0; currRow < p.ImageHeight; currRow++ {
        for currColumn := 0; currColumn < p.ImageWidth; currColumn++ {
            d.ioOutput <- world[currRow][currColumn]
        }
    }
}

//create necessary channels and start go routines.
func Run(p Params, events chan<- Event, keyPresses <-chan rune) {

    //creates all necessary channels
    ioCommand := make(chan ioCommand)
    ioIdle := make(chan bool)
    ioFilename := make(chan string)
    ioOutput := make(chan uint8)
    ioInput := make(chan uint8)

    distributorChannels := distributorChannels{
    	events,
    	ioCommand,
    	ioIdle,
    	ioFilename,
    	ioInput,
    	ioOutput,
    }

    go engine(p, distributorChannels, keyPresses)

    ioChannels := ioChannels{
        command:  ioCommand,
    	idle:     ioIdle,
    	filename: ioFilename,
    	output:   ioOutput,
    	input:    ioInput,
    }
    go startIo(p, ioChannels)
}

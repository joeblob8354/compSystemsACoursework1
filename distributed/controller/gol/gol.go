package gol

import (
    "uk.ac.bris.cs/gameoflife/util"
    "net/rpc"
    "log"
    "strconv"
    //"fmt"
    "time"
    "os"
)

// Params provides the details of how to run the Game of Life and which image to load.
type Params struct {
	Turns       int
	Threads     int
	ImageWidth  int
	ImageHeight int
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
    serverAddress := "localhost:8030"
    client, err := rpc.Dial("tcp", serverAddress)

    if err != nil {
        log.Fatal("connection error", err)
    }

    var turn int

    var x int
    var turnReply int
    client.Call("Engine.CheckTurnNumber", x, &turnReply)
    turn = turnReply

    //create var of type Data to store necessary data to send to logic engine.
    var data Data
    if turn == 0 {
        data.World = newWorld
    }
    data.TheParams = p

    worldReply := newWorld
    if data.World == nil {
        client.Call("Engine.GetWorld", x, &worldReply)
        data.World = worldReply
    }

    ///create a reply variable to receive the updated world from the logic engine
    var reply [][]byte

    //create a new ticker and start it.
    tk := time.NewTicker(time.Second*1)
    go ticker(tk, &data.World, &turn, d, p)

    d.events <- StateChange{CompletedTurns: turn, NewState: Paused}

    //call the Run method on the server and send it the world
    for turn = turnReply; turn < p.Turns; turn++ {
        data.Turn = turn
        client.Call("Engine.Run", data, &reply)
        data.World = reply
        var key rune
        select {
            case key = <- k:
                if key == 's' {
                    outputPgmFile(d, p, data.World, turn)
                } else if key == 'p' {
                    d.events <- StateChange{CompletedTurns: turn, NewState: Paused}
                    tk.Stop()
                    key = <- k
                    for key != 'p' {
                        key = <- k
                    }
                    d.events <- StateChange{CompletedTurns: turn, NewState: Executing}
                    tk = time.NewTicker(time.Second*1)
                    go ticker(tk, &data.World, &turn, d, p)
                } else if key == 'q' {
                    d.events <- StateChange{CompletedTurns: turn, NewState: Quitting}
                    os.Exit(0)
                }
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

//ticker function that loops every 2 seconds and sends AliveCellsCount events
func ticker(tk *time.Ticker, world *[][]byte, turn *int, d distributorChannels, p Params) {
    for range tk.C{
        d.events <- AliveCellsCount{CompletedTurns: *turn, CellsCount: checkNumberOfAliveCells(p, *world)}
    }
}

//returns number of alive cells in a world state
func checkNumberOfAliveCells(p Params, world [][]byte) int {
    numberOfAliveCells := 0
    for currRow := 0; currRow < p.ImageHeight; currRow++ {
	    for currColumn := 0; currColumn < p.ImageWidth; currColumn++ {
	        if world[currRow][currColumn] == 255 {
	            numberOfAliveCells++
	        }
	    }
	}
	return numberOfAliveCells
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

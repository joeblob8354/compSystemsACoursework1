package gol

import (
    "uk.ac.bris.cs/gameoflife/util"
    "net/rpc"
    "log"
    "strconv"
    //"fmt"
    "time"
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

func engine(p Params, d distributorChannels) {

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
    serverAddress := "18.232.162.211:8030"
    client, err := rpc.Dial("tcp", serverAddress)

    if err != nil {
        log.Fatal("connection error", err)
    }

    var data Data
    data.World = newWorld
    data.TheParams = p
    turn := 0

    var reply [][]byte
    tk := time.NewTicker(time.Second*1)
    go ticker(tk, &data.World, &turn, d, p)
    //call the Run method on the server and send it the world
    for turn = 0; turn < p.Turns; turn++ {
        client.Call("Engine.Run", data, &reply)
        data.World = reply
    }

    tk.Stop()

    //send array of alive cells for testing
    aliveCells := calculateAliveCells(p, data.World)
    d.events <- FinalTurnComplete{CompletedTurns: p.Turns, Alive: aliveCells}

    //send command to io to let make it execute the writePgmImage() function.
    d.ioCommand <- 0
    //send the filename to the writePgmImage() function.
    d.ioFilename <- (strconv.Itoa(p.ImageHeight) + "x" + strconv.Itoa(p.ImageWidth) + "x" + strconv.Itoa(turn))

    //Scan across the updated world and send bytes 1 at a time to the writePgmImage() function via the ioOutput channel.
    for currRow := 0; currRow < p.ImageHeight; currRow++ {
        for currColumn := 0; currColumn < p.ImageWidth; currColumn++ {
            d.ioOutput <- data.World[currRow][currColumn]
        }
    }

    // Make sure that the Io has finished any output before exiting.
 	d.ioCommand <- ioCheckIdle
 	<-d.ioIdle

    close(d.events)
}

func ticker(tk *time.Ticker, world *[][]byte, turn *int, d distributorChannels, p Params) {
    for range tk.C{
        d.events <- AliveCellsCount{CompletedTurns: *turn, CellsCount: checkNumberOfAliveCells(p, *world)}
    }
}

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

// Connect to server and run the Run method on there.
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

    go engine(p, distributorChannels)

    ioChannels := ioChannels{
        command:  ioCommand,
    	idle:     ioIdle,
    	filename: ioFilename,
    	output:   ioOutput,
    	input:    ioInput,
    }
    go startIo(p, ioChannels)
}

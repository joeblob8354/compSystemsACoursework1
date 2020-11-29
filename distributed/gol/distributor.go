package gol

import (
    "uk.ac.bris.cs/gameoflife/util"
    //"fmt"
    "strconv"
)

//all channels distributor has access to
type distributorChannels struct {
	events     chan<- Event
	ioCommand  chan<- ioCommand
	ioIdle     <-chan bool
	ioFilename chan<- string
	ioInput    <-chan uint8
	ioOutput   chan<- uint8
}

//finds a cells neighbours and increments amount if teh neighbour is alive
func findNeighbour(cellColumn int, cellRow int, rowChange int, columnChange int, world [][]byte, p Params, amount int) int {
    newRow := cellRow + rowChange
    newColumn := cellColumn + columnChange
    //loops round screen if limits exceeded
    if cellRow + rowChange < 0 {
        newRow = p.ImageHeight - 1
    } else if cellRow + rowChange > p.ImageHeight - 1 {
        newRow = 0
    }
    //loops round screen if limits exceeded
    if cellColumn + columnChange < 0 {
        newColumn = p.ImageWidth - 1
    } else if cellColumn + columnChange > p.ImageWidth - 1 {
        newColumn = 0
    }
    if world[newRow][newColumn] == 255 {
        amount++
    }
    return amount
}

//returns the number of neighbouring alive cells of a given cell
func getNumberOfNeighbours(cellColumn int, cellRow int, world [][]byte, p Params) int {
    amount := 0 //number of alive neighbour cells
    rowChange := []int{1, 1, 1, 0, 0, -1, -1, -1}
    columnChange := []int{-1, 0, 1, -1, 1, -1, 0, 1}
    for i := 0; i < 8; i++ {
        amount = findNeighbour(cellColumn, cellRow, rowChange[i], columnChange[i], world, p, amount)
    }
    return amount
}

//returns an updated state sending cell flipped events when necessary
func calculateNextState(p Params, world [][]byte, turn int, c distributorChannels) [][]byte {
	//creates a blank new state for us to populate
	newState := make([][]byte, p.ImageHeight)
	for i := 0; i < p.ImageHeight; i++ {
	    newState[i] = make([]byte, p.ImageWidth)
	}

	for currRow := 0; currRow < p.ImageHeight; currRow++ {
	    for currColumn := 0; currColumn < p.ImageWidth; currColumn++ {
	        numberOfNeighbours := getNumberOfNeighbours(currColumn, currRow, world, p)
            currentCell := world[currRow][currColumn]
            if currentCell == 255 {
                if numberOfNeighbours < 2 {
                    newState[currRow][currColumn] = 0
                    c.events <- CellFlipped{CompletedTurns: turn, Cell: util.Cell{X: currRow, Y: currColumn}}
                } else if numberOfNeighbours > 3 {
                    newState[currRow][currColumn] = 0
                    c.events <- CellFlipped{CompletedTurns: turn, Cell: util.Cell{X: currRow, Y: currColumn}}
                } else {
                    newState[currRow][currColumn] = 255
                  }
            } else if currentCell == 0 {
                if numberOfNeighbours == 3 {
                    newState[currRow][currColumn] = 255
                    c.events <- CellFlipped{CompletedTurns: turn, Cell: util.Cell{X: currRow, Y: currColumn}}
                } else {
                    newState[currRow][currColumn] = 0
                  }
              }
	    }
	}
    return newState
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

// distributor divides the work between workers and interacts with other goroutines.
func distributor(p Params, c distributorChannels) {

	//Creates a 2D slice to store the world.
	newWorld := make([][]byte, p.ImageHeight)
	for i := 0; i < p.ImageHeight; i++ {
	    newWorld[i] = make([]byte, p.ImageWidth)
	}
	//sends command to io so it executes the readPgmImage() function
    c.ioCommand <- ioInput
    //sends created filename to readPgmImage()
	c.ioFilename <- (strconv.Itoa(p.ImageHeight) + "x" + strconv.Itoa(p.ImageWidth))

    //reads in bytes 1 at a time from the ioInput channel and populates the world
	for currRow := 0; currRow < p.ImageHeight; currRow++ {
    	    for currColumn := 0; currColumn < p.ImageWidth; currColumn++ {
    	        newWorld[currRow][currColumn] = <- c.ioInput
    	    }
    }

	turn := 0

	//For all initially alive cells send a CellFlipped Event.
	for currRow := 0; currRow < p.ImageHeight; currRow++ {
	    for currColumn := 0; currColumn < p.ImageWidth; currColumn++ {
	        currentCell := newWorld[currRow][currColumn]
	        if currentCell == 255 {
	            c.events <- CellFlipped{CompletedTurns: turn, Cell: util.Cell{X: currRow, Y: currColumn}}
	        }
	    }
	}
    //changes state to executing
	c.events <- StateChange{CompletedTurns: turn, NewState: Executing}

	//Execute all turns of the Game of Life.
	for turn = 0; turn < p.Turns; turn++ {
	    newWorld = calculateNextState(p, newWorld, turn, c)
	    c.events <- TurnComplete{CompletedTurns: turn}
	}

    //send array of alive cells for testing
    aliveCells := calculateAliveCells(p, newWorld)
	c.events <- FinalTurnComplete{CompletedTurns: turn, Alive: aliveCells}

    //send command to io to let make it execute the writePgmImage() function.
	c.ioCommand <- ioOutput
	//send the filename to the writePgmImage() function.
	c.ioFilename <- (strconv.Itoa(p.ImageHeight) + "x" + strconv.Itoa(p.ImageWidth))
	//Scan across the updated world and send bytes 1 at a time to the writePgmImage() function via the ioOutput channel.
	for currRow := 0; currRow < p.ImageHeight; currRow++ {
        for currColumn := 0; currColumn < p.ImageWidth; currColumn++ {
        	c.ioOutput <- newWorld[currRow][currColumn]
        	}
    }
    // Make sure that the Io has finished any output before exiting.
	c.ioCommand <- ioCheckIdle
	<-c.ioIdle

    //updates state to quitting
	c.events <- StateChange{turn, Quitting}
	// Close the channel to stop the SDL goroutine gracefully. Removing may cause deadlock.
	close(c.events)
}
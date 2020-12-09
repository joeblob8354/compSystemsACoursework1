package gol

import (
	"uk.ac.bris.cs/gameoflife/util"
	"strconv"
//	"time"
//	"fmt"
    "os"
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

//finds a cells neighbours and increments amount if the neighbour is alive
func findNeighbour(cellColumn int, cellRow int, rowChange int, columnChange int, world [][]byte, p Params, amount int) int {
	newRow := cellRow + rowChange
	newColumn := cellColumn + columnChange

	//loops round screen if limits exceeded
	if newRow < 0 {
		newRow = p.ImageHeight - 1
	} else if newRow > p.ImageHeight-1 {
		newRow = 0
	}
	//loops round screen if limits exceeded
	if newColumn < 0 {
		newColumn = p.ImageWidth - 1
	} else if newColumn > p.ImageWidth-1 {
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
func calculateNextState(p Params, startY int, endY int, world [][]byte, turn int, c distributorChannels, out chan<- [][]byte) {
	//creates a blank new state for us to populate
	sectionHeight := endY - startY
	newState := make([][]byte, sectionHeight)
	for i := 0; i < sectionHeight; i++ {
		newState[i] = make([]byte, p.ImageWidth)
	}

	for currRow := startY; currRow < endY; currRow++ {
		for currColumn := 0; currColumn < p.ImageWidth; currColumn++ {
			numberOfNeighbours := getNumberOfNeighbours(currColumn, currRow, world, p)
			currentCell := world[currRow][currColumn]
			if currentCell == 255 {
				if numberOfNeighbours < 2 {
					newState[currRow-startY][currColumn] = 0
					c.events <- CellFlipped{CompletedTurns: turn, Cell: util.Cell{X: currRow, Y: currColumn}}
				} else if numberOfNeighbours > 3 {
					newState[currRow-startY][currColumn] = 0
					c.events <- CellFlipped{CompletedTurns: turn, Cell: util.Cell{X: currRow, Y: currColumn}}
				} else {
					newState[currRow-startY][currColumn] = 255
				}
			} else if currentCell == 0 {
				if numberOfNeighbours == 3 {
					newState[currRow-startY][currColumn] = 255
					c.events <- CellFlipped{CompletedTurns: turn, Cell: util.Cell{X: currRow, Y: currColumn}}
				} else {
					newState[currRow-startY][currColumn] = 0
				}
			}
		}
	}

	//send section channel
	out <- newState
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
func distributor(p Params, c distributorChannels, isClosed chan bool, sendAlive chan Event, tickerAvail chan bool, k <-chan rune) {

	//Creates a 2D slice to store the world.
	newWorld := make([][]byte, p.ImageHeight)
	for i := 0; i < p.ImageHeight; i++ {
		newWorld[i] = make([]byte, p.ImageWidth)
	}
	//sends command to io so it executes the readPgmImage() function
	c.ioCommand <- ioInput
	//sends created filename to readPgmImage()
	c.ioFilename <- (strconv.Itoa(p.ImageHeight) + "x" + strconv.Itoa(p.ImageWidth) /*+ "x" + strconv.Itoa(p.Turns)*/)

	//reads in bytes 1 at a time from the ioInput channel and populates the world
	for currRow := 0; currRow < p.ImageHeight; currRow++ {
		for currColumn := 0; currColumn < p.ImageWidth; currColumn++ {
			newWorld[currRow][currColumn] = <-c.ioInput 
		}
	}
	//sets start turn to 0.
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

	//notifies the ticker that the events channel is open
	isClosed <- false
	
	//sets number of sections == to number of threads (workers).
	numberOfSections := p.Threads
	//sets the height of each section to be an equal proportion of total height.
	heightOfSection := p.ImageHeight / p.Threads

	//creates a slice of channels that will be used to receive incoming completed sections from workers.
	chanSlice := make([]chan [][]byte, numberOfSections)
	for i := range chanSlice {
		chanSlice[i] = make(chan [][]byte)
	}

	//Execute all turns of the Game of Life.
	for turn = 0; turn < p.Turns; turn++ { 

		//if numberOfSections == 1 start just 1 worker routine to handle the entire world
		if numberOfSections == 1 {
			go calculateNextState(p, 0, p.ImageHeight, newWorld, turn, c, chanSlice[0])
			// if numberOfSections > 1 then create a worker go routine for each section and allocate it a section of the world to work on.
		} else {
			for section := 0; section < numberOfSections-1; section++ {
				go calculateNextState(p, 0+section*heightOfSection, heightOfSection+section*heightOfSection, newWorld, turn, c, chanSlice[section])
			}
			//this last section is separate to handle cases where number of workers is odd as a row can be left out otherwise due to rounding.
			go calculateNextState(p, (numberOfSections-1)*heightOfSection, p.ImageHeight, newWorld, turn, c, chanSlice[numberOfSections-1])
		}

		//sends the number of alive cells to the ticker
		if turn > 0 {
			select {
				case <-tickerAvail:
					alive := calculateAliveCells(p, newWorld)
					sendAlive <- AliveCellsCount{CompletedTurns: turn, CellsCount: len(alive)}
				default:

			}
			
		}
		
		//reset world state
		newWorld = nil

		//receives incoming sections from workers and appends them in order to the newWorld state.
		for section := 0; section < numberOfSections; section++ {
			part := <-chanSlice[section]
			newWorld = append(newWorld, part...)
		}
		
		//sends an event to say the turn is complete
		c.events <- TurnComplete{CompletedTurns: turn}

        var key rune
        //listening for incoming key-presses without blocking
        select {
            case key = <- k:
                //if s is pressed output a pgm img of the current world state and the corresponding turn.
                if key == 's' {
                    outputPgmFile(c, p, newWorld, turn)
                // if q is pressed, change state to quitting and exit.
                } else if key == 'q' {
                    c.events <- StateChange{CompletedTurns: turn, NewState: Quitting}
                    os.Exit(0)
                }
        }
	}

	//send array of alive cells for testing
	aliveCells := calculateAliveCells(p, newWorld)
	c.events <- FinalTurnComplete{CompletedTurns: turn, Alive: aliveCells}

	//send command to io to let make it execute the writePgmImage() function.
	c.ioCommand <- ioOutput
	//send the filename to the writePgmImage() function.
	c.ioFilename <- (strconv.Itoa(p.ImageHeight) + "x" + strconv.Itoa(p.ImageWidth) + "x" + strconv.Itoa(p.Turns))

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

	//notifies the ticker that c.events is closed
	isClosed <- true

}

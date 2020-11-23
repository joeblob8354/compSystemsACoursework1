package gol

type distributorChannels struct {
	events    chan<- Event
	ioCommand chan<- ioCommand
	ioIdle    <-chan bool
}

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

func getNumberOfNeighbours(cellColumn int, cellRow int, world [][]byte, p Params) int {
    amount := 0 //number of alive neighbour cells
    rowChange := []int{1, 1, 1, 0, 0, -1, -1, -1}
    columnChange := []int{-1, 0, 1, -1, 1, -1, 0, 1}
    for i := 0; i < 8; i++ {
        amount = findNeighbour(cellColumn, cellRow, rowChange[i], columnChange[i], world, p, amount)
    }
    return amount
}

func calculateNextState(p Params, world [][]byte) [][]byte {
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
                } else if numberOfNeighbours > 3 {
                    newState[currRow][currColumn] = 0
                } else {
                    newState[currRow][currColumn] = 255
                  }
            } else if currentCell == 0 {
                if numberOfNeighbours == 3 {
                    newState[currRow][currColumn] = 255
                } else {
                    newState[currRow][currColumn] = 0
                  }
              }
	    }
	}
    return newState
}

/*func calculateAliveCells(p golParams, world [][]byte) []cell {
	var aliveCells []cell
	for currRow := 0; currRow < p.imageHeight; currRow++ {
	    for currColumn := 0; currColumn < p.imageWidth; currColumn++ {
	        if world[currRow][currColumn] == 255 {
	            aliveCells = append(aliveCells, cell{x: currColumn, y: currRow})
	        }
	    }
	}

	return aliveCells
}*/

// distributor divides the work between workers and interacts with other goroutines.
func distributor(p Params, c distributorChannels) {

	// TODO: Create a 2D slice to store the world.
	world := make([][]byte, p.ImageHeight)
	for i := 0; i < p.ImageHeight; i++ {
	    world[i] = make([]byte, p.ImageWidth)
	}

	turn := 0

	// TODO: For all initially alive cells send a CellFlipped Event.
	for currRow := 0; currRow < p.ImageHeight; currRow++ {
	    for currColumn := 0; currColumn < p.ImageWidth; currColumn++ {
	        currentCell := world[currRow][currColumn]
	        if currentCell == 255 {
	            cellToBeFlipped := CellFlipped{CompletedTurns: 0, Cell: util.Cell{X: 0, Y:0}}
	        }
	    }
	}

	// TODO: Execute all turns of the Game of Life.
	for turn = 0; turn < p.Turns; turn++ {
	    world = calculateNextState(p, world)
	}

	// TODO: Send correct Events when required, e.g. CellFlipped, TurnComplete and FinalTurnComplete.
	//		 See event.go for a list of all events.

	// Make sure that the Io has finished any output before exiting.
	c.ioCommand <- ioCheckIdle
	<-c.ioIdle

	c.events <- StateChange{turn, Quitting}
	// Close the channel to stop the SDL goroutine gracefully. Removing may cause deadlock.
	close(c.events)
}

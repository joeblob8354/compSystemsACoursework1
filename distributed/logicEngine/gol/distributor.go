package gol

import (
    //"uk.ac.bris.cs/gameoflife/util"
    "fmt"
    //"strconv"
    "time"
)

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

func ticker (tk time.Ticker, p Params, world [][]byte) {

    numberOfAliveCells := 0
    for range tk.C{
        numberOfAliveCells = checkNumberOfAliveCells(p, world)
        fmt.Println(numberOfAliveCells)
    }
}

// distributor divides the work between workers and interacts with other goroutines.
func Distributor(p Params, world [][]byte) [][]byte {

    //Creates a 2D slice to store the world.
    newWorld := make([][]byte, p.ImageHeight)
    for i := 0; i < p.ImageHeight; i++ {
        newWorld[i] = make([]byte, p.ImageWidth)
    }
	turn := 0

    newWorld = world

    duration := time.Duration(2) * time.Second

    tk := time.NewTicker(duration)

    go ticker(*tk, p, newWorld)

	//Execute all turns of the Game of Life.
    for turn = 0; turn < p.Turns; turn++ {
	    newWorld = calculateNextState(p, newWorld)
	}
    tk.Stop()
	return newWorld
}
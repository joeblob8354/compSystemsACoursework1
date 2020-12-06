package gol

import (
    //"uk.ac.bris.cs/gameoflife/util"
    //"fmt"
    //"strconv"
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
func CalculateNextState(p Params, startY int, endY int, world [][]byte) [][]byte {

	sectionHeight := endY - startY
	//creates a blank new state for us to populate
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
                    newState[currRow - startY][currColumn] = 0
                } else if numberOfNeighbours > 3 {
                    newState[currRow - startY][currColumn] = 0
                } else {
                    newState[currRow - startY][currColumn] = 255
                  }
            } else if currentCell == 0 {
                if numberOfNeighbours == 3 {
                    newState[currRow - startY][currColumn] = 255
                } else {
                    newState[currRow - startY][currColumn] = 0
                  }
              }
	    }
	}
    return newState
}
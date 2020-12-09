package gol
import (
	"time"
	//"fmt"
)

// Params provides the details of how to run the Game of Life and which image to load.
type Params struct {
	Turns       int
	Threads     int
	ImageWidth  int
	ImageHeight int
}

func Simple(x int) {
    for n := 0; n == 1000; n++ {
        x = x + x
    }
}

// Run starts the processing of Game of Life. It should initialise channels and goroutines.
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

	//used for sending the number of alive cells from the distributor to the ticker
	sendAlive := make(chan Event)
	//used for notifiying the ticker function about whether events is open or not
	isClosed := make(chan bool)
	//used in the distributor function in the select statement to avoid blocking
	tickerAvail := make(chan bool)
	tickerKeyControl := make(chan bool)

	go distributor(p, distributorChannels, isClosed, sendAlive, tickerAvail, keyPresses, tickerKeyControl)
	//receive the number of alive cells every 2 seconds while the events channel is open, stop the goroutine when the channel is closed
	ticker := time.NewTicker(2 * time.Second)
	go func() {
		for {
			select {
				case closed := <-isClosed:
					if closed == true {
						return
					} 
				case <-ticker.C:
					tickerAvail <- true
					send := <-sendAlive
					events <- send 
				case pause := <-tickerKeyControl:
					if pause == true {
						for {
						if pause {
							pause = <-tickerKeyControl
							} else {
								break
							}
						
						}	
					}
				default:
					
			} 
		}
	}()

	ioChannels := ioChannels{
		command:  ioCommand,
		idle:     ioIdle,
		filename: ioFilename,
		output:   ioOutput,
		input:    ioInput,
	}
	go startIo(p, ioChannels)
}

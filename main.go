package main

import (
	"fmt"
	"io"
	"os"
	"os/signal"
	"regexp"
	"syscall"
	"time"
	"flag"
)

const USAGE = "Usage: ftail [-delay <time_delay>] <filename> <regex_pattern>\n" 
const BUFFER_SIZE = 4096

func main() {

	var help bool

    delay := flag.Int("delay", 100, "The time delay (ms) between file reads. Minimum is 1.")
	flag.BoolVar(&help, "h", false, "Display program usage")

	flag.Parse()

	if help {
		fmt.Fprintf(os.Stderr, USAGE)
		flag.PrintDefaults()
		os.Exit(0)
	}

	if flag.NArg() != 2 {
		fmt.Fprintf(os.Stderr, USAGE)
		os.Exit(1)
	}

	regexPattern := flag.Arg(0)
    filename := flag.Arg(1)

    if (*delay <= 0 ) {
        *delay = 1
    }

    time_delay := time.Duration(*delay)*time.Millisecond

	regex, err := regexp.Compile(regexPattern)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error compiling regex pattern: %v\n", err)
		os.Exit(1)
	}

	file, err := os.Open(filename)
	if err != nil {
		fmt.Fprintf(os.Stderr,"Error opening the file: %v\n", err)
		os.Exit(1)
	}
	defer file.Close()

	fileInfo, err := file.Stat()
	if err != nil {
		fmt.Fprintf(os.Stderr,"Error getting file info: %v\n", err)
		os.Exit(1)
	}
	initialSize := fileInfo.Size()
	currentPosition := initialSize

	interrupt := make(chan os.Signal, 1)
	signal.Notify(interrupt, os.Interrupt, syscall.SIGTERM)

	go func() {
		buffer := make([]byte, BUFFER_SIZE) 
		for {
			select {
			case <-interrupt:
				fmt.Println("Received termination signal. Exiting...")
				return
			default:
				fileInfo, err := file.Stat()
				if err != nil {
					fmt.Printf("Error getting file info: %v\n", err)
					time.Sleep(time.Second) 
					continue
				}

				if fileInfo.Size() > currentPosition {
					n, err := file.ReadAt(buffer, currentPosition)
					if err != nil && err != io.EOF {
						fmt.Printf("Error reading file: %v\n", err)
						time.Sleep(time.Second) 
						continue
					}

					if n > 0 {
						newData := buffer[:n]
						currentPosition += int64(n)

						if regex.Match(newData) {
							fmt.Printf("%s", newData)
						}
					}
				}

				time.Sleep(time_delay) 
			}
		}
	}()

	<-interrupt
}


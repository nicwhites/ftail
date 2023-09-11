package main

import (
	"fmt"
	"io"
	"os"
	"os/signal"
	"regexp"
	"syscall"
	"time"
)

func main() {
	if len(os.Args) != 3 {
		fmt.Println("Usage: ftail <regex_pattern> <filename>")
		os.Exit(1)
	}

	filename := os.Args[2]
	regexPattern := os.Args[1]


    sleep_time := 25 * time.Millisecond

	// Compile the regex pattern
	regex, err := regexp.Compile(regexPattern)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error compiling regex pattern: %v\n", err)
		os.Exit(1)
	}

	// Open the file for reading
	file, err := os.Open(filename)
	if err != nil {
		fmt.Fprintf(os.Stderr,"Error opening the file: %v\n", err)
		os.Exit(1)
	}
	defer file.Close()

	// Get the initial file size
	fileInfo, err := file.Stat()
	if err != nil {
		fmt.Fprintf(os.Stderr,"Error getting file info: %v\n", err)
		os.Exit(1)
	}
	initialSize := fileInfo.Size()
	currentPosition := initialSize

	// Create a channel to receive termination signals (Ctrl+C)
	interrupt := make(chan os.Signal, 1)
	signal.Notify(interrupt, os.Interrupt, syscall.SIGTERM)

	// Start a goroutine to periodically check and print matching new data
	go func() {
		buffer := make([]byte, 1024) // Adjust buffer size as needed
		for {
			select {
			case <-interrupt:
				fmt.Println("Received termination signal. Exiting...")
				return
			default:
				// Get the current file size
				fileInfo, err := file.Stat()
				if err != nil {
					fmt.Printf("Error getting file info: %v\n", err)
					time.Sleep(time.Second) // Wait before retrying
					continue
				}

				// If the file size has increased, read and check for matching new data
				if fileInfo.Size() > currentPosition {
					n, err := file.ReadAt(buffer, currentPosition)
					if err != nil && err != io.EOF {
						fmt.Printf("Error reading file: %v\n", err)
						time.Sleep(time.Second) // Wait before retrying
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

				time.Sleep(sleep_time) // Adjust the check frequency as needed
			}
		}
	}()

	// Block the main goroutine until a termination signal is received
	<-interrupt
}


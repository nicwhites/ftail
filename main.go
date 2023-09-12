package main

import (
	"flag"
	"fmt"
	"io"
	"os"
	"os/signal"
	"regexp"
	"syscall"
	"time"
)

const USAGE = "Usage: ftail [-delay <time_delay>] <regex_pattern> <filename>\n" 
const BUFFER_SIZE = 1024

func main() {

    // Flags 
	var help bool

    delay_int := flag.Int("delay", 100, "The time delay between file reads (ms)")
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

    delay := time.Duration(*delay_int) * time.Millisecond

	regexPattern := flag.Arg(0)
	filename := flag.Arg(1)
    // end of flags

	regex, err := regexp.Compile(regexPattern)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error compiling regex pattern: %v\n", err)
		os.Exit(1)
	}

	file, err := os.Open(filename)

	if err != nil {
		fmt.Fprintf(os.Stderr, "Error opening the file: %v\n", err)
		os.Exit(1)
	}

	defer file.Close()

	fileInfo, err := file.Stat()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error getting file info: %v\n", err)
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
				return
			default:
				fileInfo, err := file.Stat()
				if err != nil {
					fmt.Fprintf(os.Stderr, "Error getting file info: %v\n", err)
					time.Sleep(delay)
                    os.Exit(1)
				}

				currentSize := fileInfo.Size()
				if currentSize < currentPosition {
					file.Close()
					f, err := os.Open(filename)
                    file = f
					if err != nil {
						fmt.Fprintf(os.Stderr, "Error reopening the file: %v\n", err)
						os.Exit(1)
					}
					currentPosition = currentSize
					continue
				}

				if currentSize > currentPosition {
					n, err :=  file.ReadAt(buffer, currentPosition)
					if err != nil && err != io.EOF {
						fmt.Fprintf(os.Stderr, "Error reading file: %v\n", err)
						time.Sleep(delay)
						continue
					}

					if n > 0 {
                        newData := buffer[:n]
                        currentPosition += int64(n)
						if regex.Match(newData) {
							fmt.Printf("%s", newData)

							currentPosition = currentSize
						}
					}
				}

				time.Sleep(delay)
			}
		}
	}()

	<-interrupt
}


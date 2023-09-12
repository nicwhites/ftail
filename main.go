package main

import (
	"bufio"
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

func main() {

    // Flags 
	var delay time.Duration
	var help bool

	flag.DurationVar(&delay, "delay", 100*time.Millisecond, "The time delay between file reads (ms)")
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

	reader := bufio.NewReader(file)

	go func() {
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

					reader = bufio.NewReader(file)
					currentPosition = currentSize
					continue
				}

				if currentSize > currentPosition {
					newData, err := reader.ReadString('\n')
					if err != nil && err != io.EOF {
						fmt.Fprintf(os.Stderr, "Error reading file: %v\n", err)
						time.Sleep(delay)
						continue
					}

					if len(newData) > 0 {
						if regex.MatchString(newData) {
							fmt.Printf(newData)

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


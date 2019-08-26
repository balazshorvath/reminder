package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"os/signal"
	"strconv"
	"sync"
	"syscall"
	"time"

	"github.com/faiface/beep"
	"github.com/faiface/beep/mp3"
	"github.com/faiface/beep/speaker"
)

const (
	defaultInterval time.Duration = 10
)

var (
	interval time.Duration
)

func main() {
	if len(os.Args) < 0 {
		interval = defaultInterval
		fmt.Printf("Using default interval of %d minutes\n", interval)
	}
	num, err := strconv.ParseInt(os.Args[0], 10, 64)
	if err != nil {
		panic(err)
	}
	interval = time.Duration(num)
	run()
}

func run() {
	tick := time.NewTicker(interval * time.Minute).C
	group := &sync.WaitGroup{}
	quit := make(chan os.Signal)
	ok := make(chan interface{})

	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	ctx, cancel := context.WithCancel(context.Background())
	group.Add(1)
	go func() {
		defer group.Done()
		for {
			select {
			case <-tick:
				go func() {
					group.Add(1)
					defer group.Done()
					for {
						select {
						case <-ok:
							return
						case <-ctx.Done():
							return
						default:
							play()
						}
					}
				}()
				var s string
				_, _ = fmt.Scanln(&s)
				ok <- nil
			case <-quit:
				cancel()
				return
			}
		}
	}()
	// Wait for all the async tasks to finish.
	group.Wait()
}

func play() {
	file, err := os.Open("sound.mp3")
	if err != nil {
		panic(err)
	}
	streamer, format, err := mp3.Decode(file)
	if err != nil {
		log.Fatal(err)
	}
	defer streamer.Close()

	speaker.Init(format.SampleRate, format.SampleRate.N(time.Second/10))

	done := make(chan bool)
	speaker.Play(beep.Seq(streamer, beep.Callback(func() {
		done <- true
	})))
	fmt.Println("WAKE UP!!")

	<-done
}

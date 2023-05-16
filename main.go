package main

import (
	"bufio"
	"log"
	"os"
	"strconv"
	"strings"
	"sync"
	"time"
)

type RingBuffer struct {
	array         []int
	size          int
	drainInterval time.Duration
	pos           int
	m             sync.Mutex
}

func NewRingBuffer(size int, drainInterval time.Duration) *RingBuffer {
	return &RingBuffer{make([]int, size), size, drainInterval, -1, sync.Mutex{}}
}

func (r *RingBuffer) Push(e int) {
	r.m.Lock()
	defer r.m.Unlock()
	if r.pos == r.size-1 {
		for i := 1; i < r.size; i++ {
			r.array[i-1] = r.array[i]
		}
	} else {
		r.pos++
	}
	r.array[r.pos] = e
}

func (r *RingBuffer) Pop() (int, bool) {
	if r.pos < 0 {
		return 0, false
	}
	r.m.Lock()
	defer r.m.Unlock()
	var output = r.array[r.pos]
	r.pos--
	return output, true
}

func (r *RingBuffer) Get() []int {
	if r.pos < 0 {
		return nil
	}
	r.m.Lock()
	defer r.m.Unlock()
	var output = r.array[:r.pos+1]
	r.pos = -1
	return output
}

func read(input chan<- int, done chan bool) {
	scanner := bufio.NewScanner(os.Stdin)
	for scanner.Scan() {
		data := scanner.Text()
		if strings.EqualFold(data, "exit") {
			log.Println("The program has ended")
			close(done)
			return
		}
		i, err := strconv.Atoi(data)
		if err != nil {
			log.Println("The program only accepts integers")
			continue
		}
		input <- i
	}
}

func negativeFilter(input <-chan int, output chan<- int, done <-chan bool) {
	for {
		select {
		case data := <-input:
			if data >= 0 {
				output <- data
			} else {
				log.Println("Filtered by negative filter", data)
			}
		case <-done:
			return
		}
	}
}

func notDividedThreeFilter(input <-chan int, output chan<- int, done <-chan bool) {
	for {
		select {
		case data := <-input:
			if data%3 == 0 {
				output <- data
			} else {
				log.Println("Filtered by not divided three filter", data)
			}
		case <-done:
			return
		}
	}
}

func (r *RingBuffer) bufferStage(input <-chan int, output chan<- int, done <-chan bool) {
	for {
		select {
		case data := <-input:
			r.Push(data)
		case <-time.After(r.drainInterval):
			//for {
			//	e, err := r.Pop()
			//	if err {
			//		return
			//	}
			//	output <- e
			//}
			a := r.Get()
			if a != nil {
				for _, e := range a {
					output <- e
				}
			}
		case <-done:
			return
		}
	}
}

func main() {
	input := make(chan int)
	done := make(chan bool)
	log.Println("Start input data stage")
	go read(input, done)

	negativeFiltered := make(chan int)
	log.Println("Start negative filter stage")
	go negativeFilter(input, negativeFiltered, done)

	notDividedThreeFiltered := make(chan int)
	log.Println("Start not divided filter stage")
	go notDividedThreeFilter(negativeFiltered, notDividedThreeFiltered, done)

	buffer := NewRingBuffer(10, 10*time.Second)
	buffered := make(chan int)
	go buffer.bufferStage(notDividedThreeFiltered, buffered, done)

	for {
		select {
		case data := <-buffered:
			log.Println("Processed data, ", data)
		case <-done:
			return
		}
	}
}

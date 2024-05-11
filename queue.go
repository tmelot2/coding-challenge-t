package main

import (
	// "fmt"
	"sync"
)

type Job struct {
	query string
	start string
	end   string
	host  string
}

type Queue struct {
	jobs chan Job
	stop chan struct {}
	wg   sync.WaitGroup
}

func NewQueue() *Queue {
	return &Queue {
		jobs: make(chan Job),
		stop: make(chan struct{}),
	}
}

func (q *Queue) Start() {
	for {
		select {
		case job := <-q.jobs:
			go func(job Job) {
				processJob(q, job)
			} (job)
		case <-q.stop:
			return
		}
	}
}

func (q *Queue) Stop() {
	close(q.jobs)
	close(q.stop)
}

func (q *Queue) Enqueue(job Job) {
	q.wg.Add(1)
	q.jobs <- job
}

func processJob(q *Queue, job Job) {
	defer q.wg.Done()
}

// func main() {
// 	count := 3
// 	queue := NewQueue()
// 	go queue.Start()
// 	for i := 0; i < count; i++ {
// 		queue.Enqueue(i)
// 	}
// 	queue.wg.Wait()
// 	queue.Stop()
// }

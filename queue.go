package main

import (
	// "fmt"
	"sync"
	"time"
)

type Job struct {
	Start string
	End   string
	Host  string
	F func(start,end,host string) time.Duration
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
		case job := <- q.jobs:
			job.F(job.Start, job.End, job.Host)
			q.wg.Done()
		case <- q.stop:
			return
		}
	}
}

func (q *Queue) Wait() {
	q.wg.Wait()
}

func (q *Queue) Stop() {
	q.wg.Wait()
	close(q.jobs)
	close(q.stop)
}

func (q *Queue) Enqueue(job Job) {
	q.wg.Add(1)
	q.jobs <- job
}

// func processJob(q *Queue, job Job) {
// 	defer q.wg.Done()
// 	job.F(job.Start, job.End, job.Host)
// }

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

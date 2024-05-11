package main

import (
	"sync"
	"time"
)

// Job holds necessary data to run a db query, including a ref to the query function.
type Job struct {
	Start string
	End   string
	Host  string
	F func(start,end,host string) time.Duration
}

// Queue keeps track of running jobs.
type Queue struct {
	jobs chan Job
	stop chan struct {}
	wg   sync.WaitGroup
}

// Returns an empty queue.
func NewQueue() *Queue {
	return &Queue {
		jobs: make(chan Job),
		stop: make(chan struct{}),
	}
}

// Starts the queue, runs until the stop channel gets a signal.
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

// Tells the queue wait group to wait
func (q *Queue) Wait() {
	q.wg.Wait()
}

// Stops all channels in the queue
func (q *Queue) Stop() {
	close(q.jobs)
	close(q.stop)
}

// Adds a new job to the queue to be run
func (q *Queue) Enqueue(job Job) {
	q.wg.Add(1)
	q.jobs <- job
}

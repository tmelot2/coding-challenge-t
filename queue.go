package main

import (
	"fmt"
	"sync"
	"time"
)

// Job holds necessary data to run a db query, including a ref to the query function.
type Job struct {
	start  string
	end    string
	host   string
	// Function that performs the actual work of the job
	f      func(job Job) time.Duration
	jobNum int
}

// Queue keeps track of running jobs.
type Queue struct {
	jobs chan Job
	stop chan struct{}
	wg   sync.WaitGroup
}

// Returns an empty queue.
func NewQueue() *Queue {
	return &Queue{
		jobs: make(chan Job),
		stop: make(chan struct{}),
	}
}

// Starts the queue, runs until the stop channel gets a signal.
func (q *Queue) Start() {
	for {
		select {
		case job := <-q.jobs:
			elapsedTime := job.f(job)
			fmt.Printf("Job %d: Finished %s at %s, took %s\n", job.jobNum, job.host, time.Now(), elapsedTime)
			q.wg.Done()
		case <-q.stop:
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
	/*
	   NOTE: We *must* close stop chan before jobs chan, or else it will error with a nil pointer
	   dereference. I burned a LOT of time trying to figure that out.

	   When it's reversed, it errors on

	     Start(): elapsedTime := job.f(job)

	   Stop() only runs after Wait() has been called, ensuring that all jobs are finished. This
	   implies no new jobs are coming in. However that line runs with nil jobs, hence the error.

	   I think it makes sense to send the stop close signal first, so the Start() goroutine returns & ends.
	   This way, additional jobs that come in will be ignored, because nothing is watching the
	   channel anymore. Where I'm confused is there shouldn't *be* additional jobs, because at that
	   point we already waited on all the jobs to finish.
	*/
	close(q.stop)
	close(q.jobs)
}

// Adds a new job to the queue to be run
func (q *Queue) Enqueue(job Job) {
	q.wg.Add(1)
	fmt.Printf("Job %d: Submitting %s at %s\n", job.jobNum, job.host, time.Now())
	q.jobs <- job
}

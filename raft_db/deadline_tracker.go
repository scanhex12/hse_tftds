package raft_db

import (
	"time"
	"sync"
)

type DeadlineTracker struct {
	election_client *ElectionClient
	deadline_ms int
	last_warmup time.Time
	mutex       sync.Mutex
}

func NewDeadlineTracker(election_client *ElectionClient, deadline_ms int) *DeadlineTracker {
	return &DeadlineTracker{
		election_client: election_client,
		deadline_ms: deadline_ms,
	}
}

func (tracker *DeadlineTracker) Warmup() {
	tracker.mutex.Lock()
	tracker.last_warmup = time.Now()
	tracker.mutex.Unlock()

	go func(tracker *DeadlineTracker) {
		time.Sleep(time.Duration((tracker.deadline_ms + 1) * int(time.Millisecond)))
		
		milliseconds_passed := time.Now().Sub(tracker.last_warmup).Milliseconds()
		if milliseconds_passed >= int64(tracker.deadline_ms) {
			tracker.election_client.Run()
		}
	}(tracker)
}

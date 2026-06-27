//
// Copyright 2026 Nestybox, Inc.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//    https://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.
//

package lifecycleIO

import (
	"sync"
	"time"
)

type Operation string

const (
	DefaultLimit             = 2
	OperationRsync Operation = "rsync"
	OperationChown Operation = "chown"
)

var Default = NewLimiter(DefaultLimit)

type OpStats struct {
	Started          uint64
	Succeeded        uint64
	Failed           uint64
	LastDuration     time.Duration
	TotalDuration    time.Duration
	LastTarget       string
	LastError        string
	MaxObservedQueue int
}

type Stats struct {
	Rsync OpStats
	Chown OpStats
}

type Limiter struct {
	slots chan struct{}

	mu      sync.Mutex
	inQueue int
	stats   Stats
}

func NewLimiter(limit int) *Limiter {
	return &Limiter{slots: make(chan struct{}, limit)}
}

func (l *Limiter) Run(op Operation, target string, fn func() error) error {
	l.recordQueued(op)
	l.slots <- struct{}{}
	l.recordStarted(op, target)
	defer func() {
		<-l.slots
	}()

	startedAt := time.Now()
	err := fn()
	l.recordFinished(op, target, time.Since(startedAt), err)
	return err
}

func (l *Limiter) Stats() Stats {
	l.mu.Lock()
	defer l.mu.Unlock()

	return l.stats
}

func (l *Limiter) recordQueued(op Operation) {
	l.mu.Lock()
	defer l.mu.Unlock()

	l.inQueue++
	stats := l.opStats(op)
	if l.inQueue > stats.MaxObservedQueue {
		stats.MaxObservedQueue = l.inQueue
	}
	l.setOpStats(op, stats)
}

func (l *Limiter) recordStarted(op Operation, target string) {
	l.mu.Lock()
	defer l.mu.Unlock()

	l.inQueue--
	stats := l.opStats(op)
	stats.Started++
	stats.LastTarget = target
	l.setOpStats(op, stats)
}

func (l *Limiter) recordFinished(op Operation, target string, duration time.Duration, err error) {
	l.mu.Lock()
	defer l.mu.Unlock()

	stats := l.opStats(op)
	stats.LastDuration = duration
	stats.TotalDuration += duration
	stats.LastTarget = target
	if err != nil {
		stats.Failed++
		stats.LastError = err.Error()
	} else {
		stats.Succeeded++
		stats.LastError = ""
	}
	l.setOpStats(op, stats)
}

func (l *Limiter) opStats(op Operation) OpStats {
	if op == OperationChown {
		return l.stats.Chown
	}
	return l.stats.Rsync
}

func (l *Limiter) setOpStats(op Operation, stats OpStats) {
	if op == OperationChown {
		l.stats.Chown = stats
		return
	}
	l.stats.Rsync = stats
}

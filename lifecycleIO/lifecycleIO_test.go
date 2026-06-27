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
	"errors"
	"strings"
	"sync"
	"testing"
)

func TestLimiterRunLimitsConcurrencyAndRecordsStats(t *testing.T) {
	// Given
	limiter := NewLimiter(2)
	started := make(chan struct{}, 3)
	release := make(chan struct{})
	var wg sync.WaitGroup

	runBlockedOp := func() {
		defer wg.Done()
		_ = limiter.Run(OperationRsync, "/tmp/src", func() error {
			started <- struct{}{}
			<-release
			return nil
		})
	}

	// When
	wg.Add(3)
	go runBlockedOp()
	go runBlockedOp()
	go runBlockedOp()

	<-started
	<-started

	// Then
	select {
	case <-started:
		t.Fatal("third lifecycle IO operation started before a slot was released")
	default:
	}

	close(release)
	wg.Wait()

	stats := limiter.Stats()
	if stats.Rsync.Started != 3 {
		t.Fatalf("started rsync operations = %d, want 3", stats.Rsync.Started)
	}
	if stats.Rsync.Succeeded != 3 {
		t.Fatalf("successful rsync operations = %d, want 3", stats.Rsync.Succeeded)
	}
	if stats.Rsync.Failed != 0 {
		t.Fatalf("failed rsync operations = %d, want 0", stats.Rsync.Failed)
	}
}

func TestLimiterRunRecordsFailures(t *testing.T) {
	// Given
	limiter := NewLimiter(1)
	wantErr := errors.New("boom")

	// When
	gotErr := limiter.Run(OperationChown, "/tmp/tree", func() error {
		return wantErr
	})

	// Then
	if !errors.Is(gotErr, wantErr) {
		t.Fatalf("Run() error = %v, want %v", gotErr, wantErr)
	}

	stats := limiter.Stats()
	if stats.Chown.Started != 1 {
		t.Fatalf("started chown operations = %d, want 1", stats.Chown.Started)
	}
	if stats.Chown.Succeeded != 0 {
		t.Fatalf("successful chown operations = %d, want 0", stats.Chown.Succeeded)
	}
	if stats.Chown.Failed != 1 {
		t.Fatalf("failed chown operations = %d, want 1", stats.Chown.Failed)
	}
	if stats.Chown.LastError != "boom" {
		t.Fatalf("last chown error = %q, want %q", stats.Chown.LastError, "boom")
	}
}

func TestNewLimiterNormalizesInvalidLimit(t *testing.T) {
	// Given
	limiter := NewLimiter(0)

	// When
	err := limiter.Run(OperationRsync, "/tmp/src", func() error { return nil })

	// Then
	if err != nil {
		t.Fatalf("Run() error = %v, want nil", err)
	}
}

func TestLimiterStatsKeepPerOperationQueues(t *testing.T) {
	// Given
	limiter := NewLimiter(1)
	release := make(chan struct{})
	rsyncStarted := make(chan struct{})
	chownDone := make(chan struct{})

	go func() {
		_ = limiter.Run(OperationRsync, "/tmp/src", func() error {
			close(rsyncStarted)
			<-release
			return nil
		})
	}()
	<-rsyncStarted

	// When
	go func() {
		_ = limiter.Run(OperationChown, "/tmp/tree", func() error { return nil })
		close(chownDone)
	}()

	for {
		stats := limiter.Stats()
		if stats.Chown.MaxObservedQueue == 1 {
			if stats.Rsync.MaxObservedQueue != 1 {
				t.Fatalf("rsync max queue = %d, want 1", stats.Rsync.MaxObservedQueue)
			}
			close(release)
			<-chownDone
			return
		}
	}
}

func TestStatsStringExposesOperationMetrics(t *testing.T) {
	// Given
	limiter := NewLimiter(1)
	_ = limiter.Run(OperationRsync, "/tmp/src", func() error { return nil })
	_ = limiter.Run(OperationChown, "/tmp/tree", func() error { return errors.New("boom") })

	// When
	snapshot := limiter.Stats().String()

	// Then
	for _, want := range []string{
		"rsync started=1 succeeded=1 failed=0",
		"chown started=1 succeeded=0 failed=1",
		"last_target=/tmp/tree",
		"last_error=boom",
	} {
		if !strings.Contains(snapshot, want) {
			t.Fatalf("stats snapshot %q does not contain %q", snapshot, want)
		}
	}
}

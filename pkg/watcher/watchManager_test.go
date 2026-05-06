package watcher

import (
	"context"
	"sync/atomic"
	"testing"
	"time"

	"github.com/fsnotify/fsnotify"
	"github.com/microcks/microcks-cli/pkg/config"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestDebounceCollapsesRapidEvents(t *testing.T) {
	var callCount int32

	wm := &WatchManager{
		watchEntries: map[string]config.WatchEntry{
			"/tmp/test-spec.yaml": {FilePath: "/tmp/test-spec.yaml", Context: []string{"ctx1"}},
		},
		pending:     make(map[string]*time.Timer),
		importQueue: make(chan config.WatchEntry, 1),
		ctx:         context.Background(),
		triggerFunc: func(_ context.Context, _ config.WatchEntry) {
			atomic.AddInt32(&callCount, 1)
		},
	}

	go wm.worker()

	for i := 0; i < 5; i++ {
		wm.debounce("/tmp/test-spec.yaml", config.WatchEntry{FilePath: "/tmp/test-spec.yaml"})
	}

	time.Sleep(500 * time.Millisecond)

	wm.Stop()

	assert.Equal(t, int32(1), atomic.LoadInt32(&callCount), "debounce should collapse 5 rapid events into 1 import")
}

func TestDebounceDifferentFilesQueued(t *testing.T) {
	var calls []string
	var callMu chan struct{} = make(chan struct{}, 10)

	wm := &WatchManager{
		watchEntries: map[string]config.WatchEntry{
			"/tmp/a.yaml": {FilePath: "/tmp/a.yaml", Context: []string{"ctx1"}},
			"/tmp/b.yaml": {FilePath: "/tmp/b.yaml", Context: []string{"ctx1"}},
		},
		pending:     make(map[string]*time.Timer),
		importQueue: make(chan config.WatchEntry, 1),
		ctx:         context.Background(),
		triggerFunc: func(_ context.Context, entry config.WatchEntry) {
			calls = append(calls, entry.FilePath)
			callMu <- struct{}{}
		},
	}

	go wm.worker()

	wm.debounce("/tmp/a.yaml", config.WatchEntry{FilePath: "/tmp/a.yaml"})
	wm.debounce("/tmp/b.yaml", config.WatchEntry{FilePath: "/tmp/b.yaml"})

	time.Sleep(500 * time.Millisecond)

	assert.Len(t, calls, 2, "two different files should each trigger one import")
	assert.Contains(t, calls, "/tmp/a.yaml")
	assert.Contains(t, calls, "/tmp/b.yaml")

	wm.Stop()
}

func TestContextCancellationStopsRun(t *testing.T) {
	fw, err := fsnotify.NewWatcher()
	require.NoError(t, err)

	ctx, cancel := context.WithCancel(context.Background())

	wm := &WatchManager{
		fileWatcher:  fw,
		watchEntries: make(map[string]config.WatchEntry),
		pending:      make(map[string]*time.Timer),
		importQueue:  make(chan config.WatchEntry, 1),
		ctx:          ctx,
		cancel:       cancel,
		triggerFunc:  func(_ context.Context, _ config.WatchEntry) {},
	}

	done := make(chan struct{})
	go func() {
		wm.Run()
		close(done)
	}()

	cancel()

	select {
	case <-done:
	case <-time.After(2 * time.Second):
		t.Fatal("Run() did not exit after context cancellation")
	}
}

func TestHandleEventIgnoresIrrelevantOps(t *testing.T) {
	fw, err := fsnotify.NewWatcher()
	require.NoError(t, err)

	var callCount int32

	ctx, cancel := context.WithCancel(context.Background())

	wm := &WatchManager{
		fileWatcher: fw,
		watchEntries: map[string]config.WatchEntry{
			"/tmp/test.yaml": {FilePath: "/tmp/test.yaml", Context: []string{"ctx1"}},
		},
		pending:     make(map[string]*time.Timer),
		importQueue: make(chan config.WatchEntry, 1),
		ctx:         ctx,
		cancel:      cancel,
		triggerFunc: func(_ context.Context, _ config.WatchEntry) {
			atomic.AddInt32(&callCount, 1)
		},
	}

	wm.handleEvent(fsnotify.Event{Name: "/tmp/test.yaml", Op: fsnotify.Chmod})

	time.Sleep(100 * time.Millisecond)

	assert.Equal(t, int32(0), atomic.LoadInt32(&callCount), "Chmod events should be ignored")

	cancel()
}

func TestHandleEventProcessesWrite(t *testing.T) {
	var callCount int32

	ctx, cancel := context.WithCancel(context.Background())

	wm := &WatchManager{
		watchEntries: map[string]config.WatchEntry{
			"/tmp/test.yaml": {FilePath: "/tmp/test.yaml", Context: []string{"ctx1"}},
		},
		pending:     make(map[string]*time.Timer),
		importQueue: make(chan config.WatchEntry, 1),
		ctx:         ctx,
		cancel:      cancel,
		triggerFunc: func(_ context.Context, _ config.WatchEntry) {
			atomic.AddInt32(&callCount, 1)
		},
	}

	go wm.worker()

	wm.handleEvent(fsnotify.Event{Name: "/tmp/test.yaml", Op: fsnotify.Write})

	time.Sleep(500 * time.Millisecond)

	assert.Equal(t, int32(1), atomic.LoadInt32(&callCount), "Write event should trigger import")

	wm.Stop()
}

func TestHandleEventProcessesCreateAndRename(t *testing.T) {
	var callCount int32

	ctx, cancel := context.WithCancel(context.Background())

	fw, err := fsnotify.NewWatcher()
	require.NoError(t, err)

	wm := &WatchManager{
		fileWatcher: fw,
		watchEntries: map[string]config.WatchEntry{
			"/tmp/test.yaml": {FilePath: "/tmp/test.yaml", Context: []string{"ctx1"}},
		},
		pending:     make(map[string]*time.Timer),
		importQueue: make(chan config.WatchEntry, 1),
		ctx:         ctx,
		cancel:      cancel,
		triggerFunc: func(_ context.Context, _ config.WatchEntry) {
			atomic.AddInt32(&callCount, 1)
		},
	}

	go wm.worker()

	wm.handleEvent(fsnotify.Event{Name: "/tmp/test.yaml", Op: fsnotify.Create})
	wm.handleEvent(fsnotify.Event{Name: "/tmp/test.yaml", Op: fsnotify.Rename})

	time.Sleep(500 * time.Millisecond)

	assert.Equal(t, int32(1), atomic.LoadInt32(&callCount), "Create+Rename bursts should be debounced to 1 import")

	wm.Stop()
}

func TestDrainPendingTimers(t *testing.T) {
	wm := &WatchManager{
		pending: make(map[string]*time.Timer),
	}

	wm.pending["a"] = time.AfterFunc(10*time.Second, func() {})
	wm.pending["b"] = time.AfterFunc(10*time.Second, func() {})

	wm.drainPendingTimers()

	assert.Len(t, wm.pending, 0, "all pending timers should be drained")
}

func TestStopCancelsContext(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	wm := &WatchManager{
		ctx:    ctx,
		cancel: cancel,
	}

	wm.Stop()

	assert.Error(t, wm.ctx.Err(), context.Canceled, "Stop() should cancel the context")
}

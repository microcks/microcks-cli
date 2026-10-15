/*
 * Copyright The Microcks Authors.
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 *  http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 */

package watcher

import (
	"context"
	"os"
	"path/filepath"
	"sync/atomic"
	"testing"
	"time"

	"github.com/microcks/microcks-cli/pkg/config"
)

func startTestWatcher(t *testing.T, filenames []string, trigger func(context.Context, config.WatchEntry)) ([]string, context.CancelFunc) {
	t.Helper()

	dir := t.TempDir()
	files := make([]string, 0, len(filenames))
	entries := make([]config.WatchEntry, 0, len(filenames))
	for _, filename := range filenames {
		filePath := filepath.Join(dir, filename)
		if err := os.WriteFile(filePath, []byte("initial"), 0600); err != nil {
			t.Fatal(err)
		}
		files = append(files, filePath)
		entries = append(entries, config.WatchEntry{
			FilePath: filePath,
			Context:  []string{"test"},
		})
	}

	configPath := filepath.Join(dir, "watch.yaml")
	if err := config.WriteLocalWatchConfig(config.WatchConfig{Entries: entries}, configPath); err != nil {
		t.Fatal(err)
	}

	manager, err := NewWatchManger(configPath)
	if err != nil {
		t.Fatal(err)
	}
	manager.triggerImport = trigger

	ctx, cancel := context.WithCancel(context.Background())
	done := make(chan struct{})
	go func() {
		manager.RunContext(ctx)
		close(done)
	}()
	t.Cleanup(func() {
		cancel()
		select {
		case <-done:
		case <-time.After(2 * time.Second):
			t.Error("watch manager did not stop after cancellation")
		}
	})

	return files, cancel
}

func TestRunContextDebouncesWriteBurst(t *testing.T) {
	imports := make(chan string, 10)
	files, cancel := startTestWatcher(t, []string{"api.yaml"}, func(_ context.Context, entry config.WatchEntry) {
		imports <- entry.FilePath
	})
	defer cancel()

	for i := 0; i < 20; i++ {
		if err := os.WriteFile(files[0], []byte("updated"), 0600); err != nil {
			t.Fatal(err)
		}
		time.Sleep(10 * time.Millisecond)
	}

	select {
	case got := <-imports:
		if got != files[0] {
			t.Fatalf("imported %q, want %q", got, files[0])
		}
	case <-time.After(3 * time.Second):
		t.Fatal("timed out waiting for debounced import")
	}

	select {
	case got := <-imports:
		t.Fatalf("write burst triggered a duplicate import for %q", got)
	case <-time.After(450 * time.Millisecond):
	}
}

func TestRunContextSerializesImportsAndWatchesAtomicReplacement(t *testing.T) {
	started := make(chan string, 10)
	release := make(chan struct{}, 10)
	var active atomic.Int32
	var maxActive atomic.Int32
	files, cancel := startTestWatcher(t, []string{"api.yaml"}, func(ctx context.Context, entry config.WatchEntry) {
		current := active.Add(1)
		for observed := maxActive.Load(); current > observed && !maxActive.CompareAndSwap(observed, current); observed = maxActive.Load() {
		}
		started <- entry.FilePath
		select {
		case <-release:
		case <-ctx.Done():
		}
		active.Add(-1)
	})
	defer cancel()

	target := files[0]
	temp := target + ".tmp"
	if err := os.WriteFile(temp, []byte("replacement"), 0600); err != nil {
		t.Fatal(err)
	}
	if err := os.Remove(target); err != nil {
		t.Fatal(err)
	}
	if err := os.Rename(temp, target); err != nil {
		t.Fatal(err)
	}

	select {
	case got := <-started:
		if got != target {
			t.Fatalf("imported %q, want %q", got, target)
		}
	case <-time.After(3 * time.Second):
		t.Fatal("atomic replacement did not trigger an import")
	}

	if err := os.WriteFile(target, []byte("changed during import"), 0600); err != nil {
		t.Fatal(err)
	}
	time.Sleep(importDebounce + 100*time.Millisecond)
	select {
	case got := <-started:
		t.Fatalf("started concurrent import for %q", got)
	default:
	}

	release <- struct{}{}
	select {
	case got := <-started:
		if got != target {
			t.Fatalf("imported %q, want %q", got, target)
		}
	case <-time.After(3 * time.Second):
		t.Fatal("change during import was not imported")
	}
	if got := maxActive.Load(); got != 1 {
		t.Fatalf("maximum concurrent imports = %d, want 1", got)
	}
	release <- struct{}{}
}

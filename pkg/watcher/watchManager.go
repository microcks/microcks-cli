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
	"fmt"
	"log"
	"path/filepath"
	"sync"
	"time"

	"github.com/fsnotify/fsnotify"
	"github.com/microcks/microcks-cli/pkg/config"
)

const importDebounce = 300 * time.Millisecond

type pendingImport struct {
	entry    config.WatchEntry
	deadline time.Time
}

type WatchManager struct {
	fileWatcher   *fsnotify.Watcher
	configPath    string
	watchEntries  map[string]config.WatchEntry
	watchDirs     map[string]struct{}
	triggerImport func(context.Context, config.WatchEntry)
	lock          sync.Mutex
}

func NewWatchManger(configPath string) (*WatchManager, error) {
	absConfigPath, err := filepath.Abs(configPath)
	if err != nil {
		return nil, err
	}

	fw, err := fsnotify.NewWatcher()
	if err != nil {
		return nil, err
	}

	configDir := filepath.Dir(absConfigPath)
	if err := fw.Add(configDir); err != nil {
		fw.Close()
		return nil, err
	}

	wm := &WatchManager{
		fileWatcher:  fw,
		configPath:   absConfigPath,
		watchEntries: make(map[string]config.WatchEntry),
		watchDirs:    map[string]struct{}{configDir: {}},
	}

	if err := wm.Reload(); err != nil {
		fw.Close()
		return nil, err
	}

	return wm, nil
}

func (wm *WatchManager) Reload() error {
	cfg, err := LoadRegistry(wm.configPath)
	if err != nil {
		return err
	}

	newFiles := map[string]config.WatchEntry{}
	newDirs := map[string]struct{}{filepath.Dir(wm.configPath): {}}
	for _, entry := range cfg.Entries {
		absPath, err := filepath.Abs(entry.FilePath)
		if err != nil {
			log.Printf("[WARN] Cannot resolve watched file %s: %v", entry.FilePath, err)
			continue
		}
		newFiles[filepath.Clean(absPath)] = entry
		newDirs[filepath.Dir(filepath.Clean(absPath))] = struct{}{}
	}

	// Watch containing directories so replacing a file atomically doesn't
	// detach the watcher from its inode.
	for dir := range newDirs {
		if _, exists := wm.watchDirs[dir]; exists {
			continue
		}
		if err := wm.fileWatcher.Add(dir); err != nil {
			log.Printf("[WARN] Cannot watch directory %s: %v", dir, err)
			delete(newDirs, dir)
			continue
		}
		log.Printf("[INFO] Watcher added on directory %s", dir)
	}

	for dir := range wm.watchDirs {
		if _, exists := newDirs[dir]; exists {
			continue
		}
		if err := wm.fileWatcher.Remove(dir); err != nil {
			log.Printf("[WARN] Cannot remove watcher for directory %s: %v", dir, err)
		}
	}

	for file := range newFiles {
		if _, exists := wm.watchEntries[file]; !exists {
			log.Printf("[INFO] Watching file %s", file)
		}
	}

	wm.watchEntries = newFiles
	wm.watchDirs = newDirs
	return nil
}

func (wm *WatchManager) Run() {
	wm.RunContext(context.Background())
}

func (wm *WatchManager) RunContext(ctx context.Context) {
	runCtx, cancel := context.WithCancel(ctx)
	defer cancel()
	defer wm.fileWatcher.Close()

	imports := make(chan config.WatchEntry)
	go func() {
		for {
			select {
			case <-runCtx.Done():
				return
			case entry := <-imports:
				if runCtx.Err() != nil {
					return
				}
				if wm.triggerImport != nil {
					wm.triggerImport(runCtx, entry)
				} else {
					TriggerImportWithContext(runCtx, entry)
				}
			}
		}
	}()

	pending := make(map[string]pendingImport)
	queued := make(map[string]config.WatchEntry)
	ready := make([]string, 0)
	debounceTimer := time.NewTimer(time.Hour)
	if !debounceTimer.Stop() {
		<-debounceTimer.C
	}
	defer debounceTimer.Stop()

	eventErrors := wm.fileWatcher.Errors
	for {
		var debounceC <-chan time.Time
		if len(pending) > 0 {
			var deadline time.Time
			for _, item := range pending {
				if deadline.IsZero() || item.deadline.Before(deadline) {
					deadline = item.deadline
				}
			}
			if !debounceTimer.Stop() {
				select {
				case <-debounceTimer.C:
				default:
				}
			}
			delay := time.Until(deadline)
			if delay < 0 {
				delay = 0
			}
			debounceTimer.Reset(delay)
			debounceC = debounceTimer.C
		}

		var importC chan config.WatchEntry
		var nextImport config.WatchEntry
		if len(ready) > 0 {
			importC = imports
			nextImport = queued[ready[0]]
		}

		select {
		case <-runCtx.Done():
			return
		case event, ok := <-wm.fileWatcher.Events:
			if !ok {
				return
			}
			if event.Op&(fsnotify.Write|fsnotify.Create|fsnotify.Rename) == 0 {
				continue
			}

			eventPath, err := filepath.Abs(event.Name)
			if err != nil {
				continue
			}
			eventPath = filepath.Clean(eventPath)
			if eventPath == wm.configPath {
				fmt.Println("[INFO] Reloading config...")
				wm.lock.Lock()
				err := wm.Reload()
				wm.lock.Unlock()
				if err != nil {
					log.Printf("[ERROR] Config reload failed, keeping previous config: %v", err)
				}
				continue
			}

			wm.lock.Lock()
			entry, exists := wm.watchEntries[eventPath]
			wm.lock.Unlock()
			if exists {
				pending[eventPath] = pendingImport{
					entry:    entry,
					deadline: time.Now().Add(importDebounce),
				}
			}
		case <-debounceC:
			now := time.Now()
			for file, item := range pending {
				if item.deadline.After(now) {
					continue
				}
				delete(pending, file)
				if _, exists := queued[file]; exists {
					queued[file] = item.entry
					continue
				}
				queued[file] = item.entry
				ready = append(ready, file)
			}
		case importC <- nextImport:
			delete(queued, ready[0])
			ready = ready[1:]
		case err, ok := <-eventErrors:
			if !ok {
				eventErrors = nil
				continue
			}
			log.Printf("[ERROR] Watcher error: %v", err)
		}
	}
}

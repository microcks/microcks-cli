package watcher

import (
	"context"
	"fmt"
	"log"
	"sync"
	"time"

	"github.com/fsnotify/fsnotify"
	"github.com/microcks/microcks-cli/pkg/config"
	"github.com/microcks/microcks-cli/pkg/errors"
)

const debounceInterval = 300 * time.Millisecond

type WatchManager struct {
	fileWatcher  *fsnotify.Watcher
	configPath   string
	watchEntries map[string]config.WatchEntry
	lock         sync.Mutex
	ctx          context.Context
	cancel       context.CancelFunc
	pending      map[string]*time.Timer
	importQueue  chan config.WatchEntry
	triggerFunc  func(ctx context.Context, entry config.WatchEntry)
}

func NewWatchManger(ctx context.Context, configPath string) (*WatchManager, error) {
	fw, err := fsnotify.NewWatcher()
	if err != nil {
		return nil, err
	}

	err = fw.Add(configPath)
	if err != nil {
		return nil, err
	}

	childCtx, cancel := context.WithCancel(ctx)

	wm := &WatchManager{
		fileWatcher:  fw,
		configPath:   configPath,
		watchEntries: make(map[string]config.WatchEntry),
		ctx:          childCtx,
		cancel:       cancel,
		pending:      make(map[string]*time.Timer),
		importQueue:  make(chan config.WatchEntry, 1),
		triggerFunc:  TriggerImport,
	}

	err = wm.Reload()
	if err != nil {
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
	for _, entry := range cfg.Entries {
		newFiles[entry.FilePath] = entry
	}

	for file := range wm.watchEntries {
		if _, exists := newFiles[file]; !exists {
			wm.fileWatcher.Remove(file)
		}
	}

	for file := range newFiles {
		if _, exists := wm.watchEntries[file]; !exists {
			err := wm.fileWatcher.Add(file)
			if err != nil {
				log.Printf("[WARN] Cannot watch file %s: %v", file, err)
				continue
			}
			log.Printf("[INFO] Watcher added on %s", file)
		}
	}

	wm.watchEntries = newFiles
	return nil
}

func (wm *WatchManager) Stop() {
	if wm.cancel != nil {
		wm.cancel()
	}
}

func (wm *WatchManager) Run() {
	go wm.worker()

	for {
		select {
		case <-wm.ctx.Done():
			wm.drainPendingTimers()
			close(wm.importQueue)
			return
		case event := <-wm.fileWatcher.Events:
			wm.handleEvent(event)
		case err := <-wm.fileWatcher.Errors:
			log.Printf("[ERROR] Watcher error: %v", err)
		}
	}
}

func (wm *WatchManager) handleEvent(event fsnotify.Event) {
	if event.Op&(fsnotify.Write|fsnotify.Create|fsnotify.Rename) == 0 {
		return
	}

	if event.Name == wm.configPath {
		if event.Op&fsnotify.Write == fsnotify.Write {
			fmt.Println("[INFO] Reloading config...")
			wm.lock.Lock()
			err := wm.Reload()
			wm.lock.Unlock()
			if err != nil {
				errors.CheckError(err)
			}
		}
		return
	}

	wm.lock.Lock()
	entry, exists := wm.watchEntries[event.Name]
	wm.lock.Unlock()

	if !exists {
		return
	}

	if event.Op&(fsnotify.Create|fsnotify.Rename) != 0 {
		wm.fileWatcher.Remove(event.Name)
		if err := wm.fileWatcher.Add(event.Name); err != nil {
			log.Printf("[WARN] Cannot re-watch file %s: %v", event.Name, err)
		}
	}

	wm.debounce(event.Name, entry)
}

func (wm *WatchManager) debounce(path string, entry config.WatchEntry) {
	wm.lock.Lock()
	defer wm.lock.Unlock()

	if t, ok := wm.pending[path]; ok {
		t.Stop()
		delete(wm.pending, path)
	}

	wm.pending[path] = time.AfterFunc(debounceInterval, func() {
		wm.lock.Lock()
		delete(wm.pending, path)
		wm.lock.Unlock()

		select {
		case <-wm.ctx.Done():
			return
		case wm.importQueue <- entry:
		}
	})
}

func (wm *WatchManager) drainPendingTimers() {
	wm.lock.Lock()
	defer wm.lock.Unlock()

	for path, t := range wm.pending {
		t.Stop()
		delete(wm.pending, path)
	}
}

func (wm *WatchManager) worker() {
	for entry := range wm.importQueue {
		wm.triggerFunc(wm.ctx, entry)
	}
}

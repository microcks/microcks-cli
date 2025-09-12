package watcher

import (
	"fmt"
	"log"
	"sync"

	"github.com/fsnotify/fsnotify"
	"github.com/microcks/microcks-cli/pkg/config"
	"github.com/microcks/microcks-cli/pkg/errors"
)

type WatchManager struct {
	fileWatcher  *fsnotify.Watcher
	configPath   string
	watchEntries map[string]config.WatchEntry
	lock         sync.Mutex
}

func NewWatchManger(configPath string) (*WatchManager, error) {
	fw, err := fsnotify.NewWatcher()
	if err != nil {
		return nil, err
	}

	err = fw.Add(configPath)
	if err != nil {
		return nil, err
	}

	wm := &WatchManager{
		fileWatcher:  fw,
		configPath:   configPath,
		watchEntries: make(map[string]config.WatchEntry),
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

	// Remove stale watchers
	for file := range wm.watchEntries {
		if _, exists := newFiles[file]; !exists {
			wm.fileWatcher.Remove(file)
		}
	}

	// Add new watchers
	for file := range newFiles {
		if _, exists := wm.watchEntries[file]; !exists {
			err := wm.fileWatcher.Add(file)
			if err != nil {
				log.Printf("[WARN] Cannot watch file %s: %v", file, err)
				continue
			}
		}
	}

	wm.watchEntries = newFiles
	return nil
}

func (wm *WatchManager) Run() {
	for {
		select {
		case event := <-wm.fileWatcher.Events:
			if event.Op&fsnotify.Write == fsnotify.Write {
				if event.Name == wm.configPath {
					fmt.Println("[INFO] Reloading config...")
					wm.lock.Lock()
					err := wm.Reload()
					wm.lock.Unlock()
					if err != nil {
						errors.CheckError(err)
					}
				} else {
					wm.lock.Lock()
					entry, exists := wm.watchEntries[event.Name]
					wm.lock.Unlock()
					if exists {
						go TriggerImport(entry)
					}
				}
			}
		case err := <-wm.fileWatcher.Errors:
			log.Printf("[ERROR] Watcher error: %v", err)
		}
	}
}

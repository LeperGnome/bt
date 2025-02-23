package tree

import (
	"github.com/fsnotify/fsnotify"
)

type NodeChange struct {
	Path string
}

func runFSWatcher(watcher *fsnotify.Watcher) <-chan NodeChange {
	ch := make(chan NodeChange)
	go func() {
		defer watcher.Close()
		for {
			select {
			case event, ok := <-watcher.Events:
				if !ok {
					return
				}
				if event.Has(fsnotify.Remove) || event.Has(fsnotify.Create) || event.Has(fsnotify.Rename) || event.Has(fsnotify.Write) {
					ch <- NodeChange{Path: event.Name}
				}
			case _, ok := <-watcher.Errors:
				if !ok {
					return
				}
			}
		}
	}()
	return ch
}

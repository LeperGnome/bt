package state

import (
	"github.com/fsnotify/fsnotify"
)

type NodeChange struct{} // TODO

func runFSWatcher(watcher *fsnotify.Watcher) <-chan NodeChange { // TODO
	ch := make(chan NodeChange)
	go func() {
		defer watcher.Close()
		for {
			select {
			case event, ok := <-watcher.Events:
				if !ok {
					return
				}
				if event.Has(fsnotify.Remove) {
					// TODO
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

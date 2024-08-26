package state

import (
	"time"
)

type NodeChange struct{} // TODO

func InitFakeFSWatcher() <-chan NodeChange { // TODO
	ch := make(chan NodeChange)
	go func() {
		for {
			time.Sleep(time.Second)
			ch <- NodeChange{}
		}
	}()
	return ch
}

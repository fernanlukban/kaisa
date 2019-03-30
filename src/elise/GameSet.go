package gameset

import "sync"

type GameSet struct {
	mutex *sync.Mutex
}

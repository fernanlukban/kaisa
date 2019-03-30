package elise

import "sync"

// Set is an interface used to add and check existence of items
type Set interface {
	Add(item int64) bool
	Exists(item int64) bool
}

// GameSet works to keep track of the games seen
type GameSet struct {
	mutex *sync.Mutex
	data  map[int64]bool
}

// Add adds item to GameSet
func (gs *GameSet) Add(item int64) bool {
	gs.mutex.Lock()
	gs.data[item] = true
	gs.mutex.Unlock()
	return true
}

// Exists returns whether or not item exists
func (gs *GameSet) Exists(item int64) bool {
	gs.mutex.Lock()
	_, exists := gs.data[item]
	gs.mutex.Unlock()
	return exists
}

// New returns a new GameSet
func New() *GameSet {
	return &GameSet{
		mutex: &sync.Mutex{},
		data:  make(map[int64]bool),
	}
}

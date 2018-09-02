package services

import (
	"sync"
	"sync/atomic"
)

type StateStore struct {
	lock sync.Mutex
	rack atomic.Value
}

type Map = map[string]interface{}

func NewStateStore() (*StateStore) {
	ss := &StateStore{}
	ss.rack.Store(make(Map))
	return ss
}

func (ss *StateStore) Load(key string) (interface{}) {
	rack := ss.rack.Load().(Map)
	return rack[key]
}

func (ss *StateStore) Save(key string, val interface{}) (*StateStore) {
	ss.lock.Lock()
	defer ss.lock.Unlock()
	m1 := ss.rack.Load().(Map)
	m2 := make(Map)
	for k, v := range m1 {
		m2[k] = v
	}
	m2[key] = val
	ss.rack.Store(m2)
	return ss
}

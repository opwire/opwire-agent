package services

import (
	"encoding/json"
	"sync"
	"sync/atomic"
)

type StateStore struct {
	lock sync.Mutex
	rack atomic.Value
	cachedJSON map[string][]byte
}

type Map = map[string]interface{}

func NewStateStore() (*StateStore, error) {
	ss := &StateStore{}
	ss.rack.Store(make(Map))
	ss.cachedJSON = make(map[string][]byte)
	return ss, nil
}

func (ss *StateStore) Get(key string) (interface{}) {
	rack := ss.rack.Load().(Map)
	return rack[key]
}

func (ss *StateStore) GetAsJSON(key string) ([]byte, error) {
	if cachedJSON, ok := ss.cachedJSON[key]; ok {
		return cachedJSON, nil
	}
	if ref := ss.Get(key); ref != nil {
		var err error = nil
		ss.cachedJSON[key], err = json.Marshal(ref)
		return ss.cachedJSON[key], err
	}
	return nil, nil
}

func (ss *StateStore) Store(key string, val interface{}) (*StateStore) {
	ss.lock.Lock()
	defer ss.lock.Unlock()
	m1 := ss.rack.Load().(Map)
	m2 := make(Map)
	for k, v := range m1 {
		m2[k] = v
	}
	m2[key] = val
	ss.rack.Store(m2)
	delete(ss.cachedJSON, key)
	return ss
}

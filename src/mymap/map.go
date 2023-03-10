package ordered_sync_map

import (
	"container/list"
	"sync"
)

type mapElement struct {
	key   interface{}
	value interface{}
}

// Map is a thread safe and ordered implementation of standard map.
type Map struct {
	mp  map[interface{}]*list.Element
	mu  sync.RWMutex
	dll *list.List
}

// New returns an initialized Map.
func New() *Map {
	m := new(Map)
	m.mp = make(map[interface{}]*list.Element)
	m.dll = list.New()
	return m
}

// Get returns the value stored in the map for a key, or nil if no
// value is present.
// The ok result indicates whether value was found in the map.
func (m *Map) Get(key interface{}) (interface{}, bool) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	v, ok := m.mp[key]
	if !ok {
		return nil, false
	}

	me := v.Value.(mapElement)
	return me.value, ok
}

// Put sets the value for the given key.
// It will replace the value if the key already exists in the map
// even if the values are same.
func (m *Map) Put(key interface{}, val interface{}) {
	m.mu.Lock()
	defer m.mu.Unlock()

	if e, ok := m.mp[key]; !ok {
		m.mp[key] = m.dll.PushFront(mapElement{key: key, value: val})
	} else {
		e.Value = mapElement{key: key, value: val}
	}
}

// Delete deletes the value for a key.
// It returns a boolean indicating weather the key existed and it was deleted.
func (m *Map) Delete(key interface{}) bool {
	m.mu.Lock()
	defer m.mu.Unlock()

	e, ok := m.mp[key]
	if !ok {
		return false
	}

	m.dll.Remove(e)
	delete(m.mp, key)
	return true
}

// UnorderedRange will range over the map in an unordered sequence.
// This is same as ranging over a map using the "for range" syntax.
// Parameter func f should not call any method of the Map, eg Get, Put, Delete, UnorderedRange, OrderedRange etc
// It will cause a deadlock.
func (m *Map) UnorderedRange(f func(key interface{}, value interface{})) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	for k, v := range m.mp {
		f(k, v.Value.(mapElement).value)
	}
}

// OrderedRange will range over the map in ab ordered sequence.
// This is way faster than UnorderedRange. For a map containing 10_000_000 items
// UnorderedRange completes in ~1.7 seconds,
// OrderedRange completes in ~98 milli seconds.
// Parameter func f should not call any method of the Map, eg Get, Put, Delete, UnorderedRange, OrderedRange etc
// It will cause a deadlock.
func (m *Map) OrderedRange(f func(key interface{}, value interface{})) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	cur := m.dll.Back()
	for cur != nil {
		me := cur.Value.(mapElement)
		f(me.key, me.value)
		cur = cur.Prev()
	}
}

// Length will return the length of Map.
func (m *Map) Length() int {
	m.mu.RLock()
	defer m.mu.RUnlock()

	return m.dll.Len()
}

// GetOrPut will return the existing value if the key exists in the Map.
// If the key did not exist previously it will be added to the Map.
// updated will be true if the key existed previously
// otherwise it will be false if the key did not exist and was added to the Map.
func (m *Map) GetOrPut(key interface{}, value interface{}) (finalValue interface{}, updated bool) {
	m.mu.Lock()
	defer m.mu.Unlock()

	if e, exists := m.mp[key]; exists {
		e.Value = mapElement{key: key, value: value}
		return value, true
	} else {
		m.mp[key] = m.dll.PushFront(mapElement{key: key, value: value})
		return value, false
	}
}

// GetAndDelete will get the value saved against the given key.
// deleted will be true if the key existed previously
// otherwise it will be false.
func (m *Map) GetAndDelete(key interface{}) (value interface{}, deleted bool) {
	m.mu.Lock()
	defer m.mu.Unlock()

	if e, exists := m.mp[key]; exists {
		m.dll.Remove(e)
		delete(m.mp, key)
		return e.Value, true
	} else {
		return nil, false
	}
}
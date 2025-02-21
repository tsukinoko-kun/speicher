package speicher

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"strings"
	"sync"
	"time"
)

type (
	MemoryMap[T any] struct {
		data     map[string]T
		location string
		mut      sync.RWMutex

		timerMut     sync.Mutex
		saveTimer    *time.Timer
		maxSaveTimer *time.Timer
		saveOnce     *sync.Once
	}

	// Update the interface first
	Map[T any] interface {
		Get(key string) (T, bool)
		Find(func(T) bool) (value T, found bool)
		FindAll(func(T) bool) (values []T)
		Has(key string) bool
		Set(key string, value T)
		Overwrite(map[string]T)
		RangeKV() (<-chan MapRangeEl[T], func())
		RangeV() (<-chan T, func())
		Save() error
		Lock()
		Unlock()
		RLock()
		RUnlock()
	}

	MapRangeEl[T any] struct {
		Key   string
		Value T
	}
)

// Update RangeKV method
func (m *MemoryMap[T]) RangeKV() (<-chan MapRangeEl[T], func()) {
	ch := make(chan MapRangeEl[T])
	done := make(chan struct{})
	cancel := func() {
		select {
		case <-done:
			// channel already closed, no-op
			return
		default:
			if done != nil {
				close(done)
				done = nil
			}
		}
	}

	go func() {
		defer func() {
			if done != nil {
				close(done)
				done = nil
			}
			if ch != nil {
				close(ch)
				ch = nil
			}
		}()
		for key, value := range m.data {
			select {
			case <-done:
				return
			case ch <- MapRangeEl[T]{Key: key, Value: value}:
			}
		}
	}()

	return ch, cancel
}

// Update RangeV method
func (m *MemoryMap[T]) RangeV() (<-chan T, func()) {
	ch := make(chan T)
	done := make(chan struct{})
	cancel := func() {
		select {
		case <-done:
			// channel already closed, no-op
			return
		default:
			close(done)
		}
	}

	go func() {
		defer close(ch)
		defer close(done)
		for _, value := range m.data {
			select {
			case <-done:
				return
			case ch <- value:
			}
		}
	}()

	return ch, cancel
}

func (m *MemoryMap[T]) Get(key string) (value T, found bool) {
	value, found = m.data[key]
	return
}

func (m *MemoryMap[T]) Find(f func(T) bool) (value T, found bool) {
	for _, value = range m.data {
		if f(value) {
			found = true
			return
		}
	}
	found = false
	return
}

func (m *MemoryMap[T]) FindAll(f func(T) bool) (values []T) {
	for _, value := range m.data {
		if f(value) {
			values = append(values, value)
		}
	}
	return
}

func (m *MemoryMap[T]) Has(key string) bool {
	_, ok := m.data[key]
	return ok
}

func (m *MemoryMap[T]) Set(key string, value T) {
	m.data[key] = value
}

func (m *MemoryMap[T]) Overwrite(values map[string]T) {
	m.data = values
}

func (m *MemoryMap[T]) Lock() {
	m.mut.Lock()
}
func (m *MemoryMap[T]) Unlock() {
	m.mut.Unlock()
	notifyChanged(m)
}

func (m *MemoryMap[T]) RLock() {
	m.mut.RLock()
}
func (m *MemoryMap[T]) RUnlock() {
	m.mut.RUnlock()
}

func (m *MemoryMap[T]) Save() error {
	m.RLock()
	defer m.RUnlock()

	f, err := os.Create(m.location)
	if err != nil {
		return errors.Join(fmt.Errorf("failed to open file '%s'", m.location), err)
	}
	encoder := json.NewEncoder(f)
	if err := encoder.Encode(m.data); err != nil {
		return errors.Join(fmt.Errorf("failed to encode json file '%s'", m.location), err)
	}
	return nil
}

func LoadMap[T any](location string) (Map[T], error) {
	if strings.HasSuffix(location, ".json") {
		if m, err := loadMapFromJsonFile[T](location); err != nil {
			return nil, errors.Join(fmt.Errorf("unable to load map from file '%s'", location), err)
		} else {
			return m, nil
		}
	}
	return nil, fmt.Errorf("unable to find loader for '%s'", location)
}

func loadMapFromJsonFile[T any](location string) (Map[T], error) {
	f, err := os.Open(location)
	if err != nil {
		return nil, errors.Join(fmt.Errorf("failed to open file '%s'", location), err)
	}
	decoder := json.NewDecoder(f)
	m := &MemoryMap[T]{location: location}
	if err := decoder.Decode(&m.data); err != nil {
		return nil, errors.Join(fmt.Errorf("failed to decode json file '%s'", location), err)
	}
	return m, nil
}

func (m *MemoryMap[T]) getSaveTimer() *time.Timer {
	m.timerMut.Lock()
	defer m.timerMut.Unlock()
	return m.saveTimer
}

func (m *MemoryMap[T]) setSaveTimer(t *time.Timer) {
	m.timerMut.Lock()
	defer m.timerMut.Unlock()
	m.saveTimer = t
}

func (m *MemoryMap[T]) getMaxSaveTimer() *time.Timer {
	m.timerMut.Lock()
	defer m.timerMut.Unlock()
	return m.maxSaveTimer
}

func (m *MemoryMap[T]) setMaxSaveTimer(t *time.Timer) {
	m.timerMut.Lock()
	defer m.timerMut.Unlock()
	m.maxSaveTimer = t
}

func (m *MemoryMap[T]) getSaveOnce() *sync.Once {
	m.timerMut.Lock()
	defer m.timerMut.Unlock()
	return m.saveOnce
}

func (m *MemoryMap[T]) setSaveOnce(o *sync.Once) {
	m.timerMut.Lock()
	defer m.timerMut.Unlock()
	m.saveOnce = o
}

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
	MemoryList[T any] struct {
		data     []T
		location string
		mut      sync.RWMutex

		timerMut     sync.Mutex
		saveTimer    *time.Timer
		maxSaveTimer *time.Timer
		saveOnce     *sync.Once
	}

	List[T any] interface {
		Get(index int) (T, error)
		Find(func(T) bool) (value T, found bool)
		FindAll(func(T) bool) (values []T)
		Append(value T)
		AppendUnique(value T, equal func(a, b T) bool) bool
		Set(index int, value T) error
		Overwrite([]T)
		Len() int
		Range() (<-chan T, func())
		Save() error
		Lock()
		Unlock()
		RLock()
		RUnlock()
	}
)

func (l *MemoryList[T]) Get(index int) (value T, err error) {
	if index < 0 || index >= len(l.data) {
		return value, fmt.Errorf("index out of range")
	}
	return l.data[index], nil
}

func (l *MemoryList[T]) Append(value T) {
	l.data = append(l.data, value)
}

func (l *MemoryList[T]) AppendUnique(value T, equal func(a, b T) bool) bool {
	for _, x := range l.data {
		if equal(x, value) {
			return false
		}
	}
	l.data = append(l.data, value)
	return true
}

func (l *MemoryList[T]) Find(f func(T) bool) (value T, found bool) {
	for _, value = range l.data {
		if f(value) {
			found = true
			return
		}
	}
	found = false
	return
}

func (l *MemoryList[T]) FindAll(f func(T) bool) (values []T) {
	for _, value := range l.data {
		if f(value) {
			values = append(values, value)
		}
	}
	return
}

func (l *MemoryList[T]) Set(index int, value T) error {
	if index < 0 || index >= len(l.data) {
		return fmt.Errorf("index out of range")
	}
	l.data[index] = value
	return nil
}

func (l *MemoryList[T]) Overwrite(values []T) {
	l.data = values
}

func (l *MemoryList[T]) Len() int {
	return len(l.data)
}

func (l *MemoryList[T]) Range() (<-chan T, func()) {
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
		for _, value := range l.data {
			select {
			case <-done:
				return
			case ch <- value:
			}
		}
	}()

	return ch, cancel
}

func (l *MemoryList[T]) Lock() {
	l.mut.Lock()
}

func (l *MemoryList[T]) Unlock() {
	l.mut.Unlock()
	notifyChanged(l)
}

func (l *MemoryList[T]) RLock() {
	l.mut.RLock()
}

func (l *MemoryList[T]) RUnlock() {
	l.mut.RUnlock()
}

func (l *MemoryList[T]) Save() error {
	l.RLock()
	defer l.RUnlock()

	f, err := os.Create(l.location)
	if err != nil {
		return errors.Join(fmt.Errorf("failed to open file '%s'", l.location), err)
	}
	encoder := json.NewEncoder(f)
	if err := encoder.Encode(l.data); err != nil {
		return errors.Join(fmt.Errorf("failed to encode json file '%s'", l.location), err)
	}
	return nil
}

func LoadList[T any](location string) (List[T], error) {
	if strings.HasSuffix(location, ".json") {
		if l, err := loadListFromJsonFile[T](location); err != nil {
			return nil, errors.Join(fmt.Errorf("unable to load list from file '%s'", location), err)
		} else {
			return l, nil
		}
	}
	return nil, fmt.Errorf("unable to find loader for '%s'", location)
}

func loadListFromJsonFile[T any](location string) (List[T], error) {
	f, err := os.Open(location)
	if err != nil {
		return nil, errors.Join(fmt.Errorf("failed to open file '%s'", location), err)
	}
	decoder := json.NewDecoder(f)
	l := &MemoryList[T]{
		location: location,
		data:     make([]T, 0),
	}
	if err := decoder.Decode(&l.data); err != nil {
		return nil, errors.Join(fmt.Errorf("failed to decode json file '%s'", location), err)
	}
	return l, nil
}

func (l *MemoryList[T]) getSaveTimer() *time.Timer {
	l.timerMut.Lock()
	defer l.timerMut.Unlock()
	return l.saveTimer
}

func (l *MemoryList[T]) setSaveTimer(t *time.Timer) {
	l.timerMut.Lock()
	defer l.timerMut.Unlock()
	l.saveTimer = t
}

func (l *MemoryList[T]) getMaxSaveTimer() *time.Timer {
	l.timerMut.Lock()
	defer l.timerMut.Unlock()
	return l.maxSaveTimer
}

func (l *MemoryList[T]) setMaxSaveTimer(t *time.Timer) {
	l.timerMut.Lock()
	defer l.timerMut.Unlock()
	l.maxSaveTimer = t
}

func (l *MemoryList[T]) getSaveOnce() *sync.Once {
	l.timerMut.Lock()
	defer l.timerMut.Unlock()
	return l.saveOnce
}

func (l *MemoryList[T]) setSaveOnce(o *sync.Once) {
	l.timerMut.Lock()
	defer l.timerMut.Unlock()
	l.saveOnce = o
}

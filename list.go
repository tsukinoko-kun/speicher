package speicher

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"strings"
	"sync"
)

type (
	MemoryList[T any] struct {
		data     []T
		location string
		mut      sync.RWMutex
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
	_ = l.Save()
}

func (l *MemoryList[T]) RLock() {
	l.mut.RLock()
}

func (l *MemoryList[T]) RUnlock() {
	l.mut.RUnlock()
}

func (l *MemoryList[T]) Save() error {
	f, err := os.OpenFile(l.location, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0644)
	if err != nil {
		return errors.Join(fmt.Errorf("failed to open file %s", l.location), err)
	}
	encoder := json.NewEncoder(f)
	if err := encoder.Encode(l.data); err != nil {
		return errors.Join(fmt.Errorf("failed to encode json file %s", l.location), err)
	}
	return nil
}

func LoadList[T any](location string) (List[T], error) {
	if strings.HasSuffix(location, ".json") {
		return loadListFromJsonFile[T](location)
	}
	return nil, fmt.Errorf("unable to find loader for %s", location)
}

func loadListFromJsonFile[T any](location string) (List[T], error) {
	f, err := os.OpenFile(location, os.O_RDONLY, 0644)
	if err != nil {
		return nil, errors.Join(fmt.Errorf("failed to open file %s", location), err)
	}
	decoder := json.NewDecoder(f)
	l := &MemoryList[T]{
		location: location,
		data:     make([]T, 0),
	}
	if err := decoder.Decode(&l.data); err != nil {
		return nil, errors.Join(fmt.Errorf("failed to decode json file %s", location), err)
	}
	return l, nil
}

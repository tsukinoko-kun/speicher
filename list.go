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
	// memoryList is a List implementation that keeps all elements in memory.
	memoryList[T any] struct {
		data     []T
		location string
		mut      sync.RWMutex

		timerMut     sync.Mutex
		saveTimer    *time.Timer
		maxSaveTimer *time.Timer
		saveOnce     *sync.Once
	}

	// List data store
	List[T any] interface {
		// Get returns the value at a given index of the List and a bool that indicates whether the index exists or not.
		// If no element is found, the bool result will be false.
		Get(index int) (T, bool)

		// Find traverses the List and returns the first element that satisfies the provided predicate function.
		// If no element is found, the bool result will be false.
		Find(func(T) bool) (value T, found bool)

		// FindAll returns all elements in the List that satisfy the provided predicate function.
		// If no elements match, it returns an empty slice.
		FindAll(func(T) bool) (values []T)

		// Append adds the provided value to the end of the List.
		Append(value T)

		// AppendUnique adds the provided value to the List only if no existing element is equal to it,
		// based on the supplied equality function. It returns true if the value was added,
		// and false otherwise.
		AppendUnique(value T, equal func(a, b T) bool) bool

		// Set assigns the provided value to the element at the specified index.
		// If the index is out of bounds, it returns an error.
		Set(index int, value T) error

		// Overwrite replaces the entire List with the data provided in the slice.
		Overwrite([]T)

		// Len returns the number of elements currently in the List.
		Len() int

		// Range returns a read-only channel through which the elements of the List can be iterated.
		// It also returns a cancel function to stop the iteration process if needed.
		Range() (<-chan T, func())

		// Save persists the current state of the List to its underlying data store.
		// It returns an error if the operation fails.
		Save() error

		// Lock acquires an exclusive lock on the List to ensure thread-safe operations.
		Lock()

		// Unlock releases the exclusive lock previously acquired with Lock.
		Unlock()

		// RLock acquires a read lock on the List to allow concurrent read operations.
		RLock()

		// RUnlock releases the read lock acquired with RLock.
		RUnlock()
	}
)

func (l *memoryList[T]) Get(index int) (value T, found bool) {
	if index < 0 || index >= len(l.data) {
		value = l.data[index]
		found = true
	} else {
		found = false
	}
	return
}

func (l *memoryList[T]) Append(value T) {
	l.data = append(l.data, value)
}

func (l *memoryList[T]) AppendUnique(value T, equal func(a, b T) bool) bool {
	for _, x := range l.data {
		if equal(x, value) {
			return false
		}
	}
	l.data = append(l.data, value)
	return true
}

func (l *memoryList[T]) Find(f func(T) bool) (value T, found bool) {
	for _, value = range l.data {
		if f(value) {
			found = true
			return
		}
	}
	found = false
	return
}

func (l *memoryList[T]) FindAll(f func(T) bool) (values []T) {
	for _, value := range l.data {
		if f(value) {
			values = append(values, value)
		}
	}
	return
}

func (l *memoryList[T]) Set(index int, value T) error {
	if index < 0 || index >= len(l.data) {
		return fmt.Errorf("index out of range")
	}
	l.data[index] = value
	return nil
}

func (l *memoryList[T]) Overwrite(values []T) {
	l.data = values
}

func (l *memoryList[T]) Len() int {
	return len(l.data)
}

func (l *memoryList[T]) Range() (<-chan T, func()) {
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

func (l *memoryList[T]) Lock() {
	l.mut.Lock()
}

func (l *memoryList[T]) Unlock() {
	l.mut.Unlock()
	notifyChanged(l)
}

func (l *memoryList[T]) RLock() {
	l.mut.RLock()
}

func (l *memoryList[T]) RUnlock() {
	l.mut.RUnlock()
}

func (l *memoryList[T]) Save() error {
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
	l := &memoryList[T]{
		location: location,
		data:     make([]T, 0),
	}
	if err := decoder.Decode(&l.data); err != nil {
		return nil, errors.Join(fmt.Errorf("failed to decode json file '%s'", location), err)
	}
	return l, nil
}

func (l *memoryList[T]) getSaveTimer() *time.Timer {
	l.timerMut.Lock()
	defer l.timerMut.Unlock()
	return l.saveTimer
}

func (l *memoryList[T]) setSaveTimer(t *time.Timer) {
	l.timerMut.Lock()
	defer l.timerMut.Unlock()
	l.saveTimer = t
}

func (l *memoryList[T]) getMaxSaveTimer() *time.Timer {
	l.timerMut.Lock()
	defer l.timerMut.Unlock()
	return l.maxSaveTimer
}

func (l *memoryList[T]) setMaxSaveTimer(t *time.Timer) {
	l.timerMut.Lock()
	defer l.timerMut.Unlock()
	l.maxSaveTimer = t
}

func (l *memoryList[T]) getSaveOnce() *sync.Once {
	l.timerMut.Lock()
	defer l.timerMut.Unlock()
	return l.saveOnce
}

func (l *memoryList[T]) setSaveOnce(o *sync.Once) {
	l.timerMut.Lock()
	defer l.timerMut.Unlock()
	l.saveOnce = o
}

func (l *memoryList[T]) WriteE(f func(l *memoryList[T]) (any, error)) (any, error) {
	l.Lock()
	defer l.Unlock()
	return f(l)
}

func (l *memoryList[T]) Write(f func(l *memoryList[T]) any) any {
	l.Lock()
	defer l.Unlock()
	return f(l)
}

func (l *memoryList[T]) ReadE(f func(l *memoryList[T]) (any, error)) (any, error) {
	l.RLock()
	defer l.RUnlock()
	return f(l)
}

func (l *memoryList[T]) Read(f func(l *memoryList[T]) any) any {
	l.RLock()
	defer l.RUnlock()
	return f(l)
}

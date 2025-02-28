# Package speicher

## Functions

### Err

```go
func Err() <-chan error
```

### Read

Read locks the store before executing f. After f was executed, the store gets unlocked.


```go
func Read[S Store, R any](s Store, f func(s Store) R) R
```

### ReadE

Same as Read but f returns an error.


```go
func ReadE[S Store, R any](s Store, f func(s Store) (R, error)) (R, error)
```

### Write

Write locks the store before executing f. After f was executed, the store gets unlocked.


```go
func Write[S Store, R any](s Store, f func(s Store) R) R
```

### WriteE

Same as Write but f returns an error.


```go
func WriteE[S Store, R any](s Store, f func(s Store) (R, error)) (R, error)
```

### log

```go
func log(err error)
```

### notifyChanged

```go
func notifyChanged(s savable)
```

## Types

### List

List data store


#### Methods

##### Append

Append adds the provided value to the end of the List.


```go
func Append(value T)
```

##### AppendUnique

AppendUnique adds the provided value to the List only if no existing element is equal to it,
based on the supplied equality function. It returns true if the value was added,
and false otherwise.


```go
func AppendUnique(value T, equal func(a, b T) bool) bool
```

##### Find

Find traverses the List and returns the first element that satisfies the provided predicate function.
If no element is found, the bool result will be false.


```go
func Find(func(T) bool) (value T, found bool)
```

##### FindAll

FindAll returns all elements in the List that satisfy the provided predicate function.
If no elements match, it returns an empty slice.


```go
func FindAll(func(T) bool) (values []T)
```

##### Get

Get returns the value at a given index of the List and a bool that indicates whether the index exists or not.
If no element is found, the bool result will be false.


```go
func Get(index int) (T, bool)
```

##### Len

Len returns the number of elements currently in the List.


```go
func Len() int
```

##### Lock

Lock acquires an exclusive lock on the List to ensure thread-safe operations.


```go
func Lock()
```

##### Overwrite

Overwrite replaces the entire List with the data provided in the slice.


```go
func Overwrite([]T)
```

##### RLock

RLock acquires a read lock on the List to allow concurrent read operations.


```go
func RLock()
```

##### RUnlock

RUnlock releases the read lock acquired with RLock.


```go
func RUnlock()
```

##### Range

Range returns a read-only channel through which the elements of the List can be iterated.
It also returns a cancel function to stop the iteration process if needed.


```go
func Range() (<-chan T, func())
```

##### Save

Save persists the current state of the List to its underlying data store.
It returns an error if the operation fails.


```go
func Save() error
```

##### Set

Set assigns the provided value to the element at the specified index.
If the index is out of bounds, it returns an error.


```go
func Set(index int, value T) error
```

##### Unlock

Unlock releases the exclusive lock previously acquired with Lock.


```go
func Unlock()
```

### Map

Update the interface first


#### Methods

##### Find

```go
func Find(func(T) bool) (value T, found bool)
```

##### FindAll

```go
func FindAll(func(T) bool) (values []T)
```

##### Get

```go
func Get(key string) (T, bool)
```

##### Has

```go
func Has(key string) bool
```

##### Lock

```go
func Lock()
```

##### Overwrite

```go
func Overwrite(map[string]T)
```

##### RLock

```go
func RLock()
```

##### RUnlock

```go
func RUnlock()
```

##### RangeKV

```go
func RangeKV() (<-chan MapRangeEl[T], func())
```

##### RangeV

```go
func RangeV() (<-chan T, func())
```

##### Save

```go
func Save() error
```

##### Set

```go
func Set(key string, value T)
```

##### Unlock

```go
func Unlock()
```

### MapRangeEl

```go
type MapRangeEl[T any] struct {
	Key	string
	Value	T
}
```

### MemoryMap

#### Methods

##### Find

```go
func (m *MemoryMap[T]) Find(f func(T) bool) (value T, found bool)
```

##### FindAll

```go
func (m *MemoryMap[T]) FindAll(f func(T) bool) (values []T)
```

##### Get

```go
func (m *MemoryMap[T]) Get(key string) (value T, found bool)
```

##### Has

```go
func (m *MemoryMap[T]) Has(key string) bool
```

##### Lock

```go
func (m *MemoryMap[T]) Lock()
```

##### Overwrite

```go
func (m *MemoryMap[T]) Overwrite(values map[string]T)
```

##### RLock

```go
func (m *MemoryMap[T]) RLock()
```

##### RUnlock

```go
func (m *MemoryMap[T]) RUnlock()
```

##### RangeKV

Update RangeKV method


```go
func (m *MemoryMap[T]) RangeKV() (<-chan MapRangeEl[T], func())
```

##### RangeV

Update RangeV method


```go
func (m *MemoryMap[T]) RangeV() (<-chan T, func())
```

##### Read

```go
func (m *MemoryMap[T]) Read(f func(m *MemoryMap[T]) any) any
```

##### ReadE

```go
func (m *MemoryMap[T]) ReadE(f func(m *MemoryMap[T]) (any, error)) (any, error)
```

##### Save

```go
func (m *MemoryMap[T]) Save() error
```

##### Set

```go
func (m *MemoryMap[T]) Set(key string, value T)
```

##### Unlock

```go
func (m *MemoryMap[T]) Unlock()
```

##### Write

```go
func (m *MemoryMap[T]) Write(f func(m *MemoryMap[T]) any) any
```

##### WriteE

```go
func (m *MemoryMap[T]) WriteE(f func(m *MemoryMap[T]) (any, error)) (any, error)
```

### Store

```go
type Store interface {
	Lock()
	Unlock()
	RLock()
	RUnlock()
}
```

#### Methods

##### Lock

```go
func Lock()
```

##### RLock

```go
func RLock()
```

##### RUnlock

```go
func RUnlock()
```

##### Unlock

```go
func Unlock()
```


# Package speicher

## Functions

### Err

Err returns the error channel used when saving the data stores to disk.


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

List is a thread-safe list data store interface that provides basic
CRUD operations, predicate-based search, and iteration functionality.


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
Don't forget to use Unlock when you are done.


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
Don't forget to use RUnlock when you are done.


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

Map is a thread-safe key-value data store interface that provides basic
CRUD operations, predicate-based search, and iteration functionality.


#### Methods

##### Find

Find searches for an element that satisfies the given predicate.
It returns the found value and a boolean indicating if a match was found.


```go
func Find(func(T) bool) (value T, found bool)
```

##### FindAll

FindAll retrieves all elements that satisfy the given predicate.
It returns a slice containing all matching elements.


```go
func FindAll(func(T) bool) (values []T)
```

##### Get

Get retrieves an element associated with the given key.
It returns the value and a boolean indicating whether the key exists.


```go
func Get(key string) (T, bool)
```

##### Has

Has checks if an element with the given key exists in the data store.
It returns true if the key exists.


```go
func Has(key string) bool
```

##### Lock

Lock acquires the write lock for the data store to allow safe updates.
Don't forget to use Unlock when you are done.


```go
func Lock()
```

##### Overwrite

Overwrite replaces the entire data store with the provided map.


```go
func Overwrite(map[string]T)
```

##### RLock

RLock acquires the read lock for the data store to allow safe reading.
Don't forget to use RUnlock when you are done.


```go
func RLock()
```

##### RUnlock

RUnlock releases the read lock for the data store.


```go
func RUnlock()
```

##### RangeKV

RangeKV returns a read-only channel that emits key-value pair elements
(as MapRangeEl) from the data store, along with a cancellation function
to terminate the iteration when desired.


```go
func RangeKV() (<-chan MapRangeEl[T], func())
```

##### RangeV

RangeV returns a read-only channel that emits only the values stored in the
data store, along with a cancellation function to terminate the iteration.


```go
func RangeV() (<-chan T, func())
```

##### Save

Save persists the current state of the data store.
It returns an error if the save operation fails.


```go
func Save() error
```

##### Set

Set adds or updates the element associated with the given key.
If the key already exists, its value is overwritten.


```go
func Set(key string, value T)
```

##### Unlock

Unlock releases the write lock for the data store.


```go
func Unlock()
```

### MapRangeEl

MapRangeEl represents a key-value pair element emitted by the Map's RangeKV method.


```go
type MapRangeEl[T any] struct {
	Key	string
	Value	T
}
```

### Store

#### Methods

##### Lock

Lock acquires the write lock for the data store to allow safe updates.
Don't forget to use Unlock when you are done.


```go
func Lock()
```

##### RLock

RLock acquires the read lock for the data store to allow safe reading.
Don't forget to use RUnlock when you are done.


```go
func RLock()
```

##### RUnlock

RUnlock releases the read lock for the data store.


```go
func RUnlock()
```

##### Unlock

Unlock releases the write lock for the data store.


```go
func Unlock()
```


package speicher

type Store interface {
	Lock()
	Unlock()
	RLock()
	RUnlock()
}

// Same as Write but f returns an error.
func WriteE[S Store, R any](s Store, f func(s Store) (R, error)) (R, error) {
	s.Lock()
	defer s.Unlock()
	return f(s)
}

// Write locks the store before executing f. After f was executed, the store gets unlocked.
func Write[S Store, R any](s Store, f func(s Store) R) R {
	s.Lock()
	defer s.Unlock()
	return f(s)
}

// Same as Read but f returns an error.
func ReadE[S Store, R any](s Store, f func(s Store) (R, error)) (R, error) {
	s.RLock()
	defer s.RUnlock()
	return f(s)
}

// Read locks the store before executing f. After f was executed, the store gets unlocked.
func Read[S Store, R any](s Store, f func(s Store) R) R {
	s.RLock()
	defer s.RUnlock()
	return f(s)
}

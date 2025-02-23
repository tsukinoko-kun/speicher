package speicher

type Store interface {
	Lock()
	Unlock()
	RLock()
	RUnlock()
}

func WriteE[S Store, R any](s Store, f func(s Store) (R, error)) (R, error) {
	s.Lock()
	defer s.Unlock()
	return f(s)
}

func Write[S Store, R any](s Store, f func(s Store) R) R {
	s.Lock()
	defer s.Unlock()
	return f(s)
}

func ReadE[S Store, R any](s Store, f func(s Store) (R, error)) (R, error) {
	s.RLock()
	defer s.RUnlock()
	return f(s)
}

func Read[S Store, R any](s Store, f func(s Store) R) R {
	s.RLock()
	defer s.RUnlock()
	return f(s)
}

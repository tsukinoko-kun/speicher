package speicher

import (
	"fmt"
	"sync"
	"time"
)

type savable interface {
	Save() error
	getSaveTimer() *time.Timer
	setSaveTimer(*time.Timer)
	getMaxSaveTimer() *time.Timer
	setMaxSaveTimer(*time.Timer)
	getSaveOnce() *sync.Once
	setSaveOnce(*sync.Once)
}

var errChan chan error = nil

// Err returns the error channel used when saving the data stores to disk.
func Err() <-chan error {
	if errChan == nil {
		errChan = make(chan error)
	}
	return errChan
}

func log(err error) {
	if errChan != nil {
		errChan <- err
	} else {
		fmt.Println(err.Error())
	}
}

func notifyChanged(s savable) {
	const debounceDelay = 2 * time.Second
	const maxDelay = 10 * time.Second

	// Ensure that we have a "once" for the current burst.
	once := s.getSaveOnce()
	if once == nil {
		once = new(sync.Once)
		s.setSaveOnce(once)
	}

	// This callback, attached to both timers,
	// will execute Save only once.
	callback := func() {
		once.Do(func() {
			// Clear both timers and the once pointer so that
			// a new series can start on future notifyChanged calls.
			s.setSaveTimer(nil)
			s.setMaxSaveTimer(nil)
			s.setSaveOnce(nil)

			if err := s.Save(); err != nil {
				log(err)
			}
		})
	}

	// -- Debounce timer: resets on each notifyChanged call --
	t := s.getSaveTimer()
	if t != nil {
		// Reset to debounceDelay. If Reset returns false because the
		// timer has fired (or is firing) we create a new timer.
		if !t.Reset(debounceDelay) {
			newTimer := time.AfterFunc(debounceDelay, callback)
			s.setSaveTimer(newTimer)
		}
	} else {
		newTimer := time.AfterFunc(debounceDelay, callback)
		s.setSaveTimer(newTimer)
	}

	// -- Maximum delay timer: started only once for this burst --
	tmax := s.getMaxSaveTimer()
	if tmax == nil {
		newMaxTimer := time.AfterFunc(maxDelay, callback)
		s.setMaxSaveTimer(newMaxTimer)
	}
}

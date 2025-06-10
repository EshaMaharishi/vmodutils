package vmodutils

import (
	"fmt"
	"sync"
)

type StringState struct {
	lock   sync.Mutex
	states []string
}

func (s *StringState) Push(str string) {
	s.lock.Lock()
	s.states = append(s.states, str)
	s.lock.Unlock()
}

func (s *StringState) Pop() {
	s.lock.Lock()
	if len(s.states) > 0 {
		s.states = s.states[0 : len(s.states)-1]
	}
	s.lock.Unlock()
}

func (s *StringState) String() string {
	var ss string
	s.lock.Lock()
	for i, x := range s.states {
		if i > 0 {
			ss += ","
		}
		ss += x
	}
	s.lock.Unlock()
	return ss
}

func (s *StringState) CheckEmpty() error {
	var err error
	s.lock.Lock()
	if len(s.states) > 0 {
		err = fmt.Errorf("states should be empty, is: %v", s.states)
		s.states = nil
	}
	s.lock.Unlock()
	return err
}

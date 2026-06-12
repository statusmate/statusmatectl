package tui

// StackListener is notified on stack mutations.
type StackListener interface {
	StackPushed(name string)
	StackPopped(name string)
	StackTop(name string)
}

// Stack manages an ordered navigation history of page names.
type Stack struct {
	items     []string
	listeners []StackListener
}

func newStack() *Stack { return &Stack{} }

func (s *Stack) addListener(l StackListener) {
	s.listeners = append(s.listeners, l)
}

func (s *Stack) Push(name string) {
	s.items = append(s.items, name)
	for _, l := range s.listeners {
		l.StackPushed(name)
	}
	for _, l := range s.listeners {
		l.StackTop(name)
	}
}

func (s *Stack) Pop() {
	if len(s.items) == 0 {
		return
	}
	top := s.items[len(s.items)-1]
	s.items = s.items[:len(s.items)-1]
	for _, l := range s.listeners {
		l.StackPopped(top)
	}
	if len(s.items) > 0 {
		newTop := s.items[len(s.items)-1]
		for _, l := range s.listeners {
			l.StackTop(newTop)
		}
	}
}

// Reset replaces the entire stack with a single entry.
func (s *Stack) Reset(name string) {
	s.items = []string{name}
	for _, l := range s.listeners {
		l.StackTop(name)
	}
}

func (s *Stack) Peek() []string {
	cp := make([]string, len(s.items))
	copy(cp, s.items)
	return cp
}

func (s *Stack) Top() string {
	if len(s.items) == 0 {
		return ""
	}
	return s.items[len(s.items)-1]
}

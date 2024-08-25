package ui

type stack[T any] struct {
	items []T
}

func (s *stack[T]) Push(el ...T) {
	s.items = append(s.items, el...)
}
func (s *stack[_]) Len() int {
	return len(s.items)
}
func (s *stack[T]) Pop() T {
	el := s.items[len(s.items)-1]
	s.items = s.items[:len(s.items)-1]
	return el
}

func newStack[T any](els ...T) stack[T] {
	s := stack[T]{}
	s.Push(els...)
	return s
}

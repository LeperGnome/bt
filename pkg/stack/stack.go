package stack

type Stack[T any] struct {
	items []T
}

func (s *Stack[T]) Push(el ...T) {
	s.items = append(s.items, el...)
}
func (s *Stack[_]) Len() int {
	return len(s.items)
}
func (s *Stack[T]) Pop() T {
	el := s.items[len(s.items)-1]
	s.items = s.items[:len(s.items)-1]
	return el
}

func NewStack[T any](els ...T) Stack[T] {
	s := Stack[T]{}
	s.Push(els...)
	return s
}

package transform

type StringStack struct {
	arr []string
}

func (s *StringStack) Push(node string) {
	s.arr = append(s.arr, node)
}

func (s *StringStack) Pop() string {
	last := s.arr[len(s.arr)-1]
	s.arr = s.arr[:len(s.arr)-1]

	return last
}

func (s *StringStack) Peek() string {
	return s.arr[len(s.arr)-1]
}

func (s *StringStack) Len() int {
	return len(s.arr)
}

package main

type StrSet struct  {
	set map[string]bool
}

func createSet() *StrSet {
	return &StrSet{
		set: make(map[string]bool),
	}
}

func (s *StrSet) add(str string) {
	s.set[str] = true
}

func (s *StrSet) addArr(arr []string) {
	for _, str := range arr {
		s.set[str] = true
	}
}

func (s *StrSet) addSet(other *StrSet) {
	for str := range other.set {
		s.set[str] = true
	}
}

func (s *StrSet) forEach(f func(str string)) {
	for str := range s.set {
		f(str)
	}
}

func (s *StrSet) contains(str string) bool {
	_, ok := s.set[str]
	return ok
}

func (s *StrSet) remove(str string) {
	delete(s.set, str)
}

func (s *StrSet) size() int {
	return len(s.set)
}

func (s *StrSet) equal(other *StrSet) bool {
	if s.size() != other.size() {
		return false
	}
	for str := range s.set {
		if !other.contains(str) {
			return false
		}
	}
	return true
}

func (s *StrSet) string() string {
	result := "{"
	for key := range s.set {
		result += key + ", "
	}
	result += "}"
	return result
}
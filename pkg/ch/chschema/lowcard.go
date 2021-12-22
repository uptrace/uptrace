package chschema

type lowCard struct {
	slice sliceMap
	dict  map[string]int
}

func (lc *lowCard) Add(word string) int {
	if i, ok := lc.dict[word]; ok {
		return i
	}

	if lc.dict == nil {
		lc.dict = make(map[string]int)
	}

	i := lc.slice.Add(word)
	lc.dict[word] = i

	return i
}

func (lc *lowCard) Dict() []string {
	return lc.slice.Slice()
}

//------------------------------------------------------------------------------

type sliceMap struct {
	ss []string
}

func (m sliceMap) Len() int {
	return len(m.ss)
}

func (m sliceMap) Get(word string) (int, bool) {
	for i, s := range m.ss {
		if s == word {
			return i, true
		}
	}
	return 0, false
}

func (m *sliceMap) Add(word string) int {
	m.ss = append(m.ss, word)
	return len(m.ss) - 1
}

func (m sliceMap) Slice() []string {
	return m.ss
}

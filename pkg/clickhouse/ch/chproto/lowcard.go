package chproto

type LowCard struct {
	keys         []int
	hasEmptyWord bool
	slice        []string
	dict         map[string]int
}

func NewLowCard() *LowCard { return new(LowCard) }
func (lc *LowCard) Reset() {
	lc.hasEmptyWord = false
	lc.slice = lc.slice[:0]
	clear(lc.dict)
}
func (lc *LowCard) MakeKeys(n int) []int {
	if cap(lc.keys) >= n {
		return lc.keys[:n]
	}
	lc.keys = make([]int, n)
	return lc.keys
}
func (lc *LowCard) Add(word string) int {
	if len(lc.slice) == 0 {
		if word == "" {
			lc.hasEmptyWord = true
			return 0
		}
		if lc.slice == nil {
			lc.slice = make([]string, 0, 16)
			lc.dict = make(map[string]int, 16)
		}
		if lc.hasEmptyWord {
			lc.slice = append(lc.slice, "")
			lc.dict[""] = 0
		}
	}
	if i, ok := lc.dict[word]; ok {
		return i
	}
	i := len(lc.slice)
	lc.slice = append(lc.slice, word)
	lc.dict[word] = i
	return i
}
func (lc *LowCard) Dict() []string {
	if len(lc.slice) == 0 && lc.hasEmptyWord {
		return []string{""}
	}
	return lc.slice
}

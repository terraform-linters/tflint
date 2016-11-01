package issue

type Issue struct {
	Type    string
	Message string
	Line    int
	File    string
}

type Issues []*Issue

func (is Issues) Len() int      { return len(is) }
func (is Issues) Swap(i, j int) { is[i], is[j] = is[j], is[i] }

type ByFile struct {
	Issues
}

func (b ByFile) Less(i, j int) bool { return b.Issues[i].File < b.Issues[j].File }

type ByLine struct {
	Issues
}

func (b ByLine) Less(i, j int) bool { return b.Issues[i].Line < b.Issues[j].Line }

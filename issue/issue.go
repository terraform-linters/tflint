package issue

import "fmt"

type Issue struct {
	Detector string `json:"detector"`
	Type     string `json:"type"`
	Message  string `json:"message"`
	Line     int    `json:"line"`
	File     string `json:"file"`
	Link     string `json:"link"`
}

const ERROR = "ERROR"
const WARNING = "WARNING"
const NOTICE = "NOTICE"

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

type ByFileLine struct {
	Issues
}

func (b ByFileLine) Less(i, j int) bool {
	return fmt.Sprintf("%s:%b", b.Issues[i].File, b.Issues[i].Line) < fmt.Sprintf("%s:%b", b.Issues[j].File, b.Issues[j].Line)
}

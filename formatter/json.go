package formatter

import (
	"encoding/json"
	"fmt"

	"github.com/wata727/tflint/tflint"
)

func (f *Formatter) jsonPrint(issues tflint.Issues, tferr *tflint.Error) {
	result, err := json.Marshal(issues)
	if err != nil {
		fmt.Fprint(f.Stderr, err)
	}
	fmt.Fprint(f.Stdout, string(result))
}

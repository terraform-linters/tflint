package main

import (
	"fmt"
	"os"
	"path"
	"runtime"
	"strings"

	colorable "github.com/mattn/go-colorable"
	"github.com/terraform-linters/tflint/cmd"
)

func main() {
	cli := cmd.NewCLI(colorable.NewColorable(os.Stdout), colorable.NewColorable(os.Stderr))

	defer func() {
		if r := recover(); r != nil {
			fmt.Fprintf(os.Stderr, "Panic: %v\n", r)
			for depth := 0; ; depth++ {
				pc, src, line, ok := runtime.Caller(depth)
				if !ok {
					break
				}
				fmt.Fprintf(os.Stderr, " -> %d: %s: %s(%d)\n", depth, runtime.FuncForPC(pc).Name(), strings.Replace(src, path.Dir(src), "", 1), line)
			}
			fmt.Fprintln(os.Stderr, `
TFLint crashed... :(
Please attach an output log, describe the situation and version that occurred and post an issue to https://github.com/terraform-linters/tflint/issues`)
			os.Exit(cmd.ExitCodeError)
		}
	}()

	os.Exit(cli.Run(os.Args))
}

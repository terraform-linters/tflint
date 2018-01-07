package main

import (
	"fmt"
	"os"
	"path"
	"runtime"
	"strings"
)

func main() {
	cli := &CLI{outStream: os.Stdout, errStream: os.Stderr, testMode: false}
	defer func() {
		if r := recover(); r != nil {
			fmt.Fprintf(cli.errStream, "Panic: %v\n", r)
			for depth := 0; ; depth++ {
				pc, src, line, ok := runtime.Caller(depth)
				if !ok {
					break
				}
				fmt.Fprintf(cli.errStream, " -> %d: %s: %s(%d)\n", depth, runtime.FuncForPC(pc).Name(), strings.Replace(src, path.Dir(src), "", 1), line)
			}
			fmt.Fprintln(cli.errStream, `
TFLint crashed... :(
Please attach an output log, describe the situation and version that occurred and post an issue to https://github.com/wata727/tflint/issues
`)
			os.Exit(ExitCodeError)
		}
	}()
	os.Exit(cli.Run(os.Args))
}

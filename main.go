package main

import (
	"fmt"
	"log"
	"os"
	"path"
	"runtime"
	"strings"

	"github.com/hashicorp/logutils"
	colorable "github.com/mattn/go-colorable"
)

func main() {
	cli := &CLI{
		outStream: colorable.NewColorable(os.Stdout),
		errStream: colorable.NewColorable(os.Stderr),
		testMode:  false,
	}
	filter := &logutils.LevelFilter{
		Levels:   []logutils.LogLevel{"DEBUG", "INFO", "WARN", "ERROR"},
		MinLevel: logutils.LogLevel(strings.ToUpper(os.Getenv("TFLINT_LOG"))),
		Writer:   cli.errStream,
	}
	log.SetOutput(filter)

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
Please attach an output log, describe the situation and version that occurred and post an issue to https://github.com/wata727/tflint/issues`)
			os.Exit(ExitCodeError)
		}
	}()

	os.Exit(cli.Run(os.Args))
}

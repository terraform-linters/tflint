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
	"github.com/wata727/tflint/cmd"
	"github.com/wata727/tflint/issue"
	"github.com/wata727/tflint/plugin"
	"github.com/wata727/tflint/plugin/discovery"
)

func main() {
	cli := cmd.NewCLI(colorable.NewColorable(os.Stdout), colorable.NewColorable(os.Stderr))
	filter := &logutils.LevelFilter{
		Levels:   []logutils.LogLevel{"DEBUG", "INFO", "WARN", "ERROR"},
		MinLevel: logutils.LogLevel(strings.ToUpper(os.Getenv("TFLINT_LOG"))),
		Writer:   os.Stderr,
	}
	log.SetOutput(filter)
	log.SetFlags(log.Ltime | log.Lshortfile)

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
Please attach an output log, describe the situation and version that occurred and post an issue to https://github.com/wata727/tflint/issues`)
			os.Exit(cmd.ExitCodeError)
		}
	}()

	pluginSearch := discovery.PluginSearch{}
	pluginSearch.Find()

	var pluginRuleViolations []*issue.Issue
	for _, foundPlugin := range pluginSearch.Plugins {
		client := plugin.Client(foundPlugin)

		defer client.Kill()

		// Connect via RPC
		rpcClient, err := client.Client()
		if err != nil {
			log.Fatal(err)
		}

		// Request the plugin
		raw, err := rpcClient.Dispense("rules")
		if err != nil {
			log.Fatal(err)
		}

		tflintPlugin := raw.(plugin.RuleCollection)

		pluginRuleViolations = append(pluginRuleViolations, tflintPlugin.Process(os.Args)...)
		if err != nil {
			panic(err)
		}
	}

	cli.SanityCheck(os.Args)
	cli.ProcessRules()
	cli.ReportViolations(pluginRuleViolations)

	os.Exit(cli.ExitCode)
}

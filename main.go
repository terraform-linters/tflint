package main

import (
	"fmt"
	"log"
	"os"
	"os/exec"
	"path"
	"runtime"
	"strings"

	"github.com/hashicorp/go-hclog"
	"github.com/hashicorp/go-plugin"
	"github.com/hashicorp/logutils"
	colorable "github.com/mattn/go-colorable"
	"github.com/wata727/tflint/cmd"
	"github.com/wata727/tflint/plugin/shared"
)

// handshakeConfigs are used to just do a basic handshake between
// a plugin and host. If the handshake fails, a user friendly error is shown.
// This prevents users from executing bad plugins or executing a plugin
// directory. It is a UX feature, not a security feature.
var handshakeConfig = plugin.HandshakeConfig{
	ProtocolVersion:  1,
	MagicCookieKey:   "BASIC_PLUGIN",
	MagicCookieValue: "hello",
}

// pluginMap is the map of plugins we can dispense.
var pluginMap = map[string]plugin.Plugin{
	"customRulesPlugin": &rulesplugin.RuleCollectionPlugin{},
}

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

	logger := hclog.New(&hclog.LoggerOptions{
		Name:   "tflint",
		Output: os.Stdout,
		Level:  hclog.Info,
	})

	// We're a host! Start by launching the plugin process.
	client := plugin.NewClient(&plugin.ClientConfig{
		HandshakeConfig: handshakeConfig,
		Plugins:         pluginMap,
		Cmd:             exec.Command("/home/howdoicomputer/go/src/github.com/howdoicomputer/customrules/customrules"),
		Logger:          logger,
	})
	defer client.Kill()

	// Connect via RPC
	rpcClient, err := client.Client()
	if err != nil {
		log.Fatal(err)
	}

	// Request the plugin
	raw, err := rpcClient.Dispense("customRulesPlugin")
	if err != nil {
		log.Fatal(err)
	}

	rulesPlugin := raw.(rulesplugin.RuleCollection)

	pluginRuleViolations := rulesPlugin.Process(os.Args)
	if err != nil {
		panic(err)
	}

	cli.SanityCheck(os.Args)
	cli.ProcessRules()
	cli.ReportViolations(pluginRuleViolations)

	os.Exit(cli.ExitCode)
}

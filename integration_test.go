package main

import (
	"bytes"
	"encoding/json"
	"io/ioutil"
	"log"
	"os"
	"runtime"
	"sort"
	"strings"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/wata727/tflint/cmd"
	"github.com/wata727/tflint/issue"
)

func TestMain(m *testing.M) {
	log.SetOutput(ioutil.Discard)
	os.Exit(m.Run())
}

func TestIntegration(t *testing.T) {
	cases := []struct {
		Name    string
		Command string
		Dir     string
	}{
		{
			Name:    "basic",
			Command: "./tflint --format json",
			Dir:     "basic",
		},
		{
			Name:    "override",
			Command: "./tflint --format json",
			Dir:     "override",
		},
		{
			Name:    "variables",
			Command: "./tflint --format json --var-file variables.tfvars",
			Dir:     "variables",
		},
		{
			Name:    "module",
			Command: "./tflint --format json",
			Dir:     "module",
		},
	}

	dir, _ := os.Getwd()
	defer os.Chdir(dir)

	for _, tc := range cases {
		testDir := dir + "/integration/" + tc.Dir
		os.Chdir(testDir)

		outStream, errStream := new(bytes.Buffer), new(bytes.Buffer)
		cli := cmd.NewCLI(outStream, errStream)

		args := strings.Split(tc.Command, " ")

		err := cli.SanityCheck(args)
		if err == nil {
			cli.Run()
		}

		var b []byte
		var err error
		if runtime.GOOS == "windows" && IsWindowsResultExist() {
			b, err = ioutil.ReadFile("result_windows.json")
		} else {
			b, err = ioutil.ReadFile("result.json")
		}
		if err != nil {
			t.Fatal(err)
		}

		var expectedIssues []*issue.Issue
		if err := json.Unmarshal(b, &expectedIssues); err != nil {
			t.Fatal(err)
		}
		sort.Sort(issue.ByFileLine{Issues: issue.Issues(expectedIssues)})

		var resultIssues []*issue.Issue
		if err := json.Unmarshal(outStream.Bytes(), &resultIssues); err != nil {
			t.Fatal(err)
		}
		sort.Sort(issue.ByFileLine{Issues: issue.Issues(resultIssues)})

		if !cmp.Equal(resultIssues, expectedIssues) {
			t.Fatalf("Failed `%s` test: diff=%s", tc.Name, cmp.Diff(expectedIssues, resultIssues))
		}
	}
}

func IsWindowsResultExist() bool {
	_, err := os.Stat("result_windows.json")
	return !os.IsNotExist(err)
}

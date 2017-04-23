package main

import (
	"bytes"
	"encoding/json"
	"io/ioutil"
	"os"
	"reflect"
	"sort"
	"strings"
	"testing"

	"github.com/k0kubun/pp"
	"github.com/wata727/tflint/issue"
)

func TestIntegration(t *testing.T) {
	cases := []struct {
		Name    string
		Command string
		Dir     string
	}{
		{
			Name:    "general",
			Command: "./tflint --format json",
			Dir:     "general",
		},
	}

	for _, tc := range cases {
		dir, _ := os.Getwd()
		testDir := dir + "/integration/" + tc.Dir
		os.Chdir(testDir)

		outStream, errStream := new(bytes.Buffer), new(bytes.Buffer)
		cli := &CLI{
			outStream: outStream,
			errStream: errStream,
		}
		args := strings.Split(tc.Command, " ")
		cli.Run(args)

		b, _ := ioutil.ReadFile("result.json")
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

		if !reflect.DeepEqual(resultIssues, expectedIssues) {
			t.Fatalf("\nBad: %s\nExpected: %s\n\ntestcase: %s", pp.Sprint(resultIssues), pp.Sprint(expectedIssues), tc.Name)
		}
	}
}

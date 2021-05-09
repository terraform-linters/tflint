package plugin

import (
	"errors"
	"strings"
	"testing"
)

func Test_NewChecksummer(t *testing.T) {
	cases := []struct {
		Name     string
		Input    string
		Expected error
	}{
		{
			Name: "valid checksums",
			Input: `3a61fff3689f27c89bce22893219919c629d2e10b96e7eadd5fef9f0e90bb353  tflint-ruleset-aws_darwin_amd64.zip
482419fdeed00692304e59558b5b0d915d4727868b88a5adbbbb76f5ed1b537a  tflint-ruleset-aws_linux_amd64.zip
db4eed4c0abcfb0b851da5bbfe8d0c71e1c2b6afe4fd627638a462c655045902  tflint-ruleset-aws_windows_amd64.zip
`,
			Expected: nil,
		},
		{
			Name: "No enough fields",
			Input: `3a61fff3689f27c89bce22893219919c629d2e10b96e7eadd5fef9f0e90bb353
482419fdeed00692304e59558b5b0d915d4727868b88a5adbbbb76f5ed1b537a
db4eed4c0abcfb0b851da5bbfe8d0c71e1c2b6afe4fd627638a462c655045902
`,
			Expected: errors.New("record on line 1: wrong number of fields: expected=2, actual=1"),
		},
		{
			Name: "Too many fields",
			Input: `3a61fff3689f27c89bce22893219919c629d2e10b96e7eadd5fef9f0e90bb353  tflint-ruleset-aws_darwin_amd64.zip  valid
482419fdeed00692304e59558b5b0d915d4727868b88a5adbbbb76f5ed1b537a  tflint-ruleset-aws_linux_amd64.zip  valid
db4eed4c0abcfb0b851da5bbfe8d0c71e1c2b6afe4fd627638a462c655045902  tflint-ruleset-aws_windows_amd64.zip  valid
`,
			Expected: errors.New("record on line 1: wrong number of fields: expected=2, actual=3"),
		},
	}

	for _, tc := range cases {
		_, err := NewChecksummer(strings.NewReader(tc.Input))

		if err != nil {
			if tc.Expected == nil {
				t.Fatalf("Failed `%s`: Unexpected error `%s`", tc.Name, err)
			}
			if err.Error() != tc.Expected.Error() {
				t.Fatalf("Failed `%s`: expected=%s, actual=%s", tc.Name, tc.Expected, err)
			}
		} else {
			if tc.Expected != nil {
				t.Fatalf("Failed `%s`: expected=%s, actual=no errors", tc.Name, tc.Expected)
			}
		}
	}
}

func Test_Checksummer_Verify(t *testing.T) {
	cases := []struct {
		Name      string
		Checksums string
		Input     string
		Filename  string
		Error     error
	}{
		{
			Name:      "valid checksums",
			Checksums: "f6f24a11d7cbbbc6d9440aca2eba0f6498755ca90adea14c5e233bf4c04bd928  text.txt",
			Input:     "foo bar baz\n",
			Filename:  "text.txt",
			Error:     nil,
		},
		{
			Name:      "checksum mismatched",
			Checksums: "f6f24a11d7cbbbc6d9440aca2eba0f6498755ca90adea14c5e233bf4c04bd928  text.txt",
			Input:     "baz baz foo\n",
			Filename:  "text.txt",
			Error:     errors.New("Failed to match checksums: expected=f6f24a11d7cbbbc6d9440aca2eba0f6498755ca90adea14c5e233bf4c04bd928, actual=29f7c2ecd78d7d4583604a9b8ac9c75eb5cb7ef9ee792bdd241dea563baad1b1"),
		},
		{
			Name:      "file not found in checksums",
			Checksums: "f6f24a11d7cbbbc6d9440aca2eba0f6498755ca90adea14c5e233bf4c04bd928  text.txt",
			Input:     "foo bar baz\n",
			Filename:  "text.png",
			Error:     errors.New("Failed to match checksums: expected=, actual=f6f24a11d7cbbbc6d9440aca2eba0f6498755ca90adea14c5e233bf4c04bd928"),
		},
	}

	for _, tc := range cases {
		checksumReader := strings.NewReader(tc.Checksums)
		inputReader := strings.NewReader(tc.Input)

		checksummer, err := NewChecksummer(checksumReader)
		if err != nil {
			t.Fatalf("Failed `%s`: Unexpected error `%s`", tc.Name, err)
		}

		err = checksummer.Verify(tc.Filename, inputReader)
		if err != nil {
			if tc.Error == nil {
				t.Fatalf("Failed `%s`: Unexpected error `%s`", tc.Name, err)
			}
			if err.Error() != tc.Error.Error() {
				t.Fatalf("Failed `%s`: expected=%s, actual=%s", tc.Name, tc.Error, err)
			}
		} else {
			if tc.Error != nil {
				t.Fatalf("Failed `%s`: expected=%s, actual=no errors", tc.Name, tc.Error)
			}
		}
	}
}

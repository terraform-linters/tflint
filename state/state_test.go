package state

import (
	"encoding/json"
	"testing"
)

func TestExists(t *testing.T) {
	type Input struct {
		Type string
		ID   string
	}

	cases := []struct {
		Name   string
		State  string
		Input  Input
		Result bool
	}{
		{
			Name: "exist in state",
			State: `
{
    "modules": [
        {
            "resources": {
                "aws_security_group.test": {
                    "type": "aws_security_group",
                    "depends_on": [],
                    "primary": {
                        "id": "sg-1234abcd",
                        "attributes": {
                            "id": "sg-1234abcd",
                            "name": "test",
                            "owner_id": "123456789",
                            "vpc_id": "vpc-1234abcd"
                        }
                    }
                }
            }
        }
    ]
}
`,
			Input: Input{
				Type: "aws_security_group",
				ID:   "test",
			},
			Result: true,
		},
		{
			Name: "not found in state",
			State: `
{
    "modules": [
        {
            "resources": {
                "aws_security_group.default": {
                    "type": "aws_security_group",
                    "depends_on": [],
                    "primary": {
                        "id": "sg-1234abcd",
                        "attributes": {
                            "id": "sg-1234abcd",
                            "name": "default",
                            "owner_id": "123456789",
                            "vpc_id": "vpc-1234abcd"
                        }
                    }
                }
            }
        }
    ]
}
`,
			Input: Input{
				Type: "aws_security_group",
				ID:   "test",
			},
			Result: false,
		},
	}

	for _, tc := range cases {
		tfstate := &TFState{}
		json.Unmarshal([]byte(tc.State), tfstate)

		result := tfstate.Exists(tc.Input.Type, tc.Input.ID)
		if result != tc.Result {
			t.Fatalf("\nBad: %t\nExpected: %t\n\ntestcase: %s", result, tc.Result, tc.Name)
		}
	}
}

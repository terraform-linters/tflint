package formatter

import (
	"bytes"
	"errors"
	"fmt"
	"path/filepath"
	"testing"

	"github.com/google/go-cmp/cmp"
	hcl "github.com/hashicorp/hcl/v2"
	"github.com/terraform-linters/tflint/tflint"
	"github.com/xeipuuv/gojsonschema"
)

func Test_sarifPrint(t *testing.T) {
	cases := []struct {
		Name   string
		Issues tflint.Issues
		Error  error
		Stdout string
	}{
		{
			Name:   "no issues",
			Issues: tflint.Issues{},
			Stdout: `{
  "version": "2.1.0",
  "$schema": "https://json.schemastore.org/sarif-2.1.0-rtm.5.json",
  "runs": [
    {
      "tool": {
        "driver": {
          "name": "tflint",
          "version": "0.45.0",
          "informationUri": "https://github.com/terraform-linters/tflint"
        }
      },
      "results": []
    },
    {
      "tool": {
        "driver": {
          "name": "tflint-errors",
          "version": "0.45.0",
          "informationUri": "https://github.com/terraform-linters/tflint"
        }
      },
      "results": []
    }
  ]
}`,
		},
		{
			Name: "issues",
			Issues: tflint.Issues{
				{
					Rule:    &testRule{},
					Message: "test",
					Range: hcl.Range{
						Filename: "test.tf",
						Start:    hcl.Pos{Line: 1, Column: 1, Byte: 0},
						End:      hcl.Pos{Line: 1, Column: 4, Byte: 3},
					},
				},
			},
			Stdout: `{
  "version": "2.1.0",
  "$schema": "https://json.schemastore.org/sarif-2.1.0-rtm.5.json",
  "runs": [
    {
      "tool": {
        "driver": {
          "name": "tflint",
          "version": "0.45.0",
          "informationUri": "https://github.com/terraform-linters/tflint",
          "rules": [
            {
              "id": "test_rule",
              "shortDescription": {
                "text": ""
              },
              "helpUri": "https://github.com"
            }
          ]
        }
      },
      "results": [
        {
          "ruleId": "test_rule",
          "level": "error",
          "message": {
            "text": "test"
          },
          "locations": [
            {
              "physicalLocation": {
                "artifactLocation": {
                  "uri": "test.tf"
                },
                "region": {
                  "startLine": 1,
                  "startColumn": 1,
                  "endLine": 1,
                  "endColumn": 4
                }
              }
            }
          ]
        }
      ]
    },
    {
      "tool": {
        "driver": {
          "name": "tflint-errors",
          "version": "0.45.0",
          "informationUri": "https://github.com/terraform-linters/tflint"
        }
      },
      "results": []
    }
  ]
}`,
		},
		{
			Name: "issues in directories",
			Issues: tflint.Issues{
				{
					Rule:    &testRule{},
					Message: "test",
					Range: hcl.Range{
						Filename: filepath.Join("test", "main.tf"),
						Start:    hcl.Pos{Line: 1, Column: 1, Byte: 0},
						End:      hcl.Pos{Line: 1, Column: 4, Byte: 3},
					},
				},
			},
			Stdout: `{
  "version": "2.1.0",
  "$schema": "https://json.schemastore.org/sarif-2.1.0-rtm.5.json",
  "runs": [
    {
      "tool": {
        "driver": {
          "name": "tflint",
          "version": "0.45.0",
          "informationUri": "https://github.com/terraform-linters/tflint",
          "rules": [
            {
              "id": "test_rule",
              "shortDescription": {
                "text": ""
              },
              "helpUri": "https://github.com"
            }
          ]
        }
      },
      "results": [
        {
          "ruleId": "test_rule",
          "level": "error",
          "message": {
            "text": "test"
          },
          "locations": [
            {
              "physicalLocation": {
                "artifactLocation": {
                  "uri": "test/main.tf"
                },
                "region": {
                  "startLine": 1,
                  "startColumn": 1,
                  "endLine": 1,
                  "endColumn": 4
                }
              }
            }
          ]
        }
      ]
    },
    {
      "tool": {
        "driver": {
          "name": "tflint-errors",
          "version": "0.45.0",
          "informationUri": "https://github.com/terraform-linters/tflint"
        }
      },
      "results": []
    }
  ]
}`,
		},
		{
			Name: "Issues with missing source positions",
			Issues: tflint.Issues{
				{
					Rule:    &testRule{},
					Message: "test",
					Range: hcl.Range{
						Filename: "test.tf",
					},
				},
			},
			Error: fmt.Errorf("Failed to work; %w", errors.New("I don't feel like working")),
			Stdout: `{
  "version": "2.1.0",
  "$schema": "https://json.schemastore.org/sarif-2.1.0-rtm.5.json",
  "runs": [
    {
      "tool": {
        "driver": {
          "name": "tflint",
          "version": "0.45.0",
          "informationUri": "https://github.com/terraform-linters/tflint",
          "rules": [
            {
              "id": "test_rule",
              "shortDescription": {
                "text": ""
              },
              "helpUri": "https://github.com"
            }
          ]
        }
      },
      "results": [
        {
          "ruleId": "test_rule",
          "level": "error",
          "message": {
            "text": "test"
          },
          "locations": [
            {
              "physicalLocation": {
                "artifactLocation": {
                  "uri": "test.tf"
                }
              }
            }
          ]
        }
      ]
    },
    {
      "tool": {
        "driver": {
          "name": "tflint-errors",
          "version": "0.45.0",
          "informationUri": "https://github.com/terraform-linters/tflint"
        }
      },
      "results": [
        {
          "ruleId": "application_error",
          "level": "error",
          "message": {
            "text": "Failed to work; I don't feel like working"
          }
        }
      ]
    }
  ]
}`,
		},
		{
			Name: "HCL diagnostics are surfaced as tflint-errors",
			Error: fmt.Errorf(
				"babel fish confused; %w",
				hcl.Diagnostics{
					&hcl.Diagnostic{
						Severity: hcl.DiagWarning,
						Summary:  "summary",
						Detail:   "detail",
						Subject: &hcl.Range{
							Filename: "filename",
							Start:    hcl.Pos{Line: 1, Column: 1, Byte: 0},
							End:      hcl.Pos{Line: 5, Column: 1, Byte: 4},
						},
					},
				},
			),
			Stdout: `{
  "version": "2.1.0",
  "$schema": "https://json.schemastore.org/sarif-2.1.0-rtm.5.json",
  "runs": [
    {
      "tool": {
        "driver": {
          "name": "tflint",
          "version": "0.45.0",
          "informationUri": "https://github.com/terraform-linters/tflint"
        }
      },
      "results": []
    },
    {
      "tool": {
        "driver": {
          "name": "tflint-errors",
          "version": "0.45.0",
          "informationUri": "https://github.com/terraform-linters/tflint"
        }
      },
      "results": [
        {
          "ruleId": "summary",
          "level": "warning",
          "message": {
            "text": "detail"
          },
          "locations": [
            {
              "physicalLocation": {
                "artifactLocation": {
                  "uri": "filename"
                },
                "region": {
                  "startLine": 1,
                  "startColumn": 1,
                  "endLine": 5,
                  "endColumn": 1,
                  "byteOffset": 0,
                  "byteLength": 4
                }
              }
            }
          ]
        }
      ]
    }
  ]
}`,
		},
	}

	for _, tc := range cases {
		t.Run(tc.Name, func(t *testing.T) {
			stdout := &bytes.Buffer{}
			stderr := &bytes.Buffer{}
			formatter := &Formatter{Stdout: stdout, Stderr: stderr, Format: "sarif"}

			formatter.Print(tc.Issues, tc.Error, map[string][]byte{})

			if diff := cmp.Diff(tc.Stdout, stdout.String()); diff != "" {
				t.Fatalf("Failed %s test: %s", tc.Name, diff)
			}

			schemaLoader := gojsonschema.NewReferenceLoader("http://json.schemastore.org/sarif-2.1.0")
			result, err := gojsonschema.Validate(schemaLoader, gojsonschema.NewStringLoader(stdout.String()))
			if err != nil {
				t.Error(err)
			}
			for _, err := range result.Errors() {
				t.Error(err)
			}
		})
	}
}

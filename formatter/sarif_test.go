package formatter

import (
	"bytes"
	"errors"
	"fmt"
	"os"
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
			Stdout: fmt.Sprintf(`{
  "version": "2.1.0",
  "$schema": "https://json.schemastore.org/sarif-2.1.0-rtm.5.json",
  "runs": [
    {
      "tool": {
        "driver": {
          "name": "tflint",
          "version": "%s",
          "informationUri": "https://github.com/terraform-linters/tflint"
        }
      },
      "results": []
    },
    {
      "tool": {
        "driver": {
          "name": "tflint-errors",
          "version": "%s",
          "informationUri": "https://github.com/terraform-linters/tflint"
        }
      },
      "results": []
    }
  ]
}`, tflint.Version, tflint.Version),
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
			Stdout: fmt.Sprintf(`{
  "version": "2.1.0",
  "$schema": "https://json.schemastore.org/sarif-2.1.0-rtm.5.json",
  "runs": [
    {
      "tool": {
        "driver": {
          "name": "tflint",
          "version": "%s",
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
          "version": "%s",
          "informationUri": "https://github.com/terraform-linters/tflint"
        }
      },
      "results": []
    }
  ]
}`, tflint.Version, tflint.Version),
		},
		{
			Name: "issues not on line 1",
			Issues: tflint.Issues{
				{
					Rule:    &testRule{},
					Message: "test",
					Range: hcl.Range{
						Filename: "test.tf",
						Start:    hcl.Pos{Line: 3, Column: 1, Byte: 0},
						End:      hcl.Pos{Line: 3, Column: 4, Byte: 3},
					},
				},
			},
			Stdout: fmt.Sprintf(`{
  "version": "2.1.0",
  "$schema": "https://json.schemastore.org/sarif-2.1.0-rtm.5.json",
  "runs": [
    {
      "tool": {
        "driver": {
          "name": "tflint",
          "version": "%s",
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
                  "startLine": 3,
                  "startColumn": 1,
                  "endLine": 3,
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
          "version": "%s",
          "informationUri": "https://github.com/terraform-linters/tflint"
        }
      },
      "results": []
    }
  ]
}`, tflint.Version, tflint.Version),
		},
		{
			Name: "issues spanning multiple lines",
			Issues: tflint.Issues{
				{
					Rule:    &testRule{},
					Message: "test",
					Range: hcl.Range{
						Filename: "test.tf",
						Start:    hcl.Pos{Line: 1, Column: 1, Byte: 0},
						End:      hcl.Pos{Line: 4, Column: 1, Byte: 3},
					},
				},
			},
			Stdout: fmt.Sprintf(`{
  "version": "2.1.0",
  "$schema": "https://json.schemastore.org/sarif-2.1.0-rtm.5.json",
  "runs": [
    {
      "tool": {
        "driver": {
          "name": "tflint",
          "version": "%s",
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
                  "endLine": 4,
                  "endColumn": 1
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
          "version": "%s",
          "informationUri": "https://github.com/terraform-linters/tflint"
        }
      },
      "results": []
    }
  ]
}`, tflint.Version, tflint.Version),
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
			Stdout: fmt.Sprintf(`{
  "version": "2.1.0",
  "$schema": "https://json.schemastore.org/sarif-2.1.0-rtm.5.json",
  "runs": [
    {
      "tool": {
        "driver": {
          "name": "tflint",
          "version": "%s",
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
          "version": "%s",
          "informationUri": "https://github.com/terraform-linters/tflint"
        }
      },
      "results": []
    }
  ]
}`, tflint.Version, tflint.Version),
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
			Stdout: fmt.Sprintf(`{
  "version": "2.1.0",
  "$schema": "https://json.schemastore.org/sarif-2.1.0-rtm.5.json",
  "runs": [
    {
      "tool": {
        "driver": {
          "name": "tflint",
          "version": "%s",
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
          "version": "%s",
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
}`, tflint.Version, tflint.Version),
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
			Stdout: fmt.Sprintf(`{
  "version": "2.1.0",
  "$schema": "https://json.schemastore.org/sarif-2.1.0-rtm.5.json",
  "runs": [
    {
      "tool": {
        "driver": {
          "name": "tflint",
          "version": "%s",
          "informationUri": "https://github.com/terraform-linters/tflint"
        }
      },
      "results": []
    },
    {
      "tool": {
        "driver": {
          "name": "tflint-errors",
          "version": "%s",
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
}`, tflint.Version, tflint.Version),
		},
		{
			Name: "joined errors",
			Error: errors.Join(
				errors.New("an error occurred"),
				errors.New("failed"),
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
			Stdout: fmt.Sprintf(`{
  "version": "2.1.0",
  "$schema": "https://json.schemastore.org/sarif-2.1.0-rtm.5.json",
  "runs": [
    {
      "tool": {
        "driver": {
          "name": "tflint",
          "version": "%s",
          "informationUri": "https://github.com/terraform-linters/tflint"
        }
      },
      "results": []
    },
    {
      "tool": {
        "driver": {
          "name": "tflint-errors",
          "version": "%s",
          "informationUri": "https://github.com/terraform-linters/tflint"
        }
      },
      "results": [
        {
          "ruleId": "application_error",
          "level": "error",
          "message": {
            "text": "an error occurred"
          }
        },
        {
          "ruleId": "application_error",
          "level": "error",
          "message": {
            "text": "failed"
          }
        },
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
}`, tflint.Version, tflint.Version),
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

			schema, err := os.ReadFile("sarif-2.1.0.json")
			if err != nil {
				t.Fatal(err)
			}

			result, err := gojsonschema.Validate(
				gojsonschema.NewBytesLoader(schema),
				gojsonschema.NewBytesLoader(stdout.Bytes()),
			)
			if err != nil {
				t.Error(err)
			}
			for _, err := range result.Errors() {
				t.Error(err)
			}
		})
	}
}

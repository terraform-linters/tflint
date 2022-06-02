package formatter

import (
	"crypto/md5"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"strings"

	"github.com/hashicorp/hcl/v2"
	"github.com/terraform-linters/tflint/tflint"
)

// CodeClimateIssue is a temporary structure for converting TFLint issues to CodeClimate report format.
// See specs here: https://github.com/codeclimate/platform/blob/master/spec/analyzers/SPEC.md#data-types
// We're only mapping the types for which we have data and the required ones
type CodeClimateIssue struct {
	Type           string                `json:"type"`
	CheckName      string                `json:"check_name"`
	Description    string                `json:"description"`
	Content        string                `json:"content,omitempty"`
	Categories     []string              `json:"categories"`
	Location       CodeClimateLocation   `json:"location"`
	OtherLocations []CodeClimateLocation `json:"other_locations,omitempty"`
	Fingerprint    string                `json:"fingerprint"`
	Severity       string                `json:"severity,omitempty"`
}

type CodeClimateLocation struct {
	Path      string               `json:"path"`
	Positions CodeClimatePositions `json:"positions"`
}

type CodeClimatePositions struct {
	Begin CodeClimatePosition `json:"begin"`
	End   CodeClimatePosition `json:"end,omitempty"`
}

type CodeClimatePosition struct {
	Line   int `json:"line"`
	Column int `json:"column,omitempty"`
}

// Downloads the provided link and returns it as a string
func downloadLinkContent(link string) string {
	resp, err := http.Get(link)

	// We don't care about the error as there's no way to recover from it.
	// In case of errors we just return an empty string
	if err != nil || resp.StatusCode != http.StatusOK {
		return ""
	}

	// No errors mean that we can go on and extract the text data
	defer resp.Body.Close()
	buf := new(strings.Builder)
	_, err = io.Copy(buf, resp.Body)

	// Again, we don't deal with errors
	if err != nil {
		return ""
	}

	// If everything went fine we can return the buffered string's contents
	return buf.String()
}

func getMD5Hash(text string) string {
	hasher := md5.New()
	hasher.Write([]byte(text))
	return hex.EncodeToString(hasher.Sum(nil))
}

// Map TFLint severities with the ones expected by Code Climate
func toCodeClimateSeverity(tfSeverity string) string {
	switch tfSeverity {
	case "error":
		return "critical"
	case "warning":
		return "minor"
	case "info":
		return "info"
	default:
		panic(fmt.Errorf("Unexpected severity type: %s", tfSeverity))
	}
}

func (f *Formatter) codeClimatePrint(issues tflint.Issues, appErr error) {
	ret := make([]CodeClimateIssue, len(issues))

	for idx, issue := range issues.Sort() {
		ret[idx] = CodeClimateIssue{
			Type:        "issue",
			CheckName:   issue.Rule.Name(),
			Description: issue.Message,
			Content:     downloadLinkContent(issue.Rule.Link()),
			Categories:  []string{"Style"},
			Location: CodeClimateLocation{
				Path: issue.Range.Filename,
				Positions: CodeClimatePositions{
					Begin: CodeClimatePosition{Line: issue.Range.Start.Line, Column: issue.Range.Start.Column},
					End:   CodeClimatePosition{Line: issue.Range.End.Line, Column: issue.Range.End.Column},
				},
			},
			OtherLocations: make([]CodeClimateLocation, len(issue.Callers)),
			Severity:       toCodeClimateSeverity(toSeverity(issue.Rule.Severity())),
			Fingerprint:    getMD5Hash(issue.Range.Filename + issue.Rule.Name() + issue.Message),
		}
		for i, caller := range issue.Callers {
			ret[idx].OtherLocations[i] = CodeClimateLocation{
				Path: caller.Filename,
				Positions: CodeClimatePositions{
					Begin: CodeClimatePosition{Line: caller.Start.Line, Column: caller.Start.Column},
					End:   CodeClimatePosition{Line: caller.End.Line, Column: caller.End.Column},
				},
			}
		}
	}

	if appErr != nil {
		var diags hcl.Diagnostics
		var codeClimateErrors []CodeClimateIssue
		if errors.As(appErr, &diags) {
			codeClimateErrors = make([]CodeClimateIssue, len(diags))
			for idx, diag := range diags {
				codeClimateErrors[idx] = CodeClimateIssue{
					Type:        "issue",
					CheckName:   "TFLint Error",
					Description: diag.Detail,
					Content:     diag.Summary,
					Categories:  []string{"Bug Risk"},
					Severity:    toCodeClimateSeverity(fromHclSeverity(diag.Severity)),
					Fingerprint: getMD5Hash(diag.Subject.Filename + diag.Detail),
					Location: CodeClimateLocation{
						Path: diag.Subject.Filename,
						Positions: CodeClimatePositions{
							Begin: CodeClimatePosition{Line: diag.Subject.Start.Line, Column: diag.Subject.Start.Column},
							End:   CodeClimatePosition{Line: diag.Subject.End.Line, Column: diag.Subject.End.Column},
						},
					},
				}
			}
		} else {
			codeClimateErrors = []CodeClimateIssue{{
				Type:        "issue",
				CheckName:   "TFLint Error",
				Description: appErr.Error(),
				Categories:  []string{"Bug Risk"},
				Severity:    toCodeClimateSeverity(toSeverity(tflint.ERROR)),
				Fingerprint: getMD5Hash(appErr.Error()),
				Location:    CodeClimateLocation{},
			}}
		}

		// Merge errors and issues
		ret = append(ret, codeClimateErrors...)
	}

	out, err := json.Marshal(ret)
	if err != nil {
		fmt.Fprint(f.Stderr, err)
	}
	fmt.Fprint(f.Stdout, string(out))
}

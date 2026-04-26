// SPDX-License-Identifier: MPL-2.0

package funcs

import (
	"testing"

	"github.com/terraform-linters/tflint-plugin-sdk/terraform/lang/marks"
	"github.com/zclconf/go-cty/cty"
)

func TestBase64Decode(t *testing.T) {
	tests := []struct {
		name    string
		input   cty.Value
		want    cty.Value
		wantErr string
	}{
		{
			name:  "valid string",
			input: cty.StringVal("YWJjMTIzIT8kKiYoKSctPUB+"),
			want:  cty.StringVal("abc123!?$*&()'-=@~"),
		},
		{
			name:  "preserves marks",
			input: cty.StringVal("YWJjMTIzIT8kKiYoKSctPUB+").Mark(marks.Sensitive),
			want:  cty.StringVal("abc123!?$*&()'-=@~").Mark(marks.Sensitive),
		},
		{
			name:  "unknown keeps marks",
			input: cty.UnknownVal(cty.String).Mark("a").Mark("b"),
			want:  cty.UnknownVal(cty.String).RefineNotNull().Mark("a").Mark("b"),
		},
		{
			name:    "invalid base64",
			input:   cty.StringVal("dfg"),
			wantErr: `failed to decode base64 data "dfg"`,
		},
		{
			name:    "sensitive invalid base64",
			input:   cty.StringVal("dfg").Mark(marks.Sensitive),
			wantErr: `failed to decode base64 data (sensitive value)`,
		},
		{
			name:    "invalid utf8",
			input:   cty.StringVal("whee"),
			wantErr: "the result of decoding the provided string is not valid UTF-8",
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			got, err := Base64Decode(test.input)
			assertCtyResult(t, got, err, test.want, test.wantErr)
		})
	}
}

func TestBase64Encode(t *testing.T) {
	got, err := Base64Encode(cty.StringVal("abc123!?$*&()'-=@~"))
	assertCtyResult(t, got, err, cty.StringVal("YWJjMTIzIT8kKiYoKSctPUB+"), "")
}

func TestBase64Gzip(t *testing.T) {
	got, err := Base64Gzip(cty.StringVal("test"))
	assertCtyResult(t, got, err, cty.StringVal("H4sIAAAAAAAA/ypJLS4BAAAA//8BAAD//wx+f9gEAAAA"), "")
}

func TestURLEncode(t *testing.T) {
	tests := []struct {
		name string
		in   string
		want string
	}{
		{name: "safe characters", in: "abc123-_", want: "abc123-_"},
		{name: "query string", in: "foo:bar@localhost?foo=bar&bar=baz", want: "foo%3Abar%40localhost%3Ffoo%3Dbar%26bar%3Dbaz"},
		{name: "slashes", in: "foo/bar", want: "foo%2Fbar"},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			got, err := URLEncode(cty.StringVal(test.in))
			assertCtyResult(t, got, err, cty.StringVal(test.want), "")
		})
	}
}

func TestTextEncodeBase64(t *testing.T) {
	tests := []struct {
		name    string
		input   cty.Value
		enc     cty.Value
		want    cty.Value
		wantErr string
	}{
		{
			name:  "utf16",
			input: cty.StringVal("abc123!?$*&()'-=@~"),
			enc:   cty.StringVal("UTF-16LE"),
			want:  cty.StringVal("YQBiAGMAMQAyADMAIQA/ACQAKgAmACgAKQAnAC0APQBAAH4A"),
		},
		{
			name:    "unsupported encoding",
			input:   cty.StringVal("abc123!?$*&()'-=@~"),
			enc:     cty.StringVal("NOT-EXISTS"),
			wantErr: `"NOT-EXISTS" is not a supported IANA encoding name or alias in this Terraform version`,
		},
		{
			name:    "unrepresentable string",
			input:   cty.StringVal("🤔"),
			enc:     cty.StringVal("cp437"),
			wantErr: `the given string contains characters that cannot be represented in IBM437`,
		},
		{
			name:  "unknown input",
			input: cty.UnknownVal(cty.String),
			enc:   cty.StringVal("windows-1250"),
			want:  cty.UnknownVal(cty.String).RefineNotNull(),
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			got, err := TextEncodeBase64(test.input, test.enc)
			assertCtyResult(t, got, err, test.want, test.wantErr)
		})
	}
}

func TestTextDecodeBase64(t *testing.T) {
	tests := []struct {
		name    string
		input   cty.Value
		enc     cty.Value
		want    cty.Value
		wantErr string
	}{
		{
			name:  "utf16",
			input: cty.StringVal("YQBiAGMAMQAyADMAIQA/ACQAKgAmACgAKQAnAC0APQBAAH4A"),
			enc:   cty.StringVal("UTF-16LE"),
			want:  cty.StringVal("abc123!?$*&()'-=@~"),
		},
		{
			name:    "unsupported encoding",
			input:   cty.StringVal("doesn't matter"),
			enc:     cty.StringVal("NOT-EXISTS"),
			wantErr: `"NOT-EXISTS" is not a supported IANA encoding name or alias in this Terraform version`,
		},
		{
			name:    "invalid base64",
			input:   cty.StringVal("<invalid base64>"),
			enc:     cty.StringVal("cp437"),
			wantErr: `the given value is has an invalid base64 symbol at offset 0`,
		},
		{
			name:    "invalid symbol",
			input:   cty.StringVal("gQ=="),
			enc:     cty.StringVal("windows-1250"),
			wantErr: `the given string contains symbols that are not defined for windows-1250`,
		},
		{
			name:  "unknown input",
			input: cty.UnknownVal(cty.String),
			enc:   cty.StringVal("windows-1250"),
			want:  cty.UnknownVal(cty.String).RefineNotNull(),
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			got, err := TextDecodeBase64(test.input, test.enc)
			assertCtyResult(t, got, err, test.want, test.wantErr)
		})
	}
}

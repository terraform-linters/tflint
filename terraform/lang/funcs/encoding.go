// SPDX-License-Identifier: MPL-2.0

package funcs

import (
	"bytes"
	"compress/gzip"
	"encoding/base64"
	"fmt"
	"log"
	"net/url"
	"unicode/utf8"

	"github.com/zclconf/go-cty/cty"
	"github.com/zclconf/go-cty/cty/function"
	"golang.org/x/text/encoding"
	"golang.org/x/text/encoding/ianaindex"
)

var Base64DecodeFunc = function.New(&function.Spec{
	Params: []function.Parameter{{
		Name:         "str",
		Type:         cty.String,
		AllowMarked:  true,
		AllowUnknown: true,
	}},
	Type:         function.StaticReturnType(cty.String),
	RefineResult: refineNotNull,
	Impl: func(args []cty.Value, _ cty.Type) (cty.Value, error) {
		str, strMarks := args[0].Unmark()
		if !str.IsKnown() {
			return cty.UnknownVal(cty.String).WithMarks(strMarks), nil
		}

		decoded, err := decodeBase64Bytes(str.AsString(), strMarks)
		if err != nil {
			return cty.UnknownVal(cty.String), err
		}
		if !utf8.Valid(decoded) {
			log.Printf("[DEBUG] the result of decoding the provided string is not valid UTF-8: %s", redactIfSensitive(string(decoded), strMarks))
			return cty.UnknownVal(cty.String), fmt.Errorf("the result of decoding the provided string is not valid UTF-8")
		}

		return cty.StringVal(string(decoded)).WithMarks(strMarks), nil
	},
})

var Base64EncodeFunc = function.New(&function.Spec{
	Params: []function.Parameter{{
		Name: "str",
		Type: cty.String,
	}},
	Type:         function.StaticReturnType(cty.String),
	RefineResult: refineNotNull,
	Impl: func(args []cty.Value, _ cty.Type) (cty.Value, error) {
		return cty.StringVal(base64.StdEncoding.EncodeToString([]byte(args[0].AsString()))), nil
	},
})

var TextEncodeBase64Func = function.New(&function.Spec{
	Params: []function.Parameter{
		{Name: "string", Type: cty.String},
		{Name: "encoding", Type: cty.String},
	},
	Type:         function.StaticReturnType(cty.String),
	RefineResult: refineNotNull,
	Impl: func(args []cty.Value, _ cty.Type) (cty.Value, error) {
		enc, encName, err := lookupIANAEncoding(args[1], 1)
		if err != nil {
			return cty.UnknownVal(cty.String), err
		}

		encodedInput, err := enc.NewEncoder().Bytes([]byte(args[0].AsString()))
		if err != nil {
			return cty.UnknownVal(cty.String), function.NewArgErrorf(0, "the given string contains characters that cannot be represented in %s", encName)
		}
		return cty.StringVal(base64.StdEncoding.EncodeToString(encodedInput)), nil
	},
})

var TextDecodeBase64Func = function.New(&function.Spec{
	Params: []function.Parameter{
		{Name: "source", Type: cty.String},
		{Name: "encoding", Type: cty.String},
	},
	Type:         function.StaticReturnType(cty.String),
	RefineResult: refineNotNull,
	Impl: func(args []cty.Value, _ cty.Type) (cty.Value, error) {
		enc, encName, err := lookupIANAEncoding(args[1], 1)
		if err != nil {
			return cty.UnknownVal(cty.String), err
		}

		decoded, err := base64.StdEncoding.DecodeString(args[0].AsString())
		if err != nil {
			if corrupt, ok := err.(base64.CorruptInputError); ok {
				return cty.UnknownVal(cty.String), function.NewArgErrorf(0, "the given value is has an invalid base64 symbol at offset %d", int(corrupt))
			}
			return cty.UnknownVal(cty.String), function.NewArgErrorf(0, "invalid source string: %w", err)
		}

		text, err := enc.NewDecoder().Bytes(decoded)
		if err != nil || bytes.ContainsRune(text, '�') {
			return cty.UnknownVal(cty.String), function.NewArgErrorf(0, "the given string contains symbols that are not defined for %s", encName)
		}
		return cty.StringVal(string(text)), nil
	},
})

var Base64GzipFunc = function.New(&function.Spec{
	Params: []function.Parameter{{
		Name: "str",
		Type: cty.String,
	}},
	Type:         function.StaticReturnType(cty.String),
	RefineResult: refineNotNull,
	Impl: func(args []cty.Value, _ cty.Type) (cty.Value, error) {
		encoded, err := gzipAndEncode(args[0].AsString())
		if err != nil {
			return cty.UnknownVal(cty.String), err
		}
		return cty.StringVal(encoded), nil
	},
})

var URLEncodeFunc = function.New(&function.Spec{
	Params: []function.Parameter{{
		Name: "str",
		Type: cty.String,
	}},
	Type:         function.StaticReturnType(cty.String),
	RefineResult: refineNotNull,
	Impl: func(args []cty.Value, _ cty.Type) (cty.Value, error) {
		return cty.StringVal(url.QueryEscape(args[0].AsString())), nil
	},
})

func Base64Decode(str cty.Value) (cty.Value, error) {
	return Base64DecodeFunc.Call([]cty.Value{str})
}

func Base64Encode(str cty.Value) (cty.Value, error) {
	return Base64EncodeFunc.Call([]cty.Value{str})
}

func Base64Gzip(str cty.Value) (cty.Value, error) {
	return Base64GzipFunc.Call([]cty.Value{str})
}

func URLEncode(str cty.Value) (cty.Value, error) {
	return URLEncodeFunc.Call([]cty.Value{str})
}

func TextEncodeBase64(str, enc cty.Value) (cty.Value, error) {
	return TextEncodeBase64Func.Call([]cty.Value{str, enc})
}

func TextDecodeBase64(str, enc cty.Value) (cty.Value, error) {
	return TextDecodeBase64Func.Call([]cty.Value{str, enc})
}

func decodeBase64Bytes(encoded string, marks cty.ValueMarks) ([]byte, error) {
	decoded, err := base64.StdEncoding.DecodeString(encoded)
	if err != nil {
		return nil, fmt.Errorf("failed to decode base64 data %s", redactIfSensitive(encoded, marks))
	}
	return decoded, nil
}

func lookupIANAEncoding(value cty.Value, argIndex int) (encoding.Encoding, string, error) {
	name := value.AsString()
	enc, err := ianaindex.IANA.Encoding(name)
	if err != nil || enc == nil {
		return nil, "", function.NewArgErrorf(argIndex, "%q is not a supported IANA encoding name or alias in this Terraform version", name)
	}

	canonicalName, err := ianaindex.IANA.Name(enc)
	if err != nil {
		canonicalName = name
	}
	return enc, canonicalName, nil
}

func gzipAndEncode(text string) (string, error) {
	var buffer bytes.Buffer

	writer := gzip.NewWriter(&buffer)
	if _, err := writer.Write([]byte(text)); err != nil {
		return "", fmt.Errorf("failed to write gzip raw data: %w", err)
	}
	if err := writer.Flush(); err != nil {
		return "", fmt.Errorf("failed to flush gzip writer: %w", err)
	}
	if err := writer.Close(); err != nil {
		return "", fmt.Errorf("failed to close gzip writer: %w", err)
	}

	return base64.StdEncoding.EncodeToString(buffer.Bytes()), nil
}

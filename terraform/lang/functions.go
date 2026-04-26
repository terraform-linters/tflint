// SPDX-License-Identifier: MPL-2.0

package lang

import (
	"github.com/hashicorp/hcl/v2/ext/tryfunc"
	ctyyaml "github.com/zclconf/go-cty-yaml"
	"github.com/zclconf/go-cty/cty"
	"github.com/zclconf/go-cty/cty/function"
	"github.com/zclconf/go-cty/cty/function/stdlib"

	"github.com/terraform-linters/tflint/terraform/collections"
	"github.com/terraform-linters/tflint/terraform/lang/funcs"
	terraformfuncs "github.com/terraform-linters/tflint/terraform/lang/funcs/terraform"
)

var impureFunctions = []string{
	"bcrypt",
	"timestamp",
	"uuid",
}

var filesystemFunctions = collections.NewSetCmp[string](
	"file",
	"fileexists",
	"fileset",
	"filebase64",
	"filebase64sha256",
	"filebase64sha512",
	"filemd5",
	"filesha1",
	"filesha256",
	"filesha512",
	"templatefile",
)

var templateFunctions = collections.NewSetCmp[string](
	"templatefile",
	"templatestring",
)

func (s *Scope) Functions() map[string]function.Function {
	s.funcsLock.Lock()
	defer s.funcsLock.Unlock()

	if s.funcs == nil {
		s.funcs = buildScopeFunctions(s.BaseDir, s.PureOnly)
	}
	return s.funcs
}

func buildScopeFunctions(baseDir string, pureOnly bool) map[string]function.Function {
	allFunctions := make(map[string]function.Function)
	callback := func() (map[string]function.Function, collections.Set[string], collections.Set[string]) {
		return allFunctions, filesystemFunctions, templateFunctions
	}

	coreFunctions := baseFunctionTable(baseDir, callback)
	if pureOnly {
		markImpureFunctionsUnknown(coreFunctions)
	}

	for name, fn := range coreFunctions {
		registerFunction(allFunctions, name, fn)
	}
	registerTerraformProviderFunctions(allFunctions)

	return allFunctions
}

func baseFunctionTable(baseDir string, callback func() (map[string]function.Function, collections.Set[string], collections.Set[string])) map[string]function.Function {
	return map[string]function.Function{
		"abs":              stdlib.AbsoluteFunc,
		"abspath":          funcs.AbsPathFunc,
		"alltrue":          funcs.AllTrueFunc,
		"anytrue":          funcs.AnyTrueFunc,
		"basename":         funcs.BasenameFunc,
		"base64decode":     funcs.Base64DecodeFunc,
		"base64encode":     funcs.Base64EncodeFunc,
		"base64gzip":       funcs.Base64GzipFunc,
		"base64sha256":     funcs.Base64Sha256Func,
		"base64sha512":     funcs.Base64Sha512Func,
		"bcrypt":           funcs.BcryptFunc,
		"can":              tryfunc.CanFunc,
		"ceil":             stdlib.CeilFunc,
		"chomp":            stdlib.ChompFunc,
		"cidrhost":         funcs.CidrHostFunc,
		"cidrnetmask":      funcs.CidrNetmaskFunc,
		"cidrsubnet":       funcs.CidrSubnetFunc,
		"cidrsubnets":      funcs.CidrSubnetsFunc,
		"chunklist":        stdlib.ChunklistFunc,
		"coalesce":         funcs.CoalesceFunc,
		"coalescelist":     stdlib.CoalesceListFunc,
		"compact":          stdlib.CompactFunc,
		"concat":           stdlib.ConcatFunc,
		"contains":         stdlib.ContainsFunc,
		"csvdecode":        stdlib.CSVDecodeFunc,
		"dirname":          funcs.DirnameFunc,
		"distinct":         stdlib.DistinctFunc,
		"element":          stdlib.ElementFunc,
		"endswith":         funcs.EndsWithFunc,
		"ephemeralasnull":  funcs.EphemeralAsNullFunc,
		"file":             funcs.MakeFileFunc(baseDir, false),
		"filebase64":       funcs.MakeFileFunc(baseDir, true),
		"filebase64sha256": funcs.MakeFileBase64Sha256Func(baseDir),
		"filebase64sha512": funcs.MakeFileBase64Sha512Func(baseDir),
		"fileexists":       funcs.MakeFileExistsFunc(baseDir),
		"filemd5":          funcs.MakeFileMd5Func(baseDir),
		"fileset":          funcs.MakeFileSetFunc(baseDir),
		"filesha1":         funcs.MakeFileSha1Func(baseDir),
		"filesha256":       funcs.MakeFileSha256Func(baseDir),
		"filesha512":       funcs.MakeFileSha512Func(baseDir),
		"flatten":          stdlib.FlattenFunc,
		"floor":            stdlib.FloorFunc,
		"format":           stdlib.FormatFunc,
		"formatdate":       stdlib.FormatDateFunc,
		"formatlist":       stdlib.FormatListFunc,
		"indent":           stdlib.IndentFunc,
		"index":            funcs.IndexFunc,
		"issensitive":      funcs.IssensitiveFunc,
		"join":             stdlib.JoinFunc,
		"jsondecode":       stdlib.JSONDecodeFunc,
		"jsonencode":       stdlib.JSONEncodeFunc,
		"keys":             stdlib.KeysFunc,
		"length":           funcs.LengthFunc,
		"list":             funcs.ListFunc,
		"log":              stdlib.LogFunc,
		"lookup":           funcs.LookupFunc,
		"lower":            stdlib.LowerFunc,
		"map":              funcs.MapFunc,
		"matchkeys":        funcs.MatchkeysFunc,
		"max":              stdlib.MaxFunc,
		"md5":              funcs.Md5Func,
		"merge":            stdlib.MergeFunc,
		"min":              stdlib.MinFunc,
		"nonsensitive":     funcs.NonsensitiveFunc,
		"one":              funcs.OneFunc,
		"parseint":         stdlib.ParseIntFunc,
		"pathexpand":       funcs.PathExpandFunc,
		"plantimestamp":    funcs.PlantimestampFunc,
		"pow":              stdlib.PowFunc,
		"range":            stdlib.RangeFunc,
		"regex":            stdlib.RegexFunc,
		"regexall":         stdlib.RegexAllFunc,
		"replace":          funcs.ReplaceFunc,
		"reverse":          stdlib.ReverseListFunc,
		"rsadecrypt":       funcs.RsaDecryptFunc,
		"sensitive":        funcs.SensitiveFunc,
		"setintersection":  stdlib.SetIntersectionFunc,
		"setproduct":       stdlib.SetProductFunc,
		"setsubtract":      stdlib.SetSubtractFunc,
		"setunion":         stdlib.SetUnionFunc,
		"sha1":             funcs.Sha1Func,
		"sha256":           funcs.Sha256Func,
		"sha512":           funcs.Sha512Func,
		"signum":           stdlib.SignumFunc,
		"slice":            stdlib.SliceFunc,
		"sort":             stdlib.SortFunc,
		"split":            stdlib.SplitFunc,
		"startswith":       funcs.StartsWithFunc,
		"strcontains":      funcs.StrContainsFunc,
		"strrev":           stdlib.ReverseFunc,
		"substr":           stdlib.SubstrFunc,
		"sum":              funcs.SumFunc,
		"templatefile":     funcs.MakeTemplateFileFunc(baseDir, callback),
		"templatestring":   funcs.MakeTemplateStringFunc(callback),
		"textdecodebase64": funcs.TextDecodeBase64Func,
		"textencodebase64": funcs.TextEncodeBase64Func,
		"timestamp":        funcs.TimestampFunc,
		"timeadd":          stdlib.TimeAddFunc,
		"timecmp":          funcs.TimeCmpFunc,
		"title":            stdlib.TitleFunc,
		"tobool":           funcs.MakeToFunc(cty.Bool),
		"tolist":           funcs.MakeToFunc(cty.List(cty.DynamicPseudoType)),
		"tomap":            funcs.MakeToFunc(cty.Map(cty.DynamicPseudoType)),
		"tonumber":         funcs.MakeToFunc(cty.Number),
		"toset":            funcs.MakeToFunc(cty.Set(cty.DynamicPseudoType)),
		"tostring":         funcs.MakeToFunc(cty.String),
		"transpose":        funcs.TransposeFunc,
		"trim":             stdlib.TrimFunc,
		"trimprefix":       stdlib.TrimPrefixFunc,
		"trimspace":        stdlib.TrimSpaceFunc,
		"trimsuffix":       stdlib.TrimSuffixFunc,
		"try":              tryfunc.TryFunc,
		"upper":            stdlib.UpperFunc,
		"urlencode":        funcs.URLEncodeFunc,
		"uuid":             funcs.UUIDFunc,
		"uuidv5":           funcs.UUIDV5Func,
		"values":           stdlib.ValuesFunc,
		"yamldecode":       ctyyaml.YAMLDecodeFunc,
		"yamlencode":       ctyyaml.YAMLEncodeFunc,
		"zipmap":           stdlib.ZipmapFunc,
	}
}

func registerFunction(all map[string]function.Function, name string, fn function.Function) {
	all[name] = fn
	all["core::"+name] = fn
}

func markImpureFunctionsUnknown(functions map[string]function.Function) {
	for _, name := range impureFunctions {
		if fn, ok := functions[name]; ok {
			functions[name] = function.Unpredictable(fn)
		}
	}
}

func registerTerraformProviderFunctions(functions map[string]function.Function) {
	functions["provider::terraform::encode_tfvars"] = terraformfuncs.EncodeTfvarsFunc
	functions["provider::terraform::decode_tfvars"] = terraformfuncs.DecodeTfvarsFunc
	functions["provider::terraform::encode_expr"] = terraformfuncs.EncodeExprFunc
}

func NewMockFunction(_ *FunctionCall) function.Function {
	return function.New(&function.Spec{
		VarParam: &function.Parameter{
			Type:             cty.DynamicPseudoType,
			AllowNull:        true,
			AllowUnknown:     true,
			AllowDynamicType: true,
			AllowMarked:      true,
		},
		Type: function.StaticReturnType(cty.DynamicPseudoType),
		Impl: func([]cty.Value, cty.Type) (cty.Value, error) {
			return cty.DynamicVal, nil
		},
	})
}

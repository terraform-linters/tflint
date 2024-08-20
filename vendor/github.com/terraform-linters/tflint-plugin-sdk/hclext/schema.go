package hclext

import (
	"fmt"
	"reflect"
	"sort"
	"strings"
)

// SchemaMode controls how the body's schema is declared.
//
//go:generate stringer -type=SchemaMode
type SchemaMode int32

const (
	// SchemaDefaultMode is a mode for explicitly declaring the structure of attributes and blocks.
	SchemaDefaultMode SchemaMode = iota
	// SchemaJustAttributesMode is the mode to extract body as attributes.
	// In this mode you don't need to declare schema for attributes or blocks.
	SchemaJustAttributesMode
)

// BodySchema represents the desired body.
// This structure is designed to have attributes similar to hcl.BodySchema.
type BodySchema struct {
	Mode       SchemaMode
	Attributes []AttributeSchema
	Blocks     []BlockSchema
}

// AttributeSchema represents the desired attribute.
// This structure is designed to have attributes similar to hcl.AttributeSchema.
type AttributeSchema struct {
	Name     string
	Required bool
}

// BlockSchema represents the desired block header and body schema.
// Unlike hcl.BlockHeaderSchema, this can set nested body schema.
// Instead, hclext.Block can't handle abstract values like hcl.Body,
// so you need to specify all nested schemas at once.
type BlockSchema struct {
	Type       string
	LabelNames []string

	Body *BodySchema
}

// ImpliedBodySchema is a derivative of gohcl.ImpliedBodySchema that produces hclext.BodySchema instead of hcl.BodySchema.
// Unlike gohcl.ImpliedBodySchema, it produces nested schemas.
// This method differs from gohcl.DecodeBody in several ways:
//
// - Does not support `body` and `remain` tags.
// - Does not support partial schema.
//
// @see https://github.com/hashicorp/hcl/blob/v2.11.1/gohcl/schema.go
func ImpliedBodySchema(val interface{}) *BodySchema {
	return impliedBodySchema(reflect.TypeOf(val))
}

func impliedBodySchema(ty reflect.Type) *BodySchema {
	if ty.Kind() == reflect.Ptr {
		ty = ty.Elem()
	}

	if ty.Kind() != reflect.Struct {
		panic(fmt.Sprintf("given type must be struct, not %s", ty.Name()))
	}

	var attrSchemas []AttributeSchema
	var blockSchemas []BlockSchema

	tags := getFieldTags(ty)

	attrNames := make([]string, 0, len(tags.Attributes))
	for n := range tags.Attributes {
		attrNames = append(attrNames, n)
	}
	sort.Strings(attrNames)
	for _, n := range attrNames {
		idx := tags.Attributes[n]
		optional := tags.Optional[n]
		field := ty.Field(idx)

		var required bool

		switch {
		case field.Type.Kind() != reflect.Ptr && !optional:
			required = true
		default:
			required = false
		}

		attrSchemas = append(attrSchemas, AttributeSchema{
			Name:     n,
			Required: required,
		})
	}

	blockNames := make([]string, 0, len(tags.Blocks))
	for n := range tags.Blocks {
		blockNames = append(blockNames, n)
	}
	sort.Strings(blockNames)
	for _, n := range blockNames {
		idx := tags.Blocks[n]
		field := ty.Field(idx)
		fty := field.Type
		if fty.Kind() == reflect.Slice {
			fty = fty.Elem()
		}
		if fty.Kind() == reflect.Ptr {
			fty = fty.Elem()
		}
		if fty.Kind() != reflect.Struct {
			panic(fmt.Sprintf(
				"schema 'block' tag kind cannot be applied to %s field %s: struct required", field.Type.String(), field.Name,
			))
		}
		ftags := getFieldTags(fty)
		var labelNames []string
		if len(ftags.Labels) > 0 {
			labelNames = make([]string, len(ftags.Labels))
			for i, l := range ftags.Labels {
				labelNames[i] = l.Name
			}
		}

		blockSchemas = append(blockSchemas, BlockSchema{
			Type:       n,
			LabelNames: labelNames,
			Body:       impliedBodySchema(fty),
		})
	}

	return &BodySchema{
		Attributes: attrSchemas,
		Blocks:     blockSchemas,
	}
}

type fieldTags struct {
	Attributes map[string]int
	Blocks     map[string]int
	Labels     []labelField
	Optional   map[string]bool
}

type labelField struct {
	FieldIndex int
	Name       string
}

func getFieldTags(ty reflect.Type) *fieldTags {
	ret := &fieldTags{
		Attributes: map[string]int{},
		Blocks:     map[string]int{},
		Optional:   map[string]bool{},
	}

	ct := ty.NumField()
	for i := 0; i < ct; i++ {
		field := ty.Field(i)
		tag := field.Tag.Get("hclext")
		if tag == "" {
			continue
		}

		comma := strings.Index(tag, ",")
		var name, kind string
		if comma != -1 {
			name = tag[:comma]
			kind = tag[comma+1:]
		} else {
			name = tag
			kind = "attr"
		}

		switch kind {
		case "attr":
			ret.Attributes[name] = i
		case "block":
			ret.Blocks[name] = i
		case "label":
			ret.Labels = append(ret.Labels, labelField{
				FieldIndex: i,
				Name:       name,
			})
		case "optional":
			ret.Attributes[name] = i
			ret.Optional[name] = true
		case "remain", "body":
			panic(fmt.Sprintf("'%s' tag is permitted in HCL, but not permitted in schema", kind))
		default:
			panic(fmt.Sprintf("invalid schema field tag kind %q on %s %q", kind, field.Type.String(), field.Name))
		}
	}

	return ret
}

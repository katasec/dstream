package config

import (
	"encoding/json"
	"fmt"
	"strconv"
	"strings"

	"github.com/hashicorp/hcl/v2"
	"github.com/katasec/dstream/proto"
	"github.com/zclconf/go-cty/cty"
)

type TaskBlock struct {
	Name       string       `hcl:"name,label"`
	Type       string       `hcl:"type,optional"`
	PluginPath string       `hcl:"plugin_path,optional"`
	PluginRef  string       `hcl:"plugin_ref,optional"`
	Config     *ConfigBlock `hcl:"config,block"`
	Input      *InputBlock  `hcl:"input,block"`
	Output     *OutputBlock `hcl:"output,block"`
}

type InputBlock struct {
	Provider string       `hcl:"provider,attr"`
	Config   *ConfigBlock `hcl:"config,block"`
}

type OutputBlock struct {
	Provider string       `hcl:"provider,attr"`
	Config   *ConfigBlock `hcl:"config,block"`
}

// Wrap the config block body so we can decode it later
type ConfigBlock struct {
	Remain hcl.Body `hcl:",remain"` // This captures everything in the config block
}

// ConfigAsStringMap parses the `config` block of a task and returns two maps:
//  1. A map of field names to their stringified values.
//  2. A map of field names to their detected types (e.g., "string", "list").
//
// Example:
//
// Given this HCL:
//
//	task "ingester-mssql" {
//	  config {
//	    db_connection_string = "Server=localhost"
//	    tables = ["Orders", "Customers"]
//	  }
//	}
//
// Returns:
//
//	values: {
//	  "db_connection_string": "Server=localhost",
//	  "tables": "Orders,Customers"
//	}
//	types: {
//	  "db_connection_string": "string",
//	  "tables": "list"
//	}
//
// This allows plugins to receive a uniform map of string values
func (t *TaskBlock) ConfigAsStringMap() (map[string]string, map[string]string, error) {
	schema := &hcl.BodySchema{
		Blocks: []hcl.BlockHeaderSchema{{Type: "config"}},
	}
	content, diags := t.Config.Remain.Content(schema)
	if diags.HasErrors() {
		return nil, nil, fmt.Errorf("config block error: %s", diags.Error())
	}
	if len(content.Blocks) == 0 {
		return nil, nil, fmt.Errorf("config block missing in task %q", t.Name)
	}

	attrs, diags := content.Blocks[0].Body.JustAttributes()
	if diags.HasErrors() {
		return nil, nil, fmt.Errorf("attribute decode error: %s", diags.Error())
	}

	return decodeAttributes(attrs)
}

// decodeAttributes parses attribute values and returns their stringified form and type.
func decodeAttributes(attrs hcl.Attributes) (map[string]string, map[string]string, error) {
	vals := make(map[string]string)
	types := make(map[string]string)

	for name, attr := range attrs {
		val, diags := attr.Expr.Value(nil)
		if diags.HasErrors() {
			return nil, nil, fmt.Errorf("value error for %s: %s", name, diags.Error())
		}

		strVal, fieldType := decodeCtyValue(val)
		vals[name] = strVal
		types[name] = fieldType
	}

	return vals, types, nil
}

// decodeCtyValue takes a cty.Value and returns:
//  1. A stringified representation of the value (e.g., for transmission to a plugin or logging).
//  2. A type descriptor string that represents the value's inferred type (e.g., "string", "int", "list").
//
// This function simplifies HCL-native values by flattening them into a basic string form
// and tagging them with a type label. This is primarily used to send plugin configuration
// as a uniform map[string]string along with metadata, so the receiving plugin can
// deserialize or interpret the values appropriately.
//
// Supported value types include: strings, numbers, booleans, lists/tuples of strings,
// and map/object types (which are serialized to JSON).
//
// Example:
//
//	decodeCtyValue(cty.StringVal("hello"))
//	  → ("hello", "string")
//
//	decodeCtyValue(cty.ListVal([]cty.Value{cty.StringVal("a"), cty.StringVal("b")}))
//	  → ("a,b", "list")
//
//	decodeCtyValue(cty.NumberIntVal(42))
//	  → ("42", "int")
func decodeCtyValue(val cty.Value) (string, string) {
	valType := val.Type() // Reuse this throughout to avoid repeated calls

	switch {
	case valType.Equals(cty.String):
		// Simple string value
		return val.AsString(), proto.FieldTypeString

	case valType.Equals(cty.Number):
		// Numeric value converted to string using full precision
		return val.AsBigFloat().Text('f', -1), proto.FieldTypeInt

	case valType.Equals(cty.Bool):
		// Boolean value converted to "true"/"false"
		return strconv.FormatBool(val.True()), proto.FieldTypeBool

	case valType.IsListType() && valType.ElementType().Equals(cty.String):
		// List of strings converted to comma-separated string
		return joinStrings(val.AsValueSlice()), proto.FieldTypeList

	case valType.IsTupleType():
		// If all tuple elements are strings, treat like a list
		if isTupleOfStrings(valType) {
			return joinStrings(val.AsValueSlice()), proto.FieldTypeList
		}
		return val.GoString(), proto.FieldTypeList

	case valType.IsMapType() || valType.IsObjectType():
		// Serialize map/object to JSON or fallback to Go-string
		if b, err := json.Marshal(val.GoString()); err == nil {
			return string(b), proto.FieldTypeMap
		}
		return val.GoString(), proto.FieldTypeMap

	default:
		// Fallback for unrecognized types
		return val.GoString(), proto.FieldTypeString
	}
}

// joinStrings takes a slice of cty.Value (assumed to be strings)
// and joins them into a single comma-separated string.
//
// For example:
//
//	Input:  []cty.Value{cty.StringVal("A"), cty.StringVal("B")}
//	Output: "A,B"
func joinStrings(vs []cty.Value) string {
	out := make([]string, len(vs))
	for i, v := range vs {
		out[i] = v.AsString()
	}
	return strings.Join(out, ",")
}

// isTupleOfStrings returns true if all elements in the tuple are strings.
func isTupleOfStrings(t cty.Type) bool {
	for _, et := range t.TupleElementTypes() {
		if !et.Equals(cty.String) {
			return false
		}
	}
	return true
}

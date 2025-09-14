package config

import (
	"encoding/json"
	"fmt"
	"math/big"
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
	Provider     string       `hcl:"provider,optional"`
	ProviderPath string       `hcl:"provider_path,optional"`
	ProviderRef  string       `hcl:"provider_ref,optional"`
	Config       *ConfigBlock `hcl:"config,block"`
}

type OutputBlock struct {
	Provider     string       `hcl:"provider,optional"`
	ProviderPath string       `hcl:"provider_path,optional"`
	ProviderRef  string       `hcl:"provider_ref,optional"`
	Config       *ConfigBlock `hcl:"config,block"`
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

// InputConfigAsJSON serializes the input config block as JSON with proper types
func (t *TaskBlock) InputConfigAsJSON() (string, error) {
	if t.Input == nil || t.Input.Config == nil {
		return "{}", nil
	}
	
	attrs, diags := t.Input.Config.Remain.JustAttributes()
	if diags.HasErrors() {
		return "", fmt.Errorf("input config decode error: %s", diags.Error())
	}
	
	config, err := attributesToJSON(attrs)
	if err != nil {
		return "", fmt.Errorf("input config serialization error: %w", err)
	}
	
	return config, nil
}

// OutputConfigAsJSON serializes the output config block as JSON with proper types
func (t *TaskBlock) OutputConfigAsJSON() (string, error) {
	if t.Output == nil || t.Output.Config == nil {
		return "{}", nil
	}
	
	attrs, diags := t.Output.Config.Remain.JustAttributes()
	if diags.HasErrors() {
		return "", fmt.Errorf("output config decode error: %s", diags.Error())
	}
	
	config, err := attributesToJSON(attrs)
	if err != nil {
		return "", fmt.Errorf("output config serialization error: %w", err)
	}
	
	return config, nil
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

// attributesToJSON converts HCL attributes to JSON while preserving proper types
func attributesToJSON(attrs hcl.Attributes) (string, error) {
	config := make(map[string]interface{})
	
	for name, attr := range attrs {
		val, diags := attr.Expr.Value(nil)
		if diags.HasErrors() {
			return "", fmt.Errorf("value error for %s: %s", name, diags.Error())
		}
		
		// Convert cty.Value to proper Go type for JSON
		goVal, err := ctyValueToGo(val)
		if err != nil {
			return "", fmt.Errorf("convert %s: %w", name, err)
		}
		
		config[name] = goVal
	}
	
	jsonBytes, err := json.Marshal(config)
	if err != nil {
		return "", fmt.Errorf("JSON marshal error: %w", err)
	}
	
	return string(jsonBytes), nil
}

// ctyValueToGo converts a cty.Value to appropriate Go type for JSON serialization
func ctyValueToGo(val cty.Value) (interface{}, error) {
	valType := val.Type()
	
	switch {
	case valType.Equals(cty.String):
		return val.AsString(), nil
		
	case valType.Equals(cty.Number):
		// Try to convert to int first, then float
		bf := val.AsBigFloat()
		if bf.IsInt() {
			if intVal, accuracy := bf.Int64(); accuracy == big.Exact {
				return int(intVal), nil
			}
		}
		// Fallback to float64
		if floatVal, accuracy := bf.Float64(); accuracy == big.Exact {
			return floatVal, nil
		}
		return bf.String(), nil // Fallback to string for very large numbers
		
	case valType.Equals(cty.Bool):
		return val.True(), nil
		
	case valType.IsListType() || valType.IsTupleType():
		slice := val.AsValueSlice()
		result := make([]interface{}, len(slice))
		for i, item := range slice {
			goVal, err := ctyValueToGo(item)
			if err != nil {
				return nil, fmt.Errorf("convert list item %d: %w", i, err)
			}
			result[i] = goVal
		}
		return result, nil
		
	case valType.IsMapType() || valType.IsObjectType():
		valMap := val.AsValueMap()
		result := make(map[string]interface{})
		for k, v := range valMap {
			goVal, err := ctyValueToGo(v)
			if err != nil {
				return nil, fmt.Errorf("convert map key %s: %w", k, err)
			}
			result[k] = goVal
		}
		return result, nil
		
	default:
		return val.GoString(), nil
	}
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

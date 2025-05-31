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
	Name       string `hcl:"name,label"`
	Type       string `hcl:"type,optional"`
	PluginPath string `hcl:"plugin_path,optional"`
	PluginRef  string `hcl:"plugin_ref,optional"`

	Config hcl.Body `hcl:",remain"`
}

// ConfigAsStringMap returns (values, types, error)
func (t *TaskBlock) ConfigAsStringMap() (map[string]string, map[string]string, error) {
	schema := &hcl.BodySchema{
		Blocks: []hcl.BlockHeaderSchema{{Type: "config"}},
	}
	content, diags := t.Config.Content(schema)
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

	vals := make(map[string]string)
	types := make(map[string]string)

	for name, attr := range attrs {
		val, diags := attr.Expr.Value(nil)
		if diags.HasErrors() {
			return nil, nil, fmt.Errorf("value error for %s: %s", name, diags.Error())
		}

		switch {
		case val.Type().Equals(cty.String):
			vals[name] = val.AsString()
			types[name] = proto.FieldTypeString

		case val.Type().Equals(cty.Number):
			// stringify the number
			vals[name] = val.AsBigFloat().Text('f', -1)
			types[name] = proto.FieldTypeInt

		case val.Type().Equals(cty.Bool):
			vals[name] = strconv.FormatBool(val.True())
			types[name] = proto.FieldTypeBool

		case val.Type().IsListType() && val.Type().ElementType().Equals(cty.String):
			vals[name] = joinStrings(val.AsValueSlice())
			types[name] = proto.FieldTypeList

		case val.Type().IsTupleType():
			allStr := true
			for _, et := range val.Type().TupleElementTypes() {
				if !et.Equals(cty.String) {
					allStr = false
					break
				}
			}
			if allStr {
				vals[name] = joinStrings(val.AsValueSlice())
				types[name] = proto.FieldTypeList
			} else {
				vals[name] = val.GoString()
				types[name] = proto.FieldTypeList
			}

		case val.Type().IsMapType() || val.Type().IsObjectType():
			if b, err := json.Marshal(val.GoString()); err == nil {
				vals[name] = string(b)
			} else {
				vals[name] = val.GoString()
			}
			types[name] = proto.FieldTypeMap

		default:
			vals[name] = val.GoString()
			types[name] = proto.FieldTypeString
		}
	}

	return vals, types, nil
}

func joinStrings(vs []cty.Value) string {
	out := make([]string, len(vs))
	for i, v := range vs {
		out[i] = v.AsString()
	}
	return strings.Join(out, ",")
}

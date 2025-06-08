package config

import (
	"fmt"

	"github.com/hashicorp/hcl/v2"
	"github.com/hashicorp/hcl/v2/hclsyntax"
	"github.com/zclconf/go-cty/cty"
	ctyjson "github.com/zclconf/go-cty/cty/json"
)

// DumpConfigAsJSON marshals the `config {}` block into JSON for inspection/debugging.
func (t *TaskBlock) DumpConfigAsJSON() (string, error) {
	log.Info("[DumpConfigAsJSON] Starting...")

	val, diags := decodeBodyToCty(t.Config.Remain)
	if diags.HasErrors() {
		return "", fmt.Errorf("decode error: %s", diags.Error())
	}

	log.Info("[DumpConfigAsJSON] Keys in decoded config:")
	keys := val.Type().AttributeTypes()
	if len(keys) == 0 {
		log.Info("  [None — config appears empty!]")
	} else {
		for k := range keys {
			v := val.GetAttr(k)
			log.Info(fmt.Sprintf("  - %s: %v", k, v.GoString()))
		}
	}

	log.Info("[DumpConfigAsJSON] cty.Value GoString():", val.GoString())

	jsonBytes, err := ctyjson.Marshal(val, val.Type())
	if err != nil {
		return "", fmt.Errorf("marshal to JSON failed: %w", err)
	}
	jsonOut := string(jsonBytes)

	if jsonOut == "{}" || jsonOut == "" {
		log.Warn("[DumpConfigAsJSON] WARNING: Output JSON is empty.")
	} else {
		log.Info("[DumpConfigAsJSON] Final JSON:", "json", jsonOut)
	}

	return jsonOut, nil
}

// decodeBodyToCty recursively evaluates any hcl.Body into a cty.Value object.
func decodeBodyToCty(body hcl.Body) (cty.Value, hcl.Diagnostics) {
	obj := make(map[string]cty.Value)

	ctx := &hcl.EvalContext{
		Variables: map[string]cty.Value{},
	}

	if syn, ok := body.(*hclsyntax.Body); ok {
		for name, attr := range syn.Attributes {
			val, diags := attr.Expr.Value(ctx)
			if diags.HasErrors() {
				log.Warn("[decodeBodyToCty] Attr error for", name, ":", diags.Error())
				obj[name] = cty.StringVal("[error: " + diags.Error() + "]")
			} else {
				obj[name] = val
			}
		}

		for _, block := range syn.Blocks {
			nested, _ := decodeBodyToCty(block.Body)

			if existing, ok := obj[block.Type]; ok {
				if existing.Type().IsTupleType() || existing.Type().IsListType() {
					obj[block.Type] = cty.ListVal(append(existing.AsValueSlice(), nested))
				} else {
					obj[block.Type] = cty.ListVal([]cty.Value{existing, nested})
				}
			} else {
				obj[block.Type] = nested
			}
		}
		return cty.ObjectVal(obj), nil
	}

	log.Warn("[decodeBodyToCty] Received non-hclsyntax.Body — ignoring content")
	return cty.ObjectVal(obj), nil
}

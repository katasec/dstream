package config

import (
	"encoding/json"
	"fmt"

	"github.com/hashicorp/hcl/v2"
	ctyjson "github.com/zclconf/go-cty/cty/json"
	"google.golang.org/protobuf/types/known/structpb"
)

// bodyToStructPB converts any HCL body—including nested blocks—into a
// google.protobuf.Struct suitable for shipping to a plugin.
func bodyToStructPB(body hcl.Body) (*structpb.Struct, error) {
	log.Info("[bodyToStructPB] Starting conversion...")

	// First try to decode the body to cty.Value
	val, diags := decodeBodyToCty(body)
	if diags.HasErrors() {
		return nil, fmt.Errorf("failed to decode body: %s", diags.Error())
	}

	// Convert to JSON
	jsonBytes, err := ctyjson.Marshal(val, val.Type())
	if err != nil {
		return nil, fmt.Errorf("marshal to JSON failed: %w", err)
	}
	log.Info("[bodyToStructPB] Raw JSON:", string(jsonBytes))

	// Unmarshal into a map
	var configMap map[string]interface{}
	if err := json.Unmarshal(jsonBytes, &configMap); err != nil {
		return nil, fmt.Errorf("unmarshal failed: %w", err)
	}

	// Create structpb from the map
	pb, err := structpb.NewStruct(configMap)
	if err != nil {
		return nil, fmt.Errorf("failed to create struct: %w", err)
	}

	// Log the final struct for debugging
	log.Info("[bodyToStructPB] Final config map being passed to plugin:")
	for k, v := range configMap {
		log.Info(fmt.Sprintf("  - %s: %+v", k, v))
	}

	return pb, nil
}

// TaskBlock wrapper
func (t *TaskBlock) ConfigAsStructPB() (*structpb.Struct, error) {
	log.Info("[ConfigAsStructPB] Converting config block to structpb.Struct")
	return bodyToStructPB(t.Config.Remain)
}

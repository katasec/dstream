package config

import (
	"encoding/json"
	"fmt"

	"github.com/hashicorp/hcl/v2"
	"github.com/tmccombs/hcl2json/convert"
	"google.golang.org/protobuf/types/known/structpb"
)

// bodyToStructPB converts any HCL body—including nested blocks—into a
// google.protobuf.Struct suitable for shipping to a plugin.
func bodyToStructPB(body hcl.Body) (*structpb.Struct, error) {
	log.Info("[bodyToStructPB] Starting conversion...")

	file := &hcl.File{Body: body}
	obj, err := convert.File(file, convert.Options{})
	if err != nil {
		return nil, fmt.Errorf("failed to convert HCL body: %w", err)
	}

	// Marshal and unmarshal to normalize types
	raw, err := json.Marshal(obj)
	if err != nil {
		return nil, fmt.Errorf("marshal failed: %w", err)
	}
	log.Info("[bodyToStructPB] Raw object from HCL:", string(raw))

	var top interface{}
	if err := json.Unmarshal(raw, &top); err != nil {
		log.Error("[bodyToStructPB] Unmarshal failed:", err)
		return nil, fmt.Errorf("decode config: %w", err)
	}

	if topMap, ok := top.(map[string]interface{}); ok {
		log.Info("[bodyToStructPB] Top-level is map — using as-is.")
		return structpb.NewStruct(topMap)
	}

	log.Warn("[bodyToStructPB] Top-level is NOT map — wrapping manually under 'value'")
	return structpb.NewStruct(map[string]interface{}{"value": top})
}

// TaskBlock wrapper
func (t *TaskBlock) ConfigAsStructPB() (*structpb.Struct, error) {
	log.Info("[ConfigAsStructPB] Converting config block to structpb.Struct")
	return bodyToStructPB(t.Config.Remain)
}

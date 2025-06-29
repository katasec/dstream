package config

import (
	"encoding/json"
	"fmt"

	"github.com/hashicorp/hcl/v2"
	"github.com/katasec/dstream/proto"
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
	//log.Info("[bodyToStructPB] Raw JSON:", string(jsonBytes))

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

	// // Log the final struct for debugging
	// log.Info("[bodyToStructPB] Final config map being passed to plugin:")
	// for k, v := range configMap {
	// 	log.Info(fmt.Sprintf("  - %s: %+v", k, v))
	// }

	return pb, nil
}

// TaskBlock wrapper
func (t *TaskBlock) ConfigAsStructPB() (*structpb.Struct, error) {
	log.Info("[ConfigAsStructPB] Converting config block to structpb.Struct")
	return bodyToStructPB(t.Config.Remain)
}

// InputAsStructPB converts the input block to a proto.InputConfig
func (t *TaskBlock) InputAsStructPB() (*proto.InputConfig, error) {
	if t.Input == nil {
		return nil, nil
	}
	
	var configStruct *structpb.Struct
	var err error
	
	if t.Input.Config != nil {
		configStruct, err = bodyToStructPB(t.Input.Config.Remain)
		if err != nil {
			return nil, fmt.Errorf("decode input config: %w", err)
		}
	} else {
		// Empty config
		configStruct = &structpb.Struct{Fields: make(map[string]*structpb.Value)}
	}
	
	return &proto.InputConfig{
		Provider: t.Input.Provider,
		Config:   configStruct,
	}, nil
}

// OutputAsStructPB converts the output block to a proto.OutputConfig
func (t *TaskBlock) OutputAsStructPB() (*proto.OutputConfig, error) {
	if t.Output == nil {
		return nil, nil
	}
	
	var configStruct *structpb.Struct
	var err error
	
	if t.Output.Config != nil {
		configStruct, err = bodyToStructPB(t.Output.Config.Remain)
		if err != nil {
			return nil, fmt.Errorf("decode output config: %w", err)
		}
	} else {
		// Empty config
		configStruct = &structpb.Struct{Fields: make(map[string]*structpb.Value)}
	}
	
	return &proto.OutputConfig{
		Provider: t.Output.Provider,
		Config:   configStruct,
	}, nil
}

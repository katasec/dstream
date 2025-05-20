package config

import (
	"fmt"

	"github.com/hashicorp/hcl/v2"
	"github.com/hashicorp/hcl/v2/hclsyntax"
)

// ExtractConfigBlock returns the raw `config {}` block for the given task (by type and name).
func ExtractConfigBlock(rawHCL, taskType, taskName string) ([]byte, error) {
	file, diags := hclsyntax.ParseConfig([]byte(rawHCL), "input.hcl", hcl.InitialPos)
	if diags.HasErrors() {
		return nil, fmt.Errorf("HCL parse error: %s", diags.Error())
	}

	syntaxBody, ok := file.Body.(*hclsyntax.Body)
	if !ok {
		return nil, fmt.Errorf("failed to cast body to hclsyntax.Body")
	}

	for _, block := range syntaxBody.Blocks {
		if block.Type == "task" && len(block.Labels) == 2 &&
			block.Labels[0] == taskType && block.Labels[1] == taskName {

			for _, sub := range block.Body.Blocks {
				if sub.Type == "config" {
					start := sub.Range().Start.Byte
					end := sub.Range().End.Byte
					return []byte(rawHCL[start:end]), nil
				}
			}

			return nil, fmt.Errorf("task %q %q found, but has no config block", taskType, taskName)
		}
	}

	return nil, fmt.Errorf("no task block found with type %q and name %q", taskType, taskName)
}

// func ExtractConfigBlock(rawHCL, taskType, taskName string) ([]byte, error) {
// 	file, diags := hclsyntax.ParseConfig([]byte(rawHCL), "input.hcl", hcl.InitialPos)
// 	if diags.HasErrors() {
// 		return nil, fmt.Errorf("HCL parse error: %s", diags.Error())
// 	}

// 	syntaxBody, ok := file.Body.(*hclsyntax.Body)
// 	if !ok {
// 		return nil, fmt.Errorf("failed to cast body to hclsyntax.Body")
// 	}

// 	for _, block := range syntaxBody.Blocks {
// 		if block.Type == "task" && len(block.Labels) == 2 &&
// 			block.Labels[0] == taskType && block.Labels[1] == taskName {

// 			for _, sub := range block.Body.Blocks {
// 				if sub.Type == "config" {
// 					start := sub.DefRange().Start.Byte
// 					end := sub.DefRange().End.Byte
// 					return []byte(strings.TrimSpace(rawHCL[start:end])), nil
// 				}
// 			}

// 			return nil, fmt.Errorf("task %q %q found, but has no config block", taskType, taskName)
// 		}
// 	}

// 	return nil, fmt.Errorf("no task block found with type %q and name %q", taskType, taskName)
// }

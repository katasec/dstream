package config

import (
	"fmt"

	"github.com/hashicorp/hcl/v2/hclsimple"
)

// LoadRootFile reads, templates, and decodes a full dstream HCL config file
func LoadRootFile(path string) (*RootHCL, error) {
	hclStr, err := RenderHCLTemplate(path)
	if err != nil {
		return nil, fmt.Errorf("template processing failed: %w", err)
	}

	var root RootHCL
	err = hclsimple.Decode(path, []byte(hclStr), nil, &root)
	if err != nil {
		return nil, fmt.Errorf("HCL decode failed: %w", err)
	}

	return &root, nil
}

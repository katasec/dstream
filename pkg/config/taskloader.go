package config

import (
	"bytes"
	"fmt"
	"path/filepath"
	"text/template"

	"github.com/Masterminds/sprig"
	"github.com/hashicorp/hcl/v2/hclsimple"
)

// GenerateHCL applies sprig templating to the HCL file before parsing
func GenerateHCL(filePath string) (string, error) {
	baseName := filepath.Base(filePath)
	tmpl, err := template.New(baseName).Funcs(sprig.TxtFuncMap()).ParseFiles(filePath)
	if err != nil {
		return "", fmt.Errorf("template parse error: %w", err)
	}

	var buf bytes.Buffer
	if err := tmpl.ExecuteTemplate(&buf, baseName, nil); err != nil {
		return "", fmt.Errorf("template execution error: %w", err)
	}

	return buf.String(), nil
}

// LoadTaskFile reads, templates, and decodes a task-based HCL config file
func LoadTaskFile(path string) (*TaskFile, error) {
	hclStr, err := GenerateHCL(path)
	if err != nil {
		return nil, fmt.Errorf("template processing failed: %w", err)
	}

	var tf TaskFile
	err = hclsimple.Decode(path, []byte(hclStr), nil, &tf)
	if err != nil {
		return nil, fmt.Errorf("HCL decode failed: %w", err)
	}

	return &tf, nil
}

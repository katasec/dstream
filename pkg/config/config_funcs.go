package config

import (
	"bytes"
	"fmt"
	"os"
	"path/filepath"
	"text/template"

	"github.com/Masterminds/sprig"
	"github.com/hashicorp/hcl/v2"
	"github.com/hashicorp/hcl/v2/gohcl"
	"github.com/hashicorp/hcl/v2/hclsyntax"
	"github.com/zclconf/go-cty/cty"
)

// LogConfigLoaded is a simple helper to log when configuration is loaded successfully
func LogConfigLoaded() {
	// Plugin-based architecture no longer requires database-specific validation
	log.Info("Configuration loaded successfully")
}

// RenderHCLTemplate applies sprig templating to the HCL file before parsing
func RenderHCLTemplate(filePath string) (string, error) {
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

// RenderHCLTemplateBytes renders a Sprig-enabled HCL template from byte content
func RenderHCLTemplateBytes(name string, content []byte) (string, error) {
	tmpl, err := template.New(name).Funcs(sprig.TxtFuncMap()).Parse(string(content))
	if err != nil {
		return "", fmt.Errorf("template parse error: %w", err)
	}

	var buf bytes.Buffer
	if err := tmpl.Execute(&buf, nil); err != nil {
		return "", fmt.Errorf("template execution error: %w", err)
	}

	return buf.String(), nil
}

// RenderHCLTemplate loads the file from disk and renders it via Sprig
func RenderHCLTemplateFile(filePath string) (string, error) {
	content, err := os.ReadFile(filePath)
	if err != nil {
		return "", fmt.Errorf("failed to read file: %w", err)
	}

	baseName := filepath.Base(filePath)
	return RenderHCLTemplateBytes(baseName, content)
}

// DecodeHCL returns a config object based on the provided config file
func DecodeHCL[T any](configHCL string, filePath string) T {
	// Parse HCL config starting from position 0
	src := []byte(configHCL)
	pos := hcl.Pos{Line: 0, Column: 0, Byte: 0}
	f, _ := hclsyntax.ParseConfig(src, filePath, pos)

	// Create evaluation context for locals support
	ctx := &hcl.EvalContext{
		Variables: make(map[string]cty.Value),
	}

	// First pass: Decode to extract locals
	var temp RootHCL
	decodeDiags := gohcl.DecodeBody(f.Body, ctx, &temp)

	// If locals exist, populate the context with local variables
	if temp.Locals != nil && temp.Locals.Vars != nil {
		for name, value := range temp.Locals.Vars {
			ctx.Variables[name] = interfaceToCtyValue(value)
		}
	}

	// Second pass: Decode again with locals in the context
	var c T
	decodeDiags = gohcl.DecodeBody(f.Body, ctx, &c)
	if decodeDiags.HasErrors() {
		log.Error("Error decoding HCL", "error", decodeDiags.Error())
		os.Exit(1)
	}

	return c
}

// interfaceToCtyValue converts Go interface{} to cty.Value for HCL variable resolution
func interfaceToCtyValue(val interface{}) cty.Value {
	switch v := val.(type) {
	case string:
		return cty.StringVal(v)
	case int, int32, int64:
		switch num := v.(type) {
		case int:
			return cty.NumberIntVal(int64(num))
		case int32:
			return cty.NumberIntVal(int64(num))
		case int64:
			return cty.NumberIntVal(num)
		}
	case float64:
		return cty.NumberFloatVal(v)
	case bool:
		return cty.BoolVal(v)
	case []interface{}:
		vals := make([]cty.Value, len(v))
		for i, elem := range v {
			vals[i] = interfaceToCtyValue(elem)
		}
		return cty.ListVal(vals)
	case map[string]interface{}:
		vals := make(map[string]cty.Value)
		for k, elem := range v {
			vals[k] = interfaceToCtyValue(elem)
		}
		return cty.MapVal(vals)
	default:
		return cty.StringVal(fmt.Sprintf("%v", val))
	}
	return cty.StringVal(fmt.Sprintf("%v", val))
}

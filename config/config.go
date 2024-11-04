package config

import (
	"bytes"
	"fmt"
	"log"
	"os"
	"text/template"
	"time"

	"github.com/Masterminds/sprig"
	"github.com/hashicorp/hcl/v2"
	"github.com/hashicorp/hcl/v2/gohcl"
	"github.com/hashicorp/hcl/v2/hclsyntax"
)

type TableConfig struct {
	Name            string `hcl:"name"`
	PollInterval    string `hcl:"poll_interval"`
	MaxPollInterval string `hcl:"max_poll_interval"`
}

type OutputConfig struct {
	Type string `hcl:"type"`
}

type Config struct {
	DBType                        string        `hcl:"db_type"`
	DBConnectionString            string        `hcl:"db_connection_string"`
	AzureEventHubConnectionString string        `hcl:"azure_event_hub_connection_string"`
	EventHubName                  string        `hcl:"azure_event_hub_name"`
	Output                        OutputConfig  `hcl:"output,block"`
	Tables                        []TableConfig `hcl:"tables,block"`
}

func (c *Config) CheckConfig() {
	if c.AzureEventHubConnectionString == "" {
		log.Println("Error, AzureEventHubConnectionString was not found, exitting.")
		os.Exit(0)
	}

	if c.DBConnectionString == "" {
		log.Println("Error, DBConnectionString was not found, exitting.")
		os.Exit(0)
	}

	if c.EventHubName == "" {
		log.Println("Error, EventHubName was not found, exitting.")
		os.Exit(0)
	}
}

// LoadConfig reads, processes the HCL configuration file, and replaces placeholders with environment variables
func LoadConfig(filePath string) (*Config, error) {
	var config Config

	// Gen HCL config post text templating
	hcl, err := generateHCL(filePath)
	if err != nil {
		log.Fatal(err)
	}

	// Read config from generated HCL
	config = processHCL(hcl, filePath)

	return &config, nil
}

// GetPollInterval returns the PollInterval as a time.Duration
func (t *TableConfig) GetPollInterval() (time.Duration, error) {
	return time.ParseDuration(t.PollInterval)
}

// GetMaxPollInterval returns the MaxPollInterval as a time.Duration
func (t *TableConfig) GetMaxPollInterval() (time.Duration, error) {
	return time.ParseDuration(t.MaxPollInterval)
}

// generateHCL Generates the an HCL config after processing the text templating
func generateHCL(filePath string) (hcl string, err error) {

	//Get the Sprig function map.
	fmap := sprig.TxtFuncMap()

	// Define template that *.hcl and *.tpl files in current folder
	// Ensure the Sprig functions are loaded for processing templates
	t := template.Must(template.New("test").
		Funcs(fmap).
		ParseFiles(filePath))

	buf := &bytes.Buffer{}
	err = t.ExecuteTemplate(buf, filePath, nil)
	if err != nil {
		fmt.Printf("Error during template execution: %s", err)
		return "", err
	}

	return buf.String(), nil
}

// processHCL returns a WIRE config object based on the provided config file
func processHCL(configHCL string, filePath string) (config Config) {

	// Parse HCL config starting from position 0
	src := []byte(configHCL)
	pos := hcl.Pos{Line: 0, Column: 0, Byte: 0}
	f, _ := hclsyntax.ParseConfig(src, filePath, pos)

	// Decode HCL into a config struct and return to caller
	var c Config
	decodeDiags := gohcl.DecodeBody(f.Body, nil, &c)
	if decodeDiags.HasErrors() {
		log.Fatal(decodeDiags.Error())
	}

	return c
}

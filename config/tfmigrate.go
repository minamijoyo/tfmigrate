package config

import (
	"fmt"

	"github.com/hashicorp/hcl/v2/hclsimple"
)

// SettingFile represents a file for CLI settings in HCL.
type SettingFile struct {
	// Tfmigrate is a top-level block.
	// It must contain only one block, and multiple blocks are not allowed.
	Tfmigrate TfmigrateBlock `hcl:"tfmigrate,block"`
}

// TfmigrateBlock represents a block for CLI settings in HCL.
type TfmigrateBlock struct {
	// History is a block for migration history management.
	History *HistoryBlock `hcl:"history,block"`
}

// TfmigrateConfig is a config for top-level CLI settings.
// TfmigrateBlock is just used for parsing HCL and
// TfmigrateConfig is used for building application logic.
type TfmigrateConfig struct {
	// History is a config for migration history management.
	History *HistoryConfig
}

// ParseSettingFile parses a given source of setting file and returns a TfmigrateConfig.
// Note that this method does not read a file and you should pass source of config in bytes.
// The filename is used for error message and selecting HCL syntax (.hcl and .json).
func ParseSettingFile(filename string, source []byte) (*TfmigrateConfig, error) {
	// Decode tfmigrate block.
	var f SettingFile
	err := hclsimple.Decode(filename, source, nil, &f)
	if err != nil {
		return nil, fmt.Errorf("failed to decode setting file: %s, err: %s", filename, err)
	}

	var config TfmigrateConfig
	if f.Tfmigrate.History != nil {
		history, err := parseHistoryBlock(*f.Tfmigrate.History)
		if err != nil {
			return nil, err
		}
		config.History = history
	}

	return &config, nil
}

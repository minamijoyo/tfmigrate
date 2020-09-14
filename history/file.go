package history

import (
	"encoding/json"
	"fmt"
)

// FileHeader contains a meta data for file format.
type FileHeader struct {
	// Version is a file format version.
	Version int `json:"version"`
}

// parseHistoryFile parses bytes and reteurns a History instance.
func parseHistoryFile(b []byte) (*History, error) {
	version, err := detectHistoryFileVersion(b)
	if err != nil {
		return nil, err
	}

	switch version {
	case 1:
		return parseHistoryFileV1(b)

	default:
		return nil, fmt.Errorf("unknown history file version: %d", version)
	}
}

// detectHistoryFileVersion detects a file formart version.
func detectHistoryFileVersion(b []byte) (int, error) {
	// peek a file header
	var header FileHeader
	err := json.Unmarshal(b, &header)
	if err != nil {
		return 0, err
	}

	return header.Version, nil
}

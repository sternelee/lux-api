//go:build wasm
// +build wasm

package api

import (
	"encoding/json"
	"fmt"

	"github.com/iawia002/lux/extractors"
	"github.com/iawia002/lux/request"
)

// LuxInfo extracts video info and returns it as a JSON string.
func LuxInfo(videoURL, format, quality string) (string, error) {
	request.SetOptions(request.Options{})

	data, err := extractors.Extract(videoURL, extractors.Options{})
	if err != nil {
		return "", fmt.Errorf("error: %v", err)
	}

	jsonData, err := json.Marshal(data)
	if err != nil {
		return "", fmt.Errorf("error: %v", err)
	}

	return string(jsonData), nil
}

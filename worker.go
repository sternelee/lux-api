package main

import (
	"encoding/json"
	"fmt"
	"syscall/js"

	"github.com/iawia002/lux/extractors"
	"github.com/iawia002/lux/request"
)

func main() {
	js.Global().Set("luxInfo", js.FuncOf(luxInfo))
	<-make(chan struct{})
}

func luxInfo(this js.Value, args []js.Value) interface{} {
	if len(args) < 1 {
		return createErrorResponse("Video URL is required", 400)
	}

	videoURL := args[0].String()
	
	if videoURL == "" {
		return createErrorResponse("Video URL cannot be empty", 400)
	}

	// Set request options
	request.SetOptions(request.Options{})

	// Extract video data using lux
	data, err := extractors.Extract(videoURL, extractors.Options{})
	if err != nil {
		return createErrorResponse(fmt.Sprintf("Failed to extract video: %v", err), 500)
	}

	// Convert to JSON
	jsonData, err := json.Marshal(data)
	if err != nil {
		return createErrorResponse(fmt.Sprintf("Failed to marshal response: %v", err), 500)
	}

	return string(jsonData)
}

func createErrorResponse(message string, statusCode int) interface{} {
	response := map[string]interface{}{
		"status":  "error",
		"message": message,
		"code":    statusCode,
	}
	
	jsonData, _ := json.Marshal(response)
	return string(jsonData)
}
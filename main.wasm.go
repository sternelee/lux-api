//go:build wasm
// +build wasm

package main

import (
	"encoding/json"
	"fmt"
	"syscall/js"

	"github.com/iawia002/lux/extractors"
	"github.com/iawia002/lux/request"
)

func main() {
	// Initialize the module
	fmt.Println("Lux WASM module initializing...")
	
	// Register functions
	js.Global().Set("luxInfo", js.FuncOf(luxInfo))
	js.Global().Set("luxExtract", js.FuncOf(luxExtract))
	
	// Signal that module is ready
	js.Global().Set("luxReady", true)
	
	fmt.Println("Lux WASM module ready")
	
	// Keep the Go program running
	select {}
}

// luxInfo provides basic extraction without downloading
func luxInfo(this js.Value, args []js.Value) interface{} {
	// Create a promise to handle async operations
	promiseConstructor := js.Global().Get("Promise")
	return promiseConstructor.New(js.FuncOf(func(this js.Value, promiseArgs []js.Value) interface{} {
		resolve := promiseArgs[0]
		reject := promiseArgs[1]
		
		// Run extraction in goroutine
		go func() {
			result, err := extractVideo(args, false)
			if err != nil {
				reject.Invoke(err.Error())
			} else {
				resolve.Invoke(result)
			}
		}()
		
		return nil
	}))
}

// luxExtract provides full extraction with download URLs
func luxExtract(this js.Value, args []js.Value) interface{} {
	// Create a promise to handle async operations
	promiseConstructor := js.Global().Get("Promise")
	return promiseConstructor.New(js.FuncOf(func(this js.Value, promiseArgs []js.Value) interface{} {
		resolve := promiseArgs[0]
		reject := promiseArgs[1]
		
		// Run extraction in goroutine
		go func() {
			result, err := extractVideo(args, true)
			if err != nil {
				reject.Invoke(err.Error())
			} else {
				resolve.Invoke(result)
			}
		}()
		
		return nil
	}))
}

func extractVideo(jsArgs []js.Value, includeDownload bool) (string, error) {
	// Validate arguments
	if len(jsArgs) < 1 {
		return createErrorJSON("URL parameter is required", 400), nil
	}
	
	videoURL := jsArgs[0].String()
	if videoURL == "" {
		return createErrorJSON("URL cannot be empty", 400), nil
	}
	
	// Parse optional parameters
	options := extractors.Options{}
	
	if len(jsArgs) > 1 && !jsArgs[1].IsNull() && !jsArgs[1].IsUndefined() {
		// Parse options object
		opts := jsArgs[1]
		
		// Check for playlist option
		if playlist := opts.Get("playlist"); !playlist.IsUndefined() {
			options.Playlist = playlist.Bool()
		}
		
		// Check for items option
		if items := opts.Get("items"); !items.IsUndefined() {
			options.Items = items.String()
		}
		
		// Check for cookie option
		if cookie := opts.Get("cookie"); !cookie.IsUndefined() {
			options.Cookie = cookie.String()
		}
	}
	
	// Set request options
	request.SetOptions(request.Options{})
	
	// Extract video data
	dataList, err := extractors.Extract(videoURL, options)
	if err != nil {
		return createErrorJSON(fmt.Sprintf("Extraction failed: %v", err), 500), nil
	}
	
	// Check if we got any data
	if len(dataList) == 0 {
		return createErrorJSON("No data extracted", 404), nil
	}
	
	// Process the first item (or all items if playlist)
	var results []map[string]interface{}
	
	for _, data := range dataList {
		// Check for extraction errors
		if data.Err != nil {
			continue // Skip items with errors
		}
		
		// Create response structure for this item
		item := map[string]interface{}{
			"url":     data.URL,
			"site":    data.Site,
			"title":   data.Title,
			"type":    data.Type,
			"streams": formatStreams(data.Streams),
		}
		
		// Add captions if available
		if len(data.Captions) > 0 {
			item["captions"] = data.Captions
		}
		
		// Add download info if requested
		if includeDownload && len(data.Streams) > 0 {
			downloadInfo := make(map[string]interface{})
			for id, stream := range data.Streams {
				downloadInfo[id] = map[string]interface{}{
					"quality": stream.Quality,
					"size":    stream.Size,
					"ext":     stream.Ext,
					"urls":    extractURLs(stream),
				}
			}
			item["downloads"] = downloadInfo
		}
		
		results = append(results, item)
	}
	
	// Create final response
	response := map[string]interface{}{
		"status": "success",
		"data":   results[0], // Return first item by default
		"metadata": map[string]interface{}{
			"extracted_at": js.Global().Get("Date").New().Call("toISOString").String(),
			"method":       "wasm",
		},
	}
	
	// If playlist mode and multiple results, return all
	if options.Playlist && len(results) > 1 {
		response["data"] = results
		response["metadata"].(map[string]interface{})["playlist"] = true
		response["metadata"].(map[string]interface{})["count"] = len(results)
	}
	
	// Convert to JSON
	jsonData, err := json.Marshal(response)
	if err != nil {
		return createErrorJSON(fmt.Sprintf("Failed to encode response: %v", err), 500), nil
	}
	
	return string(jsonData), nil
}

// Format streams for response
func formatStreams(streams map[string]*extractors.Stream) map[string]interface{} {
	formatted := make(map[string]interface{})
	
	for id, stream := range streams {
		formatted[id] = map[string]interface{}{
			"id":      stream.ID,
			"quality": stream.Quality,
			"size":    stream.Size,
			"ext":     stream.Ext,
		}
	}
	
	return formatted
}

// Extract URLs from stream parts
func extractURLs(stream *extractors.Stream) []string {
	urls := make([]string, 0)
	for _, part := range stream.Parts {
		urls = append(urls, part.URL)
	}
	return urls
}

// Create error JSON response
func createErrorJSON(message string, code int) string {
	response := map[string]interface{}{
		"status":  "error",
		"message": message,
		"code":    code,
	}
	
	jsonData, _ := json.Marshal(response)
	return string(jsonData)
}
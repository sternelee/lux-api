package handler

import (
	"encoding/json"
	"net/http"
	"net/url"

	"github.com/iawia002/lux/extractors"
)

// Handler is the main entry point for the Vercel serverless function
func Handler(w http.ResponseWriter, r *http.Request) {
	// Set response header to JSON
	w.Header().Set("Content-Type", "application/json")

	// Only allow GET requests
	if r.Method != http.MethodGet {
		respondWithError(w, http.StatusMethodNotAllowed, "Only GET requests are allowed")
		return
	}

	// Extract URL parameter
	urlParam := r.URL.Query().Get("url")
	if urlParam == "" {
		respondWithError(w, http.StatusBadRequest, "URL parameter is missing")
		return
	}

	format := r.URL.Query().Get("format")
    quality := r.URL.Query().Get("quality")

	// Validate URL
	_, err := url.ParseRequestURI(urlParam)
	if err != nil {
		respondWithError(w, http.StatusBadRequest, "Invalid URL provided")
		return
	}

	// Initialize options
	option := extractors.Options{}

	if format != "" {
        option.Format = format
    }
    if quality != "" {
        option.Quality = quality
    }

	// Extract data
	data, err := extractors.Extract(urlParam, option)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Failed to extract data: "+err.Error())
		return
	}

	// Respond with the extracted data as JSON
	response := map[string]interface{}{
		"status": "success",
		"data":   data,
	}

	if err := json.NewEncoder(w).Encode(response); err != nil {
		http.Error(w, "Failed to encode response", http.StatusInternalServerError)
	}
}

// respondWithError is a helper function to send error responses
func respondWithError(w http.ResponseWriter, statusCode int, message string) {
	w.WriteHeader(statusCode)
	response := map[string]string{
		"status":  "error",
		"message": message,
	}
	if err := json.NewEncoder(w).Encode(response); err != nil {
		http.Error(w, "Failed to encode error response", http.StatusInternalServerError)
	}
}

// TODO: Implement a function to limit the size of the extracted data if necessary
// func limitDataSize(data interface{}) interface{} {
//     // Implement logic to limit the size of the data
//     // This could involve truncating large text fields, limiting the number of items in arrays, etc.
//     return data
// }
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

	var urlParam string
	if r.Method == http.MethodGet {
        // Extract parameters from URL query
        urlParam = r.URL.Query().Get("url")
    } else if r.Method == http.MethodPost {
        // Extract parameters from POST body
        var params struct {
            URL     string `json:"url"`
        }
        err := json.NewDecoder(r.Body).Decode(&params)
        if err != nil {
            respondWithError(w, http.StatusBadRequest, "Invalid JSON body")
            return
        }
        urlParam = params.URL
    } else {
        respondWithError(w, http.StatusMethodNotAllowed, "Method not allowed")
        return
    }

	if urlParam == "" {
		respondWithError(w, http.StatusBadRequest, "URL parameter is missing")
		return
	}

	// Validate URL
	_, err := url.ParseRequestURI(urlParam)
	if err != nil {
		respondWithError(w, http.StatusBadRequest, "Invalid URL provided")
		return
	}

	// Initialize options
	option := extractors.Options{
		// Format: format,
		// Quality: quality,
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

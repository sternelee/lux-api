package handler

import (
	"encoding/json"
	"net/http"

	"github.com/iawia002/lux/extractors"
	"github.com/iawia002/lux/utils"
)

// Helper function for JSON responses
func writeJSON(w http.ResponseWriter, statusCode int, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)
	json.NewEncoder(w).Encode(data)
}

// Handler for the /api/extract endpoint
func Handler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		writeJSON(w, http.StatusMethodNotAllowed, map[string]string{"error": "Method not allowed"})
		return
	}

	if err := r.ParseForm(); err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "failed to parse form"})
		return
	}
	url := r.FormValue("url")
	text := r.FormValue("text")

	// If URL is provided, extract directly
	if url != "" {
		data, err := extractors.Extract(url, extractors.Options{})
		if err != nil {
			writeJSON(w, http.StatusInternalServerError, map[string]string{"error": err.Error()})
			return
		}
		writeJSON(w, http.StatusOK, data)
		return
	}

	// If no URL but text is provided, extract links from text
	if text != "" {
		links := utils.ExtractLinks(text)
		if len(links) == 0 {
			writeJSON(w, http.StatusBadRequest, map[string]string{"error": "no links found in text"})
			return
		}

		// Use the first link for extraction
		firstLink := links[0]
		data, err := extractors.Extract(firstLink, extractors.Options{})
		if err != nil {
			// If extraction fails, return the extracted links info
			writeJSON(w, http.StatusOK, map[string]interface{}{
				"links":      links,
				"first_link": firstLink,
				"error":      err.Error(),
			})
			return
		}

		// If extraction succeeds, return the extracted data and links info
		writeJSON(w, http.StatusOK, map[string]interface{}{
			"data":       data,
			"links":      links,
			"first_link": firstLink,
		})
		return
	}

	// If neither URL nor text is provided, return an error
	writeJSON(w, http.StatusBadRequest, map[string]string{"error": "either url or text parameter is required"})
}

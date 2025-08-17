package handler

import (
	"encoding/json"
	"net/http"
	"sync"

	"github.com/iawia002/lux/extractors"
	"github.com/iawia002/lux/extractors/bilibili"
	"github.com/iawia002/lux/extractors/douyin"
	"github.com/iawia002/lux/extractors/douyu"
	"github.com/iawia002/lux/extractors/eporner"
	"github.com/iawia002/lux/extractors/facebook"
	"github.com/iawia002/lux/extractors/haokan"
	"github.com/iawia002/lux/extractors/huya"
	"github.com/iawia002/lux/extractors/instagram"
	"github.com/iawia002/lux/extractors/kuaishou"
	"github.com/iawia002/lux/extractors/mgtv"
	"github.com/iawia002/lux/extractors/netease"
	"github.com/iawia002/lux/extractors/pinterest"
	"github.com/iawia002/lux/extractors/pixivision"
	"github.com/iawia002/lux/extractors/pornhub"
	"github.com/iawia002/lux/extractors/qq"
	"github.com/iawia002/lux/extractors/tiktok"
	"github.com/iawia002/lux/extractors/tumblr"
	"github.com/iawia002/lux/extractors/twitter"
	"github.com/iawia002/lux/extractors/universal"
	"github.com/iawia002/lux/extractors/weibo"
	"github.com/iawia002/lux/extractors/xiaohongshu"
	"github.com/iawia002/lux/extractors/xvideos"
	"github.com/iawia002/lux/extractors/youtube"
	"github.com/iawia002/lux/extractors/youku"
	"github.com/iawia002/lux/utils"
)

var once sync.Once

// Initialize extractors once
func initializeExtractors() {
	extractors.Register("bilibili", bilibili.New())
	extractors.Register("douyin", douyin.New())
	extractors.Register("douyu", douyu.New())
	extractors.Register("eporner", eporner.New())
	extractors.Register("facebook", facebook.New())
	extractors.Register("haokan", haokan.New())
	extractors.Register("huya", huya.New())
	extractors.Register("instagram", instagram.New())
	extractors.Register("kuaishou", kuaishou.New())
	extractors.Register("mgtv", mgtv.New())
	extractors.Register("netease", netease.New())
	extractors.Register("pinterest", pinterest.New())
	extractors.Register("pixivision", pixivision.New())
	extractors.Register("pornhub", pornhub.New())
	extractors.Register("qq", qq.New())
	extractors.Register("tiktok", tiktok.New())
	extractors.Register("tumblr", tumblr.New())
	extractors.Register("twitter", twitter.New())
	extractors.Register("universal", universal.New())
	extractors.Register("weibo", weibo.New())
	extractors.Register("xiaohongshu", xiaohongshu.New())
	extractors.Register("xvideos", xvideos.New())
	extractors.Register("youtube", youtube.New())
	extractors.Register("youku", youku.New())
}

// Helper function for JSON responses
func writeJSON(w http.ResponseWriter, statusCode int, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)
	json.NewEncoder(w).Encode(data)
}

// Handler for the /api/extract endpoint
func Handler(w http.ResponseWriter, r *http.Request) {
	// Initialize extractors once
	once.Do(initializeExtractors)
	
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

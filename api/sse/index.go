package handler

import (
	"context"
	"encoding/json"
	"net/http"

	"github.com/modelcontextprotocol/go-sdk/mcp"
	"github.com/iawia002/lux/extractors"
	"github.com/iawia002/lux/utils"
)

func initializeMCP() *mcp.Server {
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

	mcpServer := mcp.NewServer(&mcp.Implementation{Name: "lux-mcp-server", Version: "v1.0.0"}, nil)

	// 注册MCP工具
	mcp.AddTool(mcpServer, &mcp.Tool{Name: "extract_media", Description: "Extracts video or image data from a URL"}, ExtractMedia)
	mcp.AddTool(mcpServer, &mcp.Tool{Name: "extract_links", Description: "Extracts links from text content"}, ExtractLinks)

	return mcpServer
}

// ExtractMediaParams represents the parameters for the extract_media tool
type ExtractMediaParams struct {
	URL string `json:"url" jsonschema:"the URL of the page to extract media from"`
}

// ExtractLinksParams represents the parameters for the extract_links tool
type ExtractLinksParams struct {
	Text string `json:"text" jsonschema:"the text to extract links from"`
}

// ExtractMedia 工具函数
func ExtractMedia(ctx context.Context, req *mcp.ServerRequest[*mcp.CallToolParamsFor[ExtractMediaParams]]) (*mcp.CallToolResultFor[any], error) {
	data, err := extractors.Extract(req.Params.Arguments.URL, extractors.Options{})
	if err != nil {
		return &mcp.CallToolResultFor[any]{
			Content: []mcp.Content{&mcp.TextContent{Text: "Extraction failed: " + err.Error()}},
		}, nil
	}

	jsonData, err := json.MarshalIndent(data, "", "  ")
	if err != nil {
		return &mcp.CallToolResultFor[any]{
			Content: []mcp.Content{&mcp.TextContent{Text: "Failed to format result as JSON: " + err.Error()}},
		}, nil
	}

	return &mcp.CallToolResultFor[any]{
		Content: []mcp.Content{&mcp.TextContent{Text: string(jsonData)}},
	}, nil
}

// ExtractLinks 工具函数
func ExtractLinks(ctx context.Context, req *mcp.ServerRequest[*mcp.CallToolParamsFor[ExtractLinksParams]]) (*mcp.CallToolResultFor[any], error) {
	text := req.Params.Arguments.Text
	links := utils.ExtractLinks(text)
	firstLink := ""
	if len(links) > 0 {
		firstLink = links[0]
	}

	result := map[string]interface{}{
		"links":      links,
		"first_link": firstLink,
	}

	jsonData, err := json.MarshalIndent(result, "", "  ")
	if err != nil {
		return &mcp.CallToolResultFor[any]{
			Content: []mcp.Content{&mcp.TextContent{Text: "Failed to format result as JSON: " + err.Error()}},
		}, nil
	}

	return &mcp.CallToolResultFor[any]{
		Content: []mcp.Content{&mcp.TextContent{Text: string(jsonData)}},
	}, nil
}

// Handler for the /api/sse endpoint
func Handler(w http.ResponseWriter, r *http.Request) {
	mcpHandler := mcp.NewSSEHandler(func(request *http.Request) *mcp.Server {
		return initializeMCP()
	})
	mcpHandler.ServeHTTP(w, r)
}

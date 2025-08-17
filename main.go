package main

import (
	"context"
	"encoding/json"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/iawia002/lux/utils"
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
	"github.com/modelcontextprotocol/go-sdk/mcp"
	"github.com/syumai/workers"
)

type HiParams struct {
	Name string `json:"name" jsonschema:"the name of the person to greet"`
}

func SayHi(ctx context.Context, req *mcp.ServerRequest[*mcp.CallToolParamsFor[HiParams]]) (*mcp.CallToolResultFor[any], error) {
	return &mcp.CallToolResultFor[any]{
		Content: []mcp.Content{&mcp.TextContent{Text: "Hi " + req.Params.Arguments.Name}},
	}, nil
}

type ExtractMediaParams struct {
	URL string `json:"url" jsonschema:"the URL of the page to extract media from"`
}

type ExtractLinksParams struct {
	Text string `json:"text" jsonschema:"the text to extract links from"`
}

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

func main() {
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
	mcp.AddTool(mcpServer, &mcp.Tool{Name: "extract_media", Description: "Extracts video or image data from a URL"}, ExtractMedia)
	mcp.AddTool(mcpServer, &mcp.Tool{Name: "extract_links", Description: "Extracts links from text content"}, ExtractLinks)

	r := gin.Default()
	api := r.Group("/api")
	{
		api.POST("/extract", func(c *gin.Context) {
			url := c.PostForm("url")
			text := c.PostForm("text")
			
			// 如果提供了URL，直接使用URL进行提取
			if url != "" {
				data, err := extractors.Extract(url, extractors.Options{})
				if err != nil {
					c.JSON(http.StatusInternalServerError, gin.H{
						"error": err.Error(),
					})
					return
				}
				c.JSON(http.StatusOK, data)
				return
			}
			
			// 如果没有URL但有text，尝试从text中提取链接
			if text != "" {
				links := utils.ExtractLinks(text)
				if len(links) == 0 {
					c.JSON(http.StatusBadRequest, gin.H{
						"error": "no links found in text",
					})
					return
				}
				
				// 使用第一个链接进行提取
				firstLink := links[0]
				data, err := extractors.Extract(firstLink, extractors.Options{})
				if err != nil {
					// 如果提取失败，返回提取到的链接信息
					c.JSON(http.StatusOK, gin.H{
						"links":      links,
						"first_link": firstLink,
						"error":      err.Error(),
					})
					return
				}
				
				// 提取成功，返回提取结果和链接信息
				c.JSON(http.StatusOK, gin.H{
					"data":       data,
					"links":      links,
					"first_link": firstLink,
				})
				return
			}
			
			// 如果既没有URL也没有text，返回错误
			c.JSON(http.StatusBadRequest, gin.H{
				"error": "either url or text parameter is required",
			})
		})

		handler := mcp.NewSSEHandler(func(request *http.Request) *mcp.Server {
			return mcpServer
		})
		mcpRoute := "/sse/*service"
		api.GET(mcpRoute, gin.WrapH(handler))
		api.POST(mcpRoute, gin.WrapH(handler))
	}

	workers.Serve(r)
}

package main

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-lambda-go/lambda"
	"github.com/iawia002/lux/extractors"
	"github.com/iawia002/lux/request"
)

// APIResponse 定义API响应结构
type APIResponse struct {
	Success bool        `json:"success"`
	Data    interface{} `json:"data,omitempty"`
	Error   string      `json:"error,omitempty"`
}

// ExtractRequest 定义提取请求结构
type ExtractRequest struct {
	URL              string `json:"url"`
	Cookie           string `json:"cookie,omitempty"`
	UserAgent        string `json:"user_agent,omitempty"`
	Refer            string `json:"refer,omitempty"`
	Playlist         bool   `json:"playlist,omitempty"`
	Items            string `json:"items,omitempty"`
	ItemStart        int    `json:"item_start,omitempty"`
	ItemEnd          int    `json:"item_end,omitempty"`
	ThreadNumber     int    `json:"thread_number,omitempty"`
	EpisodeTitleOnly bool   `json:"episode_title_only,omitempty"`
	YoukuCcode       string `json:"youku_ccode,omitempty"`
	YoukuCkey        string `json:"youku_ckey,omitempty"`
	YoukuPassword    string `json:"youku_password,omitempty"`
}

// handler 处理API请求
func handler(ctx context.Context, request events.APIGatewayProxyRequest) (events.APIGatewayProxyResponse, error) {
	// 设置CORS头
	headers := map[string]string{
		"Content-Type":                     "application/json",
		"Access-Control-Allow-Origin":      "*",
		"Access-Control-Allow-Headers":     "Content-Type,X-Amz-Date,Authorization,X-Api-Key,X-Amz-Security-Token",
		"Access-Control-Allow-Methods":     "GET,POST,OPTIONS",
		"Access-Control-Allow-Credentials": "true",
	}

	// 处理OPTIONS请求（CORS预检）
	if request.HTTPMethod == "OPTIONS" {
		return events.APIGatewayProxyResponse{
			StatusCode: 200,
			Headers:    headers,
			Body:       "",
		}, nil
	}

	var response APIResponse

	// 根据HTTP方法处理请求
	switch request.HTTPMethod {
	case "GET":
		// GET请求：从查询参数中获取URL
		url := request.QueryStringParameters["url"]
		if url == "" {
			response = APIResponse{
				Success: false,
				Error:   "URL parameter is required",
			}
		} else {
			response = extractVideoData(url, ExtractRequest{URL: url})
		}

	case "POST":
		// POST请求：从请求体中解析JSON
		var extractReq ExtractRequest
		if err := json.Unmarshal([]byte(request.Body), &extractReq); err != nil {
			response = APIResponse{
				Success: false,
				Error:   fmt.Sprintf("Invalid JSON: %v", err),
			}
		} else if extractReq.URL == "" {
			response = APIResponse{
				Success: false,
				Error:   "URL is required in request body",
			}
		} else {
			response = extractVideoData(extractReq.URL, extractReq)
		}

	default:
		response = APIResponse{
			Success: false,
			Error:   "Method not allowed. Use GET or POST",
		}
	}

	// 序列化响应
	responseBody, err := json.Marshal(response)
	if err != nil {
		return events.APIGatewayProxyResponse{
			StatusCode: 500,
			Headers:    headers,
			Body:       `{"success":false,"error":"Internal server error"}`,
		}, nil
	}

	statusCode := 200
	if !response.Success {
		statusCode = 400
	}

	return events.APIGatewayProxyResponse{
		StatusCode: statusCode,
		Headers:    headers,
		Body:       string(responseBody),
	}, nil
}

// extractVideoData 提取视频数据
func extractVideoData(url string, req ExtractRequest) APIResponse {
	// 设置请求选项
	request.SetOptions(request.Options{
		RetryTimes: 10,
		Cookie:     req.Cookie,
		UserAgent:  req.UserAgent,
		Refer:      req.Refer,
		Debug:      false,
		Silent:     true,
	})

	// 设置提取选项
	options := extractors.Options{
		Playlist:         req.Playlist,
		Items:            req.Items,
		ItemStart:        req.ItemStart,
		ItemEnd:          req.ItemEnd,
		ThreadNumber:     req.ThreadNumber,
		EpisodeTitleOnly: req.EpisodeTitleOnly,
		Cookie:           req.Cookie,
		YoukuCcode:       req.YoukuCcode,
		YoukuCkey:        req.YoukuCkey,
		YoukuPassword:    req.YoukuPassword,
	}

	// 提取数据
	data, err := extractors.Extract(url, options)
	if err != nil {
		return APIResponse{
			Success: false,
			Error:   fmt.Sprintf("Failed to extract data: %v", err),
		}
	}

	// 处理提取结果
	var results []interface{}
	for _, item := range data {
		if item.Err != nil {
			results = append(results, map[string]interface{}{
				"url":   item.URL,
				"error": item.Err.Error(),
			})
		} else {
			// 清理数据，移除不能序列化的字段
			cleanData := map[string]interface{}{
				"url":     item.URL,
				"site":    item.Site,
				"title":   item.Title,
				"type":    item.Type,
				"streams": item.Streams,
			}

			// 添加字幕信息
			if item.Captions != nil {
				captions := make(map[string]interface{})
				for key, caption := range item.Captions {
					captions[key] = map[string]interface{}{
						"url":  caption.URL,
						"size": caption.Size,
						"ext":  caption.Ext,
					}
				}
				cleanData["captions"] = captions
			}

			results = append(results, cleanData)
		}
	}

	return APIResponse{
		Success: true,
		Data:    results,
	}
}

func main() {
	lambda.Start(handler)
}

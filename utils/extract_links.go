package utils

import (
	"regexp"
	"strings"
)

// ExtractLinks 从文本中提取所有URL链接
func ExtractLinks(text string) []string {
	// 匹配URL的正则表达式
	// 支持http, https开头的URL
	// 也支持www.开头的URL
	urlRegex := regexp.MustCompile(`(https?:\/\/[^\s]+)|(www\.[^\s]+)`)
	
	// 查找所有匹配项
	matches := urlRegex.FindAllString(text, -1)
	
	// 如果没有找到匹配项，返回空切片
	if matches == nil {
		return []string{}
	}
	
	// 清理链接（去除可能的结尾标点符号等）
	var cleanLinks []string
	for _, link := range matches {
		// 去除链接末尾可能的标点符号
		link = strings.TrimRight(link, ",.!?;:'\"")
		cleanLinks = append(cleanLinks, link)
	}
	
	return cleanLinks
}

// ExtractFirstLink 从文本中提取第一个URL链接
func ExtractFirstLink(text string) string {
	links := ExtractLinks(text)
	if len(links) > 0 {
		return links[0]
	}
	return ""
}
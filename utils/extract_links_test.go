package utils

import (
	"reflect"
	"testing"
)

func TestExtractLinks(t *testing.T) {
	tests := []struct {
		name     string
		text     string
		expected []string
	}{
		{
			name:     "小红书分享文本",
			text:     "67 我发现了一篇小红书笔记，快来看吧 😆 3h8XKv3d89H 😆 http://xhslink.com/m/3Bt1MNtLrzm ，复制本条信息，打开【小红书】App查看精彩内容！",
			expected: []string{"http://xhslink.com/m/3Bt1MNtLrzm"},
		},
		{
			name:     "多个链接",
			text:     "这里有两个链接 https://example.com 和 http://test.org",
			expected: []string{"https://example.com", "http://test.org"},
		},
		{
			name:     "带www的链接",
			text:     "访问 www.example.com 获取更多信息",
			expected: []string{"www.example.com"},
		},
		{
			name:     "没有链接",
			text:     "这段文本中没有任何链接",
			expected: []string{},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := ExtractLinks(tt.text)
			if !reflect.DeepEqual(got, tt.expected) {
				t.Errorf("ExtractLinks() = %v, want %v", got, tt.expected)
			}
		})
	}
}

func TestExtractFirstLink(t *testing.T) {
	tests := []struct {
		name     string
		text     string
		expected string
	}{
		{
			name:     "小红书分享文本",
			text:     "67 我发现了一篇小红书笔记，快来看吧 😆 3h8XKv3d89H 😆 http://xhslink.com/m/3Bt1MNtLrzm ，复制本条信息，打开【小红书】App查看精彩内容！",
			expected: "http://xhslink.com/m/3Bt1MNtLrzm",
		},
		{
			name:     "多个链接",
			text:     "这里有两个链接 https://example.com 和 http://test.org",
			expected: "https://example.com",
		},
		{
			name:     "没有链接",
			text:     "这段文本中没有任何链接",
			expected: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := ExtractFirstLink(tt.text)
			if got != tt.expected {
				t.Errorf("ExtractFirstLink() = %v, want %v", got, tt.expected)
			}
		})
	}
}
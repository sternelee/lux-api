# Lux API Service

这是一个基于SST平台部署的Lux视频提取API服务，提供了REST API接口来提取各种平台的视频信息。

## 功能特性

- 支持多种视频平台的视频信息提取
- 提供REST API接口
- 支持CORS跨域请求
- 支持GET和POST请求方式
- 返回JSON格式的响应

## 部署

### 前置要求

1. 安装SST CLI
```bash
npm install -g sst
```

2. 配置AWS凭证
```bash
aws configure
```

### 构建和部署

1. 构建Go二进制文件
```bash
chmod +x build.sh
./build.sh
```

2. 部署到AWS
```bash
sst deploy
```

3. 获取API URL
```bash
sst output
```

## API使用说明

### 基础URL

部署完成后，你会得到一个API Gateway URL，格式类似：
```
https://xxxxxxxxxx.execute-api.region.amazonaws.com
```

### 端点

#### 1. GET /extract

通过查询参数提取视频信息

**请求示例：**
```bash
curl "https://your-api-gateway-url/extract?url=https://www.youtube.com/watch?v=dQw4w9WgXcQ"
```

**参数：**
- `url` (必需): 要提取的视频URL

#### 2. POST /extract

通过JSON请求体提取视频信息

**请求示例：**
```bash
curl -X POST "https://your-api-gateway-url/extract" \
  -H "Content-Type: application/json" \
  -d '{
    "url": "https://www.youtube.com/watch?v=dQw4w9WgXcQ",
    "cookie": "your-cookie-string",
    "user_agent": "Mozilla/5.0...",
    "playlist": false
  }'
```

**请求体参数：**
- `url` (必需): 要提取的视频URL
- `cookie` (可选): Cookie字符串
- `user_agent` (可选): 用户代理字符串
- `refer` (可选): Referer头
- `playlist` (可选): 是否提取播放列表，默认false
- `items` (可选): 播放列表项目，格式如"1,5,6,8-10"
- `item_start` (可选): 播放列表起始项目
- `item_end` (可选): 播放列表结束项目
- `thread_number` (可选): 线程数
- `episode_title_only` (可选): 是否只包含剧集标题
- `youku_ccode` (可选): 优酷ccode
- `youku_ckey` (可选): 优酷ckey
- `youku_password` (可选): 优酷密码

### 响应格式

#### 成功响应
```json
{
  "success": true,
  "data": [
    {
      "url": "https://www.youtube.com/watch?v=dQw4w9WgXcQ",
      "site": "youtube",
      "title": "Rick Astley - Never Gonna Give You Up",
      "type": "video",
      "streams": {
        "720p": {
          "id": "720p",
          "quality": "720p",
          "parts": [
            {
              "url": "https://...",
              "size": 12345678,
              "ext": "mp4"
            }
          ],
          "size": 12345678,
          "ext": "mp4"
        }
      },
      "captions": {
        "en": {
          "url": "https://...",
          "size": 1234,
          "ext": "vtt"
        }
      }
    }
  ]
}
```

#### 错误响应
```json
{
  "success": false,
  "error": "Failed to extract data: invalid URL"
}
```

## 支持的平台

Lux支持多种视频平台，包括但不限于：

- YouTube
- Bilibili
- TikTok
- Instagram
- Twitter
- Facebook
- 优酷
- 爱奇艺
- 腾讯视频
- 等等

## 开发

### 本地开发

1. 安装依赖
```bash
go mod tidy
```

2. 运行本地开发服务器
```bash
sst dev
```

### 测试

```bash
# 测试GET请求
curl "http://localhost:3000/extract?url=https://www.youtube.com/watch?v=dQw4w9WgXcQ"

# 测试POST请求
curl -X POST "http://localhost:3000/extract" \
  -H "Content-Type: application/json" \
  -d '{"url": "https://www.youtube.com/watch?v=dQw4w9WgXcQ"}'
```

## 注意事项

1. Lambda函数有30秒的超时限制
2. 内存限制为1024MB
3. 支持ARM64架构以获得更好的性能
4. 所有请求都支持CORS
5. 建议在生产环境中设置适当的CORS策略

## 故障排除

### 常见问题

1. **超时错误**: 某些视频可能需要更长的处理时间，考虑增加Lambda超时时间
2. **内存不足**: 对于大型视频，可能需要增加Lambda内存限制
3. **CORS错误**: 确保客户端正确设置了CORS头

### 日志查看

```bash
sst logs
```

## 许可证

本项目基于MIT许可证开源。

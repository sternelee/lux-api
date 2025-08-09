# Lux API for Cloudflare Workers (WASM)

This project provides a Cloudflare Workers API for the [lux](https://github.com/iawia002/lux) video extraction library, compiled to WebAssembly.

## 📋 Prerequisites

- Go 1.19+ (for building WASM)
- Node.js 16+ and npm
- Wrangler CLI (`npm install -g wrangler`)
- Optional: wasm-opt for optimization (`npm install -g wasm-opt`)

## 🏗️ Building the WASM Module

1. **Install Go dependencies:**
   ```bash
   go mod download
   ```

2. **Build the WASM module:**
   ```bash
   ./build-wasm.sh
   ```
   
   This will create a `worker.wasm` file with the compiled lux functionality.

## 🚀 Deployment

### Local Development

1. **Start the development server:**
   ```bash
   wrangler dev
   ```

2. **Test the API:**
   ```bash
   # Health check
   curl http://localhost:8787/api/health
   
   # Extract video info
   curl "http://localhost:8787/api/extract?url=https://www.youtube.com/watch?v=VIDEO_ID"
   ```

### Production Deployment

1. **Login to Cloudflare:**
   ```bash
   wrangler login
   ```

2. **Deploy to Cloudflare Workers:**
   ```bash
   wrangler publish
   ```

## 🔧 Troubleshooting

### Error: "WebAssembly.instantiate(): Import #0 "gojs": module is not an object or function"

This error occurs when the Go WASM runtime is not properly initialized. The solution implemented in this project:

1. **Custom WASM Loader (`wasm-loader.js`):** Provides all necessary Go runtime functions that Cloudflare Workers supports.

2. **Go WASM Implementation (`main.wasm.go`):** Uses `syscall/js` to properly interface with JavaScript, including Promise support for async operations.

3. **Proper Import Object:** The `importObject` in `wasm-loader.js` includes all required `gojs` runtime functions.

### Common Issues and Solutions

1. **WASM module too large:**
   - Use build flags: `-ldflags="-s -w"` to strip debug info
   - Run `wasm-opt` for optimization
   - Consider splitting functionality into multiple workers

2. **Network requests fail in WASM:**
   - Cloudflare Workers has limited syscall support
   - Network requests should be handled through JavaScript fetch API
   - Pass results back to WASM for processing

3. **Memory issues:**
   - Cloudflare Workers has memory limits
   - Monitor memory usage and implement cleanup
   - Use streaming for large data when possible

## 📚 API Documentation

### GET /api/health
Health check endpoint

**Response:**
```json
{
  "status": "ok",
  "service": "lux-api",
  "version": "1.0.0",
  "wasm_loaded": true
}
```

### GET/POST /api/extract
Extract video information

**Parameters:**
- `url` (required): Video URL to extract

**Response:**
```json
{
  "status": "success",
  "data": {
    "url": "https://...",
    "site": "YouTube",
    "title": "Video Title",
    "type": "video",
    "streams": {
      "1080p": {
        "id": "1080p",
        "quality": "1080p",
        "size": 104857600,
        "ext": "mp4"
      }
    }
  },
  "metadata": {
    "extracted_at": "2024-01-01T00:00:00.000Z",
    "method": "wasm"
  }
}
```

## 🏗️ Architecture

```
┌─────────────────────────────────────┐
│     Cloudflare Workers Runtime      │
├─────────────────────────────────────┤
│         index.js (Entry)            │
│              ↓                      │
│       wasm-loader.js                │
│   (Go Runtime Implementation)       │
│              ↓                      │
│        worker.wasm                  │
│    (Compiled Go Lux Library)        │
└─────────────────────────────────────┘
```

## 🤝 Contributing

1. Fork the repository
2. Create your feature branch
3. Commit your changes
4. Push to the branch
5. Create a Pull Request

## 📄 License

This project is licensed under the MIT License - see the LICENSE file for details.

## 🙏 Acknowledgments

- [lux](https://github.com/iawia002/lux) - The video extraction library
- [Cloudflare Workers](https://workers.cloudflare.com/) - Edge computing platform
- Go WebAssembly support
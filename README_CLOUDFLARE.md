# Lux API - Cloudflare Workers

This project converts the Lux video downloader service into a Cloudflare Workers API with WebAssembly support.

## Features

- **API Routes**: RESTful API endpoints for video extraction
- **WebAssembly Support**: Go code compiled to WASM for optimal performance
- **Cloudflare Workers**: Serverless deployment with global edge network
- **CORS Support**: Cross-origin requests for web applications
- **Health Check**: Monitoring endpoint for service status

## API Endpoints

### GET /api/health

Health check endpoint to verify the service is running.

**Response:**

```json
{
  "status": "ok",
  "service": "lux-api",
  "version": "1.0.0",
  "timestamp": "2023-12-01T00:00:00.000Z"
}
```

### GET /api/extract?url=<VIDEO_URL>

Extract video information from a URL.

**Parameters:**

- `url` (required): The video URL to extract information from

**Response:**

```json
{
  "status": "success",
  "data": {
    "url": "https://example.com/video",
    "title": "Video Title",
    "streams": {...}
  }
}
```

### POST /api/extract

Extract video information using POST request.

**Request Body:**

```json
{
  "url": "https://example.com/video"
}
```

**Response:** Same as GET endpoint

## Setup

### Prerequisites

- Node.js (v16 or later)
- Go (v1.18 or later)
- Wrangler CLI (`npm install -g wrangler`)
- Cloudflare account

### Installation

1. **Clone the repository:**

   ```bash
   git clone <repository-url>
   cd lux-api
   ```

2. **Install dependencies:**

   ```bash
   go mod tidy
   ```

3. **Build WASM module:**

   ```bash
   make build-wasm
   ```

4. **Login to Cloudflare:**

   ```bash
   wrangler login
   ```

5. **Configure your worker:**

   ```bash
   wrangler config
   ```

6. **Deploy to Cloudflare Workers:**
   ```bash
   wrangler deploy
   ```

## Development

### Local Development

1. **Start local development server:**

   ```bash
   wrangler dev
   ```

2. **Test the API:**
   ```bash
   curl http://localhost:8787/api/health
   curl "http://localhost:8787/api/extract?url=https://example.com/video"
   ```

### Testing

Open `test.html` in your browser to test the API endpoints:

```bash
open test.html
```

### Build Commands

- `make build`: Build standard Go binary
- `make build-wasm`: Build WASM for Cloudflare Workers
- `make clean`: Clean build artifacts
- `make test`: Run tests

## Configuration

### Environment Variables

Configure in `wrangler.toml`:

```toml
[env.production]
vars = { ENVIRONMENT = "production" }

[env.staging]
vars = { ENVIRONMENT = "staging" }
```

### Custom Domains

Add custom domain in Cloudflare dashboard or use:

```bash
wrangler routes list
wrangler routes add <your-domain.com>/* lux-api.<your-workers-subdomain>.workers.dev
```

## API Usage Examples

### JavaScript

```javascript
// GET request
const response = await fetch(
  "https://your-worker.workers.dev/api/extract?url=https://youtube.com/watch?v=VIDEO_ID",
);
const data = await response.json();

// POST request
const response = await fetch("https://your-worker.workers.dev/api/extract", {
  method: "POST",
  headers: { "Content-Type": "application/json" },
  body: JSON.stringify({ url: "https://youtube.com/watch?v=VIDEO_ID" }),
});
const data = await response.json();
```

### curl

```bash
# Health check
curl https://your-worker.workers.dev/api/health

# GET request
curl "https://your-worker.workers.dev/api/extract?url=https://youtube.com/watch?v=VIDEO_ID"

# POST request
curl -X POST https://your-worker.workers.dev/api/extract \
  -H "Content-Type: application/json" \
  -d '{"url": "https://youtube.com/watch?v=VIDEO_ID"}'
```

## Troubleshooting

### Common Issues

1. **WASM Build Failures:**
   - Ensure Go version is 1.18 or later
   - Check for missing dependencies with `go mod tidy`

2. **Deployment Issues:**
   - Verify Cloudflare authentication with `wrangler whoami`
   - Check worker name conflicts in Cloudflare dashboard

3. **CORS Issues:**
   - The API includes CORS headers for cross-origin requests
   - Verify your client-side code includes proper headers

### Debug Mode

Enable debug logging by setting environment variables:

```bash
wrangler secret put DEBUG_ENABLED
# Enter: true
```

## License

This project inherits the license from the original Lux project.

## Contributing

1. Fork the repository
2. Create a feature branch
3. Make your changes
4. Test thoroughly
5. Submit a pull request

## Support

For issues and questions:

- Check the original Lux project documentation
- Open an issue in the repository
- Review Cloudflare Workers documentation


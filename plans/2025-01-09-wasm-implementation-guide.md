# Go WASM Integration Implementation Guide for Cloudflare Workers

## Current State Analysis

Based on the repository analysis, the project has:
- A working Cloudflare Workers setup with `index.js` as the main entry point
- Go source files ready for WASM compilation (`worker.go`, `main.wasm.go`, `api/wasm.go`)
- A minimal `wasm_exec.js` stub that lacks the complete Go runtime
- A Makefile with WASM build targets that creates an insufficient runtime file

## Root Cause of the Error

The error "WebAssembly.instantiate(): Import #0 'gojs': module is not an object or function" occurs because:

1. Go-compiled WASM modules require a specific JavaScript runtime environment
2. The current `wasm_exec.js` is a minimal stub that doesn't provide the "gojs" module
3. The WASM instantiation in `index.js` doesn't follow the Go WASM loading pattern

## Step-by-Step Implementation Guide

### Step 1: Obtain the Official wasm_exec.js

**Action Required**: Copy the official `wasm_exec.js` from your Go installation.

**Command to find and copy**:
```bash
# Find your Go installation
go env GOROOT

# Copy the official wasm_exec.js
cp "$(go env GOROOT)/misc/wasm/wasm_exec.js" ./wasm_exec_official.js
```

**Important Notes**:
- The official file is typically around 600-700 lines
- It includes the complete Go runtime for WebAssembly
- Must match your Go version exactly

### Step 2: Update the Makefile

**Modifications needed in Makefile**:

Replace the current `build-wasm` target with:
```makefile
build-wasm:
	@echo "Building WASM module..."
	GOOS=js GOARCH=wasm go build -o worker.wasm worker.go
	@echo "Copying official wasm_exec.js..."
	@if [ -f "$(shell go env GOROOT)/misc/wasm/wasm_exec.js" ]; then \
		cp "$(shell go env GOROOT)/misc/wasm/wasm_exec.js" ./wasm_exec.js; \
		echo "Successfully copied wasm_exec.js from Go distribution"; \
	else \
		echo "Error: Could not find wasm_exec.js in Go distribution"; \
		exit 1; \
	fi
	@echo "WASM build complete!"
```

### Step 3: Create a Go Runtime Loader

**Create a new file**: `go-wasm-loader.js`

**Purpose**: This file will properly initialize the Go runtime for Cloudflare Workers.

**Key Components**:
```javascript
// go-wasm-loader.js
import './wasm_exec.js';

export async function loadWasm(wasmModule) {
    // Create Go runtime instance
    const go = new Go();
    
    // Polyfill for Cloudflare Workers environment
    if (!globalThis.crypto) {
        globalThis.crypto = {
            getRandomValues: (arr) => {
                for (let i = 0; i < arr.length; i++) {
                    arr[i] = Math.floor(Math.random() * 256);
                }
            }
        };
    }
    
    // Instantiate the WASM module with Go's import object
    const instance = await WebAssembly.instantiate(wasmModule, go.importObject);
    
    // Run the Go program
    go.run(instance);
    
    // Return the instance for accessing exports
    return instance;
}
```

### Step 4: Update index.js

**Modifications needed**:

1. Import the Go loader instead of direct WASM import
2. Initialize Go runtime before using WASM functions
3. Handle the asynchronous initialization properly

**Key changes**:
```javascript
// At the top of index.js
import { loadWasm } from './go-wasm-loader.js';
import wasmModule from './worker.wasm';

// Global variable for WASM instance
let wasmInstance = null;
let isWasmReady = false;

// Initialize WASM on first request
async function ensureWasmLoaded() {
    if (!isWasmReady) {
        try {
            wasmInstance = await loadWasm(wasmModule);
            isWasmReady = true;
            console.log('WASM module loaded successfully');
        } catch (error) {
            console.error('Failed to load WASM:', error);
            throw error;
        }
    }
}

// In the fetch handler
export default {
    async fetch(request, env, ctx) {
        // Ensure WASM is loaded
        await ensureWasmLoaded();
        
        // ... rest of the handler code
    }
}
```

### Step 5: Update worker.go for Cloudflare Compatibility

**Key considerations**:
- Cloudflare Workers has limited syscall support
- Network operations must go through JavaScript fetch
- No filesystem access

**Recommended modifications**:
1. Remove or stub out filesystem operations
2. Use JavaScript fetch for HTTP requests
3. Implement timeout handling
4. Add proper error boundaries

### Step 6: Configure wrangler.toml

**Update wrangler.toml**:
```toml
name = "lux-api"
main = "index.js"
compatibility_date = "2023-12-01"

[build]
command = "make build-wasm"

[rules]
[[rules]]
type = "CompiledWasm"
globs = ["**/*.wasm"]
fallthrough = true

[env.production]
vars = { ENVIRONMENT = "production" }

[env.staging]
vars = { ENVIRONMENT = "staging" }
```

### Step 7: Create Compatibility Shims

**Create**: `worker-shims.js`

**Purpose**: Provide polyfills for Go runtime features not available in Workers.

**Key shims needed**:
- Performance timing APIs
- Process-related globals
- Filesystem stubs
- Network request proxying

### Step 8: Build Verification Script

**Create**: `scripts/verify-wasm-build.sh`

```bash
#!/bin/bash

echo "Verifying WASM build..."

# Check if worker.wasm exists
if [ ! -f "worker.wasm" ]; then
    echo "Error: worker.wasm not found"
    exit 1
fi

# Check WASM file size
SIZE=$(stat -f%z "worker.wasm" 2>/dev/null || stat -c%s "worker.wasm" 2>/dev/null)
echo "WASM size: $SIZE bytes"

if [ $SIZE -gt 5242880 ]; then
    echo "Warning: WASM file exceeds 5MB (Cloudflare Workers limit)"
fi

# Check if wasm_exec.js exists and is complete
if [ ! -f "wasm_exec.js" ]; then
    echo "Error: wasm_exec.js not found"
    exit 1
fi

# Verify wasm_exec.js has the Go runtime
if ! grep -q "class Go" wasm_exec.js; then
    echo "Error: wasm_exec.js doesn't contain Go runtime"
    exit 1
fi

echo "Build verification passed!"
```

## Testing Strategy

### Local Testing
1. Run `make build-wasm` to compile
2. Use `wrangler dev` to test locally
3. Test each endpoint with curl or the browser

### Test Cases
1. Health check endpoint
2. Video extraction with valid URL
3. Error handling with invalid URL
4. WASM fallback to simulation mode
5. Memory usage under load

## Deployment Checklist

- [ ] Go version verified and consistent
- [ ] Official wasm_exec.js copied
- [ ] WASM module compiles without errors
- [ ] File size under Cloudflare limits (5MB for Workers)
- [ ] Local testing passes all endpoints
- [ ] Error handling implemented
- [ ] Logging configured for debugging
- [ ] Performance metrics acceptable

## Common Issues and Solutions

### Issue 1: "gojs" module not found
**Solution**: Ensure you're using the official wasm_exec.js from your Go distribution

### Issue 2: WASM file too large
**Solution**: 
- Use build flags: `-ldflags="-s -w"`
- Consider TinyGo as alternative
- Split functionality into smaller modules

### Issue 3: Runtime panics in Workers
**Solution**: 
- Implement proper error boundaries
- Add recovery mechanisms
- Use defensive programming for syscalls

### Issue 4: Network requests fail from WASM
**Solution**: 
- Proxy requests through JavaScript layer
- Use fetch API instead of Go's net/http
- Implement request queuing and retry logic

## Performance Optimization

1. **Lazy Loading**: Load WASM only when needed
2. **Caching**: Cache WASM instance between requests
3. **Memory Management**: Implement proper cleanup
4. **Request Batching**: Group multiple operations
5. **Compression**: Use wasm-opt for size reduction

## Monitoring and Debugging

### Add Logging Points
- WASM initialization start/end
- Function call entry/exit
- Error conditions
- Memory usage checkpoints

### Metrics to Track
- WASM load time
- Function execution time
- Memory consumption
- Error rates
- Fallback usage

## Next Steps After Implementation

1. **Validate Core Functionality**
   - Test with various video URLs
   - Verify all extractors work
   - Check error handling

2. **Performance Testing**
   - Measure response times
   - Check memory usage
   - Test concurrent requests

3. **Production Readiness**
   - Set up monitoring
   - Configure alerts
   - Document API changes

4. **Consider Alternatives if Issues Persist**
   - TinyGo compilation
   - Hybrid JavaScript/Go approach
   - External service architecture
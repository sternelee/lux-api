# Troubleshooting Guide: Go WASM "gojs" Module Error in Cloudflare Workers

## Error Details

**Error Message**: `service core:user:lux-api: Uncaught TypeError: WebAssembly.instantiate(): Import #0 "gojs": module is not an object or function`

**Error Location**: Cloudflare Workers runtime during WASM instantiation

**Error Type**: WebAssembly Import Error

## Immediate Solutions

### Solution 1: Quick Fix with Official wasm_exec.js

**Steps**:
1. Get your Go version:
   ```bash
   go version
   ```

2. Locate and copy the official wasm_exec.js:
   ```bash
   # For macOS/Linux
   cp $(go env GOROOT)/misc/wasm/wasm_exec.js ./wasm_exec.js
   
   # Verify it's the complete file (should be 600+ lines)
   wc -l wasm_exec.js
   ```

3. Rebuild the WASM module:
   ```bash
   GOOS=js GOARCH=wasm go build -o worker.wasm worker.go
   ```

4. Update index.js to properly load Go runtime:
   ```javascript
   // Add this at the top of index.js
   import './wasm_exec.js';
   
   // Before loading WASM
   const go = new Go();
   
   // Load WASM with Go's importObject
   const wasmInstance = await WebAssembly.instantiate(wasmModule, go.importObject);
   
   // Run the Go program
   go.run(wasmInstance);
   ```

5. Deploy and test:
   ```bash
   wrangler deploy
   ```

### Solution 2: Cloudflare Workers-Specific wasm_exec.js

**Create a modified wasm_exec.js for Cloudflare Workers**:

```javascript
// wasm_exec_cloudflare.js
// This is a template - you need to merge with official wasm_exec.js

(() => {
    // Polyfills for Cloudflare Workers
    if (typeof global === 'undefined') {
        window.global = window;
    }
    
    if (!global.fs) {
        global.fs = {
            writeSync: () => {},
            write: (fd, buf, offset, length, position, callback) => {
                if (fd === 1 || fd === 2) {
                    const text = new TextDecoder("utf-8").decode(buf);
                    console.log(text);
                }
                callback(null, length);
            },
            read: () => {},
            open: () => {},
            close: () => {},
            fstat: () => {}
        };
    }
    
    if (!global.process) {
        global.process = {
            getuid: () => -1,
            getgid: () => -1,
            geteuid: () => -1,
            getegid: () => -1,
            getgroups: () => [],
            pid: -1,
            ppid: -1,
            umask: () => {},
            cwd: () => "/",
            chdir: () => {}
        };
    }
    
    if (!global.crypto) {
        global.crypto = {
            getRandomValues: (arr) => {
                for (let i = 0; i < arr.length; i++) {
                    arr[i] = Math.floor(Math.random() * 256);
                }
            }
        };
    }
    
    // Include the official Go runtime here
    // Copy the Go class definition from official wasm_exec.js
    
    class Go {
        constructor() {
            this.importObject = {
                gojs: {
                    // Go runtime functions
                    // These MUST match what your Go version expects
                    "runtime.wasmExit": (code) => {
                        console.log("Go program exited with code:", code);
                    },
                    "runtime.wasmWrite": (fd, p, n) => {
                        // Handle stdout/stderr
                        return n;
                    },
                    "runtime.nanotime1": () => {
                        return Date.now() * 1000000;
                    },
                    "runtime.walltime": () => {
                        const msec = Date.now();
                        const sec = Math.floor(msec / 1000);
                        const nsec = (msec % 1000) * 1000000;
                        return [sec, nsec];
                    },
                    "runtime.scheduleTimeoutEvent": (delay) => {
                        // Schedule timeout
                        return 0;
                    },
                    "runtime.clearTimeoutEvent": (id) => {
                        // Clear timeout
                    },
                    "runtime.getRandomData": (r) => {
                        crypto.getRandomValues(r);
                    }
                },
                // Add other required imports
                env: {
                    // Environment imports if needed
                }
            };
        }
        
        async run(instance) {
            // Initialize and run the Go program
            this._inst = instance;
            // Additional initialization code from official wasm_exec.js
        }
    }
    
    globalThis.Go = Go;
})();
```

### Solution 3: Use TinyGo Instead

**TinyGo produces smaller WASM files and has better Cloudflare Workers compatibility**:

1. Install TinyGo:
   ```bash
   # macOS
   brew tap tinygo-org/tools
   brew install tinygo
   
   # Or download from https://tinygo.org/getting-started/install/
   ```

2. Build with TinyGo:
   ```bash
   tinygo build -o worker.wasm -target wasm -no-debug worker.go
   ```

3. Use TinyGo's wasm_exec.js:
   ```bash
   cp $(tinygo env TINYGOROOT)/targets/wasm_exec.js ./wasm_exec.js
   ```

4. TinyGo-specific worker.go modifications:
   ```go
   //go:build tinygo.wasm
   
   package main
   
   import "syscall/js"
   
   func main() {
       js.Global().Set("luxInfo", js.FuncOf(luxInfo))
       select {} // Use select instead of channel for TinyGo
   }
   ```

### Solution 4: Hybrid Approach - JavaScript Wrapper

**If WASM issues persist, create a JavaScript wrapper**:

```javascript
// lux-wrapper.js
// Use JavaScript for Cloudflare Workers compatibility
// Call Go WASM only for core logic

class LuxWrapper {
    constructor() {
        this.wasmReady = false;
        this.wasmInstance = null;
    }
    
    async initialize() {
        try {
            // Try to load WASM
            const go = new Go();
            const wasmModule = await WebAssembly.compileStreaming(fetch('./worker.wasm'));
            this.wasmInstance = await WebAssembly.instantiate(wasmModule, go.importObject);
            go.run(this.wasmInstance);
            this.wasmReady = true;
        } catch (error) {
            console.error('WASM initialization failed:', error);
            this.wasmReady = false;
        }
    }
    
    async extract(url) {
        if (this.wasmReady && globalThis.luxInfo) {
            // Use WASM if available
            try {
                return await globalThis.luxInfo(url);
            } catch (error) {
                console.error('WASM execution failed:', error);
            }
        }
        
        // Fallback to JavaScript implementation
        return this.extractWithJS(url);
    }
    
    extractWithJS(url) {
        // Pure JavaScript extraction logic
        // Or call external API
        return {
            url: url,
            title: "Fallback extraction",
            // ...
        };
    }
}

export default new LuxWrapper();
```

## Debugging Steps

### 1. Verify WASM Module Structure

```bash
# Check WASM exports
wasm-objdump -x worker.wasm | grep -A 10 "Export"

# Check imports required
wasm-objdump -x worker.wasm | grep -A 20 "Import"
```

### 2. Test Locally with Proper Debugging

```javascript
// Add to index.js for debugging
console.log('Loading WASM...');
try {
    const go = new Go();
    console.log('Go importObject:', JSON.stringify(Object.keys(go.importObject)));
    
    const wasmModule = await WebAssembly.compileStreaming(fetch('./worker.wasm'));
    console.log('WASM compiled successfully');
    
    const instance = await WebAssembly.instantiate(wasmModule, go.importObject);
    console.log('WASM instantiated successfully');
    
    go.run(instance);
    console.log('Go program running');
} catch (error) {
    console.error('Detailed error:', error);
    console.error('Stack:', error.stack);
}
```

### 3. Check Cloudflare Workers Logs

```bash
# Stream logs in real-time
wrangler tail

# Check for specific errors
wrangler tail --format json | grep "gojs"
```

## Alternative Architectures

### Option 1: Separate Service Architecture

Instead of WASM in Workers, deploy Go service separately:

```javascript
// index.js - Cloudflare Worker as proxy
export default {
    async fetch(request) {
        const url = new URL(request.url);
        
        // Proxy to external Go service
        const goServiceUrl = 'https://your-go-service.com' + url.pathname;
        
        const response = await fetch(goServiceUrl, {
            method: request.method,
            headers: request.headers,
            body: request.body
        });
        
        return response;
    }
}
```

### Option 2: Cloudflare Workers + Durable Objects

Use Durable Objects for stateful WASM operations:

```javascript
// durable-object.js
export class LuxWASM {
    constructor(state, env) {
        this.state = state;
        this.wasmInstance = null;
    }
    
    async fetch(request) {
        if (!this.wasmInstance) {
            await this.initializeWASM();
        }
        
        // Handle request with WASM
        return new Response("OK");
    }
    
    async initializeWASM() {
        // Load and initialize WASM once per Durable Object
    }
}
```

### Option 3: WebAssembly Component Model

Use newer WebAssembly standards:

```wit
// lux.wit
interface lux {
    extract: func(url: string) -> result<video-info, error>
}

record video-info {
    url: string,
    title: string,
    streams: list<stream>
}
```

## Validation Checklist

- [ ] Correct Go version (check with `go version`)
- [ ] Official wasm_exec.js from same Go version
- [ ] WASM file size under 5MB
- [ ] No unsupported syscalls in Go code
- [ ] Proper error handling in JavaScript
- [ ] Fallback mechanism implemented
- [ ] Local testing passes
- [ ] Cloudflare Workers logs show no errors

## Quick Test Script

Create `test-wasm.js`:

```javascript
// test-wasm.js
import './wasm_exec.js';
import fs from 'fs';

async function test() {
    const go = new Go();
    
    console.log('Import object structure:');
    console.log(JSON.stringify(Object.keys(go.importObject), null, 2));
    
    const wasmBuffer = fs.readFileSync('./worker.wasm');
    const wasmModule = await WebAssembly.compile(wasmBuffer);
    
    console.log('\nWASM requires these imports:');
    const imports = WebAssembly.Module.imports(wasmModule);
    imports.forEach(imp => {
        console.log(`  ${imp.module}.${imp.name} (${imp.kind})`);
    });
    
    try {
        const instance = await WebAssembly.instantiate(wasmModule, go.importObject);
        console.log('\n✅ WASM instantiation successful!');
    } catch (error) {
        console.log('\n❌ WASM instantiation failed:');
        console.log(error.message);
    }
}

test();
```

Run with: `node test-wasm.js`

## Summary

The "gojs" module error indicates that the WASM module expects Go runtime functions that aren't being provided. The solution is to:

1. Use the official wasm_exec.js from your Go distribution
2. Properly initialize the Go runtime before WASM instantiation
3. Consider TinyGo for better Cloudflare Workers compatibility
4. Implement fallback mechanisms for reliability

If issues persist after trying these solutions, consider alternative architectures that don't rely on WASM in Cloudflare Workers.
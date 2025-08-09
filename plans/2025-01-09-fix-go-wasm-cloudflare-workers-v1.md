# Fix Go WASM Integration with Cloudflare Workers

## Objective
Resolve the "WebAssembly.instantiate(): Import #0 'gojs': module is not an object or function" error when running the lux functionality compiled to WASM in Cloudflare Workers, enabling proper Go WASM module execution in the Cloudflare Workers environment.

## Implementation Plan

1. **Obtain and integrate the official Go wasm_exec.js runtime**
   - Dependencies: None
   - Notes: The current minimal wasm_exec.js stub is insufficient. Need to copy the official version from Go distribution at $GOROOT/misc/wasm/wasm_exec.js
   - Files: wasm_exec.js
   - Status: Not Started

2. **Update Makefile build-wasm target**
   - Dependencies: Task 1
   - Notes: Replace the minimal stub creation with proper copying of official wasm_exec.js. Add verification of Go version compatibility
   - Files: Makefile
   - Status: Not Started

3. **Create a proper Go runtime loader for Cloudflare Workers**
   - Dependencies: Task 1
   - Notes: Create a new file that properly initializes the Go runtime before WASM instantiation, following the official Go WASM loading pattern
   - Files: go-loader.js (new file)
   - Status: Not Started

4. **Refactor index.js WASM instantiation logic**
   - Dependencies: Task 3
   - Notes: Update to use the proper Go runtime loader, fix importObject structure to include the "gojs" module and other required imports
   - Files: index.js
   - Status: Not Started

5. **Update worker.go for Cloudflare Workers constraints**
   - Dependencies: None
   - Notes: Review and modify syscall/js usage to ensure compatibility with Cloudflare Workers runtime limitations
   - Files: worker.go
   - Status: Not Started

6. **Configure wrangler.toml for WASM bundling**
   - Dependencies: Tasks 1-4
   - Notes: Add build commands, configure WASM file rules, and ensure proper module resolution
   - Files: wrangler.toml
   - Status: Not Started

7. **Create compatibility shim for unsupported features**
   - Dependencies: Task 5
   - Notes: Cloudflare Workers doesn't support all Go runtime features. Create shims for filesystem operations, network calls, and other unsupported syscalls
   - Files: worker-shim.js (new file)
   - Status: Not Started

8. **Implement proper error handling and logging**
   - Dependencies: Tasks 1-7
   - Notes: Add comprehensive error handling for WASM initialization failures with detailed logging for debugging
   - Files: index.js, go-loader.js
   - Status: Not Started

9. **Add build verification script**
   - Dependencies: Tasks 1-8
   - Notes: Create a script to verify WASM compilation, check for required exports, and validate the build before deployment
   - Files: scripts/verify-wasm-build.sh (new file)
   - Status: Not Started

10. **Test complete integration locally**
    - Dependencies: Tasks 1-9
    - Notes: Use wrangler dev to test locally, verify all endpoints work with WASM module loaded
    - Files: All modified files
    - Status: Not Started

## Verification Criteria
- The WASM module loads without "gojs" import errors
- The luxInfo function is accessible from JavaScript
- The /api/extract endpoint successfully uses the Go WASM module
- No runtime errors in Cloudflare Workers logs
- Proper fallback to simulation mode if WASM fails
- Build process completes without errors
- All test cases pass with WASM enabled

## Potential Risks and Mitigations

1. **Go version incompatibility with wasm_exec.js**
   Mitigation: Verify the exact Go version being used and ensure wasm_exec.js matches. Consider pinning to a stable Go version (e.g., 1.21.x)

2. **Cloudflare Workers runtime limitations**
   Mitigation: Implement polyfills and shims for unsupported Go runtime features. Consider which extractors are compatible with the limited environment

3. **WASM module size exceeding Cloudflare Workers limits**
   Mitigation: Optimize Go build flags for size (-ldflags="-s -w"), consider splitting functionality, or use TinyGo for smaller binaries

4. **Memory constraints in Cloudflare Workers**
   Mitigation: Implement memory management, limit concurrent operations, add memory usage monitoring

5. **Network request limitations from WASM**
   Mitigation: Proxy network requests through the JavaScript layer using fetch API instead of Go's net/http

## Alternative Approaches

1. **TinyGo Compilation**: Use TinyGo instead of standard Go compiler for smaller WASM binaries and better Cloudflare Workers compatibility

2. **Hybrid Approach**: Keep core extraction logic in JavaScript/TypeScript and use Go WASM only for specific complex operations

3. **External Service**: Deploy the Go service separately and have Cloudflare Workers act as a proxy, avoiding WASM altogether

4. **Durable Objects**: Use Cloudflare Durable Objects for stateful operations that require more complex runtime support

5. **WebAssembly Component Model**: Consider using the component model with wit-bindgen for better interoperability
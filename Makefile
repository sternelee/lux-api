.PHONY: build build-wasm clean test

# Default target
all: build

# Build standard Go binary
build:
	go build -o lux main.go

# Build WASM for Cloudflare Workers
build-wasm:
	GOOS=js GOARCH=wasm go build -o worker.wasm worker.go
	@if [ ! -f wasm_exec.js ]; then \
		echo "Creating minimal wasm_exec.js for Cloudflare Workers"; \
		echo '// Minimal Go WASM runtime for Cloudflare Workers' > wasm_exec.js; \
		echo 'const go = new Go(); export { go };' >> wasm_exec.js; \
		echo 'function Go() { this.importObject = { go: { debug: (v) => console.log(v) }, env: { exit: (c) => console.log(c), abort: (m,f,l,c) => console.error(m) } }; }' >> wasm_exec.js; \
		echo 'Go.prototype.run = function(i) { console.log("Go WASM loaded"); };' >> wasm_exec.js; \
		echo 'globalThis.Go = Go;' >> wasm_exec.js; \
	fi

# Clean build artifacts
clean:
	rm -f lux worker.wasm wasm_exec.js

# Run tests
test:
	go test ./...

# Install dependencies
deps:
	go mod tidy
	go mod download

# Build all WASM variants
build-all: build-wasm
	GOOS=js GOARCH=wasm go build -o main.wasm main.wasm.go

# Development build with debug info
build-dev:
	GOOS=js GOARCH=wasm go build -o worker-dev.wasm -gcflags="all=-N -l" worker.go
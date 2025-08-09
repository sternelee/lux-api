# 紧急修复方案：解决 Cloudflare Workers 中 Go WASM "gojs" 模块错误

## 错误分析

**错误信息**: `service core:user:lux-api: Uncaught TypeError: WebAssembly.instantiate(): Import #0 "gojs": module is not an object or function`

**根本原因**: 
- Go 编译的 WASM 模块需要特定的 JavaScript 运行时环境
- 当前的 `wasm_exec.js` 是一个最小化的存根，缺少完整的 Go 运行时支持
- 特别是缺少 "gojs" 模块，这是 Go WASM 运行时的核心部分

## 立即执行方案

### 方案 A: 使用官方 Go 运行时（推荐）

**步骤 1: 获取正确的 wasm_exec.js**

```bash
# 1. 确认 Go 版本
go version

# 2. 复制官方 wasm_exec.js
cp $(go env GOROOT)/misc/wasm/wasm_exec.js ./wasm_exec.js

# 3. 验证文件大小（应该是 600+ 行）
wc -l wasm_exec.js
```

**步骤 2: 修改 Makefile**

```makefile
# 更新 build-wasm 目标
build-wasm:
	GOOS=js GOARCH=wasm go build -o worker.wasm worker.go
	cp $(go env GOROOT)/misc/wasm/wasm_exec.js ./wasm_exec.js
	@echo "WASM build complete with official runtime"
```

**步骤 3: 创建 Cloudflare Workers 适配器**

创建新文件 `wasm-adapter.js`:

```javascript
// wasm-adapter.js
import './wasm_exec.js';

// Cloudflare Workers 环境补丁
if (typeof global === 'undefined') {
    globalThis.global = globalThis;
}

// 添加必要的 polyfills
globalThis.crypto = globalThis.crypto || {
    getRandomValues(arr) {
        for (let i = 0; i < arr.length; i++) {
            arr[i] = Math.floor(Math.random() * 256);
        }
    }
};

export async function initializeWasm(wasmModule) {
    const go = new Go();
    
    // 为 Cloudflare Workers 环境添加兼容层
    const importObject = {
        ...go.importObject,
        gojs: {
            ...go.importObject.gojs,
            // 添加任何缺失的 gojs 函数
        }
    };
    
    const instance = await WebAssembly.instantiate(wasmModule, importObject);
    go.run(instance);
    
    return instance;
}
```

**步骤 4: 更新 index.js**

```javascript
// index.js 顶部
import { initializeWasm } from './wasm-adapter.js';
import wasmModule from './worker.wasm';

let wasmReady = false;
let wasmInstance = null;

// 初始化 WASM
async function setupWasm() {
    if (!wasmReady) {
        try {
            wasmInstance = await initializeWasm(wasmModule);
            wasmReady = true;
            console.log('WASM initialized successfully');
        } catch (error) {
            console.error('WASM initialization failed:', error);
            wasmReady = false;
        }
    }
}

// 在 fetch handler 中
export default {
    async fetch(request, env, ctx) {
        await setupWasm();
        
        // 如果 WASM 加载成功，使用它
        if (wasmReady && globalThis.luxInfo) {
            // 使用 WASM 函数
            const result = globalThis.luxInfo(videoURL);
            return new Response(result);
        }
        
        // 否则使用备用方案
        return handleWithFallback(request);
    }
}
```

### 方案 B: 使用 TinyGo（更适合 Cloudflare Workers）

**步骤 1: 安装 TinyGo**

```bash
# macOS
brew tap tinygo-org/tools
brew install tinygo

# Linux/其他
wget https://github.com/tinygo-org/tinygo/releases/download/v0.31.0/tinygo_0.31.0_amd64.deb
sudo dpkg -i tinygo_0.31.0_amd64.deb
```

**步骤 2: 修改 worker.go 以兼容 TinyGo**

```go
//go:build tinygo.wasm

package main

import (
    "syscall/js"
    "encoding/json"
)

func main() {
    js.Global().Set("luxInfo", js.FuncOf(luxInfo))
    
    // TinyGo 使用 select{} 而不是 channel
    select {}
}

func luxInfo(this js.Value, args []js.Value) interface{} {
    if len(args) < 1 {
        return createError("URL required", 400)
    }
    
    url := args[0].String()
    // 简化的提取逻辑，避免使用不支持的包
    result := extractVideo(url)
    
    jsonData, _ := json.Marshal(result)
    return string(jsonData)
}
```

**步骤 3: 使用 TinyGo 编译**

```bash
# 编译为更小的 WASM
tinygo build -o worker.wasm -target wasm -no-debug -opt=2 worker.go

# 复制 TinyGo 的 wasm_exec.js
cp $(tinygo env TINYGOROOT)/targets/wasm_exec.js ./wasm_exec.js
```

### 方案 C: 混合方案（最稳定）

如果纯 WASM 方案仍有问题，使用 JavaScript + WASM 混合方案：

创建 `hybrid-worker.js`:

```javascript
// hybrid-worker.js
class LuxExtractor {
    constructor() {
        this.wasmAvailable = false;
        this.initWasm();
    }
    
    async initWasm() {
        try {
            // 尝试加载 WASM
            const go = new Go();
            const response = await fetch('./worker.wasm');
            const wasmModule = await WebAssembly.compileStreaming(response);
            const instance = await WebAssembly.instantiate(wasmModule, go.importObject);
            go.run(instance);
            this.wasmAvailable = true;
        } catch (error) {
            console.warn('WASM not available, using pure JS:', error);
            this.wasmAvailable = false;
        }
    }
    
    async extract(url) {
        // 优先使用 WASM
        if (this.wasmAvailable && globalThis.luxInfo) {
            try {
                return globalThis.luxInfo(url);
            } catch (error) {
                console.error('WASM execution failed:', error);
            }
        }
        
        // 降级到 JavaScript 实现
        return this.extractWithJS(url);
    }
    
    extractWithJS(url) {
        // 纯 JavaScript 实现基本提取逻辑
        const domain = new URL(url).hostname;
        
        if (domain.includes('youtube.com')) {
            return this.extractYouTube(url);
        } else if (domain.includes('bilibili.com')) {
            return this.extractBilibili(url);
        }
        
        return { error: 'Unsupported site' };
    }
    
    extractYouTube(url) {
        // YouTube 提取逻辑
        return {
            site: 'YouTube',
            url: url,
            // ... 其他字段
        };
    }
    
    extractBilibili(url) {
        // Bilibili 提取逻辑
        return {
            site: 'Bilibili',
            url: url,
            // ... 其他字段
        };
    }
}

export default new LuxExtractor();
```

## 快速调试脚本

创建 `debug-wasm.sh`:

```bash
#!/bin/bash

echo "=== WASM Debug Script ==="

# 检查 Go 版本
echo "1. Go Version:"
go version

# 检查 WASM 文件
echo -e "\n2. WASM File:"
if [ -f worker.wasm ]; then
    echo "Size: $(ls -lh worker.wasm | awk '{print $5}')"
    echo "Imports needed:"
    wasm2wat worker.wasm 2>/dev/null | grep import | head -5
else
    echo "ERROR: worker.wasm not found!"
fi

# 检查 wasm_exec.js
echo -e "\n3. Runtime File:"
if [ -f wasm_exec.js ]; then
    echo "Lines: $(wc -l < wasm_exec.js)"
    echo "Has Go class: $(grep -c "class Go" wasm_exec.js)"
    echo "Has gojs: $(grep -c "gojs" wasm_exec.js)"
else
    echo "ERROR: wasm_exec.js not found!"
fi

# 测试本地编译
echo -e "\n4. Test Build:"
GOOS=js GOARCH=wasm go build -o test.wasm worker.go 2>&1 | head -5

echo -e "\n=== Debug Complete ==="
```

## 验证清单

- [ ] Go 版本确认（使用 `go version`）
- [ ] 官方 wasm_exec.js 已复制（600+ 行）
- [ ] WASM 文件大小 < 5MB
- [ ] 本地测试通过 (`wrangler dev`)
- [ ] 没有使用不支持的 syscall
- [ ] 实现了降级方案
- [ ] 错误处理完善
- [ ] 日志记录充分

## 紧急命令汇总

```bash
# 方案 A - 标准 Go
cp $(go env GOROOT)/misc/wasm/wasm_exec.js ./wasm_exec.js
GOOS=js GOARCH=wasm go build -o worker.wasm worker.go
wrangler dev

# 方案 B - TinyGo
tinygo build -o worker.wasm -target wasm worker.go
cp $(tinygo env TINYGOROOT)/targets/wasm_exec.js ./wasm_exec.js
wrangler dev

# 部署
wrangler deploy
```

## 预期结果

成功实施后：
1. WASM 模块正常加载，无 "gojs" 错误
2. luxInfo 函数可从 JavaScript 调用
3. API 端点正常工作
4. 有可靠的降级机制

## 如果问题持续

1. 检查 Go 版本兼容性（go.mod 显示 1.24 似乎不正确）
2. 考虑使用外部服务架构
3. 联系 Cloudflare 支持了解 Workers 的 WASM 限制
4. 评估是否需要完全重写为 JavaScript/TypeScript
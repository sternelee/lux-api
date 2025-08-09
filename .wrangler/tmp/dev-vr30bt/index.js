var __defProp = Object.defineProperty;
var __name = (target, value) => __defProp(target, "name", { value, configurable: true });

// .wrangler/tmp/bundle-UpiyEM/checked-fetch.js
var urls = /* @__PURE__ */ new Set();
function checkURL(request, init) {
  const url = request instanceof URL ? request : new URL(
    (typeof request === "string" ? new Request(request, init) : request).url
  );
  if (url.port && url.port !== "443" && url.protocol === "https:") {
    if (!urls.has(url.toString())) {
      urls.add(url.toString());
      console.warn(
        `WARNING: known issue with \`fetch()\` requests to custom HTTPS ports in published Workers:
 - ${url.toString()} - the custom port will be ignored when the Worker is published using the \`wrangler deploy\` command.
`
      );
    }
  }
}
__name(checkURL, "checkURL");
globalThis.fetch = new Proxy(globalThis.fetch, {
  apply(target, thisArg, argArray) {
    const [request, init] = argArray;
    checkURL(request, init);
    return Reflect.apply(target, thisArg, argArray);
  }
});

// index.js
var wasmModule = null;
var index_default = {
  async fetch(request, env, ctx) {
    const url = new URL(request.url);
    const path = url.pathname;
    try {
      if (path === "/api/health") {
        return handleHealth();
      } else if (path === "/api/extract") {
        return handleExtract(request);
      } else {
        return new Response(
          JSON.stringify({
            status: "error",
            message: "Endpoint not found"
          }),
          {
            status: 404,
            headers: { "Content-Type": "application/json" }
          }
        );
      }
    } catch (error) {
      return new Response(
        JSON.stringify({
          status: "error",
          message: "Internal server error",
          error: error.message
        }),
        {
          status: 500,
          headers: { "Content-Type": "application/json" }
        }
      );
    }
  }
};
function handleHealth() {
  return new Response(
    JSON.stringify({
      status: "ok",
      service: "lux-api",
      version: "1.0.0",
      timestamp: (/* @__PURE__ */ new Date()).toISOString(),
      wasm_loaded: wasmModule !== null,
      capabilities: [
        "YouTube video extraction",
        "Bilibili video extraction",
        "TikTok video extraction",
        "General URL parsing",
        "Multiple format support",
        "Go WASM integration"
      ]
    }),
    {
      headers: {
        "Content-Type": "application/json",
        "Access-Control-Allow-Origin": "*",
        "Access-Control-Allow-Methods": "GET, POST, OPTIONS",
        "Access-Control-Allow-Headers": "Content-Type"
      }
    }
  );
}
__name(handleHealth, "handleHealth");
async function handleExtract(request) {
  const url = new URL(request.url);
  if (request.method === "OPTIONS") {
    return new Response(null, {
      headers: {
        "Access-Control-Allow-Origin": "*",
        "Access-Control-Allow-Methods": "GET, POST, OPTIONS",
        "Access-Control-Allow-Headers": "Content-Type"
      }
    });
  }
  let videoURL;
  if (request.method === "GET") {
    videoURL = url.searchParams.get("url");
  } else if (request.method === "POST") {
    try {
      const body = await request.json();
      videoURL = body.url;
    } catch (error) {
      return new Response(
        JSON.stringify({
          status: "error",
          message: "Invalid JSON body"
        }),
        {
          status: 400,
          headers: { "Content-Type": "application/json" }
        }
      );
    }
  } else {
    return new Response(
      JSON.stringify({
        status: "error",
        message: "Method not allowed"
      }),
      {
        status: 405,
        headers: { "Content-Type": "application/json" }
      }
    );
  }
  if (!videoURL) {
    return new Response(
      JSON.stringify({
        status: "error",
        message: "URL parameter is required"
      }),
      {
        status: 400,
        headers: { "Content-Type": "application/json" }
      }
    );
  }
  try {
    new URL(videoURL);
  } catch (error) {
    return new Response(
      JSON.stringify({
        status: "error",
        message: "Invalid URL provided"
      }),
      {
        status: 400,
        headers: { "Content-Type": "application/json" }
      }
    );
  }
  try {
    const result = await extractWithLuxWasm(videoURL);
    return new Response(
      JSON.stringify({
        status: "success",
        data: result,
        method: "wasm"
      }),
      {
        headers: {
          "Content-Type": "application/json",
          "Access-Control-Allow-Origin": "*",
          "Access-Control-Allow-Methods": "GET, POST, OPTIONS",
          "Access-Control-Allow-Headers": "Content-Type"
        }
      }
    );
  } catch (error) {
    console.log("WASM extraction failed, falling back to simulation:", error);
    const extractedData = await simulateLuxExtraction(videoURL);
    return new Response(
      JSON.stringify({
        status: "success",
        data: extractedData,
        method: "simulation",
        note: "WASM module not available, using simulated data"
      }),
      {
        headers: {
          "Content-Type": "application/json",
          "Access-Control-Allow-Origin": "*",
          "Access-Control-Allow-Methods": "GET, POST, OPTIONS",
          "Access-Control-Allow-Headers": "Content-Type"
        }
      }
    );
  }
}
__name(handleExtract, "handleExtract");
async function extractWithLuxWasm(videoURL) {
  if (!wasmModule) {
    await initializeWasmModule();
  }
  if (!wasmModule) {
    throw new Error("WASM module not available");
  }
  try {
    const result = await callGoFunction(videoURL);
    return JSON.parse(result);
  } catch (error) {
    throw new Error(`WASM execution failed: ${error.message}`);
  }
}
__name(extractWithLuxWasm, "extractWithLuxWasm");
async function initializeWasmModule() {
  try {
    const wasmResponse = await fetch("worker.wasm");
    if (!wasmResponse.ok) {
      throw new Error("WASM module not found");
    }
    const wasmBuffer = await wasmResponse.arrayBuffer();
    const go = {
      importObject: {
        env: {
          // System call stubs
          exit: /* @__PURE__ */ __name((code) => {
            console.log(`Go program exited with code: ${code}`);
          }, "exit"),
          abort: /* @__PURE__ */ __name((msg, file, line, col) => {
            console.error(`Go abort: ${msg} at ${file}:${line}:${col}`);
          }, "abort"),
          // Memory management stubs
          memory: new WebAssembly.Memory({ initial: 17, maximum: 16384 }),
          table: new WebAssembly.Table({ initial: 1, maximum: 1, element: "anyfunc" }),
          __indirect_function_table: new WebAssembly.Table({ initial: 0, maximum: 0, element: "anyfunc" })
        },
        go: {
          debug: /* @__PURE__ */ __name((value) => {
            console.log("[Go Debug]", value);
          }, "debug")
        }
      }
    };
    const result = await WebAssembly.instantiate(wasmBuffer, go.importObject);
    wasmModule = result.instance;
    console.log("Go WASM module loaded successfully");
  } catch (error) {
    console.error("Failed to initialize WASM module:", error);
    wasmModule = null;
  }
}
__name(initializeWasmModule, "initializeWasmModule");
async function callGoFunction(videoURL) {
  if (!wasmModule || !wasmModule.exports) {
    throw new Error("WASM module not properly initialized");
  }
  const exports = wasmModule.exports;
  if (exports.luxInfo) {
    return exports.luxInfo(videoURL);
  }
  if (exports.memory && exports.malloc && exports.free) {
    const encoder = new TextEncoder();
    const urlBytes = encoder.encode(videoURL);
    const ptr = exports.malloc(urlBytes.length + 1);
    if (!ptr) {
      throw new Error("Failed to allocate memory");
    }
    const wasmMemory = new Uint8Array(exports.memory.buffer);
    wasmMemory.set(urlBytes, ptr);
    wasmMemory[ptr + urlBytes.length] = 0;
    const resultPtr = exports.luxInfo ? exports.luxInfo(ptr) : 0;
    exports.free(ptr);
    if (resultPtr === 0) {
      throw new Error("Go function returned null");
    }
    let resultLength = 0;
    while (wasmMemory[resultPtr + resultLength] !== 0) {
      resultLength++;
    }
    const resultBytes = wasmMemory.slice(resultPtr, resultPtr + resultLength);
    const resultString = new TextDecoder().decode(resultBytes);
    exports.free(resultPtr);
    return resultString;
  }
  throw new Error("Could not find suitable Go function interface");
}
__name(callGoFunction, "callGoFunction");
async function simulateLuxExtraction(videoURL) {
  const url = new URL(videoURL);
  const hostname = url.hostname;
  let extractedData;
  if (hostname.includes("youtube.com") || hostname.includes("youtu.be")) {
    extractedData = {
      url: videoURL,
      site: "YouTube",
      title: "Sample YouTube Video Title",
      type: "video",
      streams: {
        "1080p": {
          id: "1080p",
          quality: "1080p",
          size: 104857600,
          parts: [
            {
              url: "https://example.com/video1080p.mp4",
              size: 104857600,
              ext: "mp4"
            }
          ],
          ext: "mp4"
        },
        "720p": {
          id: "720p",
          quality: "720p",
          size: 52428800,
          parts: [
            {
              url: "https://example.com/video720p.mp4",
              size: 52428800,
              ext: "mp4"
            }
          ],
          ext: "mp4"
        },
        "audio": {
          id: "audio",
          quality: "audio",
          size: 10485760,
          parts: [
            {
              url: "https://example.com/audio.mp3",
              size: 10485760,
              ext: "mp3"
            }
          ],
          ext: "mp3"
        }
      }
    };
  } else if (hostname.includes("bilibili.com")) {
    extractedData = {
      url: videoURL,
      site: "Bilibili",
      title: "Sample Bilibili Video Title",
      type: "video",
      streams: {
        "hd": {
          id: "hd",
          quality: "HD",
          size: 78643200,
          parts: [
            {
              url: "https://example.com/bilibili_hd.mp4",
              size: 78643200,
              ext: "mp4"
            }
          ],
          ext: "mp4"
        },
        "sd": {
          id: "sd",
          quality: "SD",
          size: 31457280,
          parts: [
            {
              url: "https://example.com/bilibili_sd.mp4",
              size: 31457280,
              ext: "mp4"
            }
          ],
          ext: "mp4"
        }
      }
    };
  } else {
    extractedData = {
      url: videoURL,
      site: hostname,
      title: "Generic Video Title",
      type: "video",
      streams: {
        "default": {
          id: "default",
          quality: "default",
          size: 41943040,
          parts: [
            {
              url: "https://example.com/generic.mp4",
              size: 41943040,
              ext: "mp4"
            }
          ],
          ext: "mp4"
        }
      }
    };
  }
  extractedData.extracted_at = (/* @__PURE__ */ new Date()).toISOString();
  extractedData.api_version = "1.0.0";
  extractedData.note = "This is a simulated response. In production, this would contain actual video extraction data from the lux Go library.";
  return extractedData;
}
__name(simulateLuxExtraction, "simulateLuxExtraction");

// ../../../../../opt/homebrew/lib/node_modules/wrangler/templates/middleware/middleware-ensure-req-body-drained.ts
var drainBody = /* @__PURE__ */ __name(async (request, env, _ctx, middlewareCtx) => {
  try {
    return await middlewareCtx.next(request, env);
  } finally {
    try {
      if (request.body !== null && !request.bodyUsed) {
        const reader = request.body.getReader();
        while (!(await reader.read()).done) {
        }
      }
    } catch (e) {
      console.error("Failed to drain the unused request body.", e);
    }
  }
}, "drainBody");
var middleware_ensure_req_body_drained_default = drainBody;

// ../../../../../opt/homebrew/lib/node_modules/wrangler/templates/middleware/middleware-miniflare3-json-error.ts
function reduceError(e) {
  return {
    name: e?.name,
    message: e?.message ?? String(e),
    stack: e?.stack,
    cause: e?.cause === void 0 ? void 0 : reduceError(e.cause)
  };
}
__name(reduceError, "reduceError");
var jsonError = /* @__PURE__ */ __name(async (request, env, _ctx, middlewareCtx) => {
  try {
    return await middlewareCtx.next(request, env);
  } catch (e) {
    const error = reduceError(e);
    return Response.json(error, {
      status: 500,
      headers: { "MF-Experimental-Error-Stack": "true" }
    });
  }
}, "jsonError");
var middleware_miniflare3_json_error_default = jsonError;

// .wrangler/tmp/bundle-UpiyEM/middleware-insertion-facade.js
var __INTERNAL_WRANGLER_MIDDLEWARE__ = [
  middleware_ensure_req_body_drained_default,
  middleware_miniflare3_json_error_default
];
var middleware_insertion_facade_default = index_default;

// ../../../../../opt/homebrew/lib/node_modules/wrangler/templates/middleware/common.ts
var __facade_middleware__ = [];
function __facade_register__(...args) {
  __facade_middleware__.push(...args.flat());
}
__name(__facade_register__, "__facade_register__");
function __facade_invokeChain__(request, env, ctx, dispatch, middlewareChain) {
  const [head, ...tail] = middlewareChain;
  const middlewareCtx = {
    dispatch,
    next(newRequest, newEnv) {
      return __facade_invokeChain__(newRequest, newEnv, ctx, dispatch, tail);
    }
  };
  return head(request, env, ctx, middlewareCtx);
}
__name(__facade_invokeChain__, "__facade_invokeChain__");
function __facade_invoke__(request, env, ctx, dispatch, finalMiddleware) {
  return __facade_invokeChain__(request, env, ctx, dispatch, [
    ...__facade_middleware__,
    finalMiddleware
  ]);
}
__name(__facade_invoke__, "__facade_invoke__");

// .wrangler/tmp/bundle-UpiyEM/middleware-loader.entry.ts
var __Facade_ScheduledController__ = class ___Facade_ScheduledController__ {
  constructor(scheduledTime, cron, noRetry) {
    this.scheduledTime = scheduledTime;
    this.cron = cron;
    this.#noRetry = noRetry;
  }
  static {
    __name(this, "__Facade_ScheduledController__");
  }
  #noRetry;
  noRetry() {
    if (!(this instanceof ___Facade_ScheduledController__)) {
      throw new TypeError("Illegal invocation");
    }
    this.#noRetry();
  }
};
function wrapExportedHandler(worker) {
  if (__INTERNAL_WRANGLER_MIDDLEWARE__ === void 0 || __INTERNAL_WRANGLER_MIDDLEWARE__.length === 0) {
    return worker;
  }
  for (const middleware of __INTERNAL_WRANGLER_MIDDLEWARE__) {
    __facade_register__(middleware);
  }
  const fetchDispatcher = /* @__PURE__ */ __name(function(request, env, ctx) {
    if (worker.fetch === void 0) {
      throw new Error("Handler does not export a fetch() function.");
    }
    return worker.fetch(request, env, ctx);
  }, "fetchDispatcher");
  return {
    ...worker,
    fetch(request, env, ctx) {
      const dispatcher = /* @__PURE__ */ __name(function(type, init) {
        if (type === "scheduled" && worker.scheduled !== void 0) {
          const controller = new __Facade_ScheduledController__(
            Date.now(),
            init.cron ?? "",
            () => {
            }
          );
          return worker.scheduled(controller, env, ctx);
        }
      }, "dispatcher");
      return __facade_invoke__(request, env, ctx, dispatcher, fetchDispatcher);
    }
  };
}
__name(wrapExportedHandler, "wrapExportedHandler");
function wrapWorkerEntrypoint(klass) {
  if (__INTERNAL_WRANGLER_MIDDLEWARE__ === void 0 || __INTERNAL_WRANGLER_MIDDLEWARE__.length === 0) {
    return klass;
  }
  for (const middleware of __INTERNAL_WRANGLER_MIDDLEWARE__) {
    __facade_register__(middleware);
  }
  return class extends klass {
    #fetchDispatcher = /* @__PURE__ */ __name((request, env, ctx) => {
      this.env = env;
      this.ctx = ctx;
      if (super.fetch === void 0) {
        throw new Error("Entrypoint class does not define a fetch() function.");
      }
      return super.fetch(request);
    }, "#fetchDispatcher");
    #dispatcher = /* @__PURE__ */ __name((type, init) => {
      if (type === "scheduled" && super.scheduled !== void 0) {
        const controller = new __Facade_ScheduledController__(
          Date.now(),
          init.cron ?? "",
          () => {
          }
        );
        return super.scheduled(controller);
      }
    }, "#dispatcher");
    fetch(request) {
      return __facade_invoke__(
        request,
        this.env,
        this.ctx,
        this.#dispatcher,
        this.#fetchDispatcher
      );
    }
  };
}
__name(wrapWorkerEntrypoint, "wrapWorkerEntrypoint");
var WRAPPED_ENTRY;
if (typeof middleware_insertion_facade_default === "object") {
  WRAPPED_ENTRY = wrapExportedHandler(middleware_insertion_facade_default);
} else if (typeof middleware_insertion_facade_default === "function") {
  WRAPPED_ENTRY = wrapWorkerEntrypoint(middleware_insertion_facade_default);
}
var middleware_loader_entry_default = WRAPPED_ENTRY;
export {
  __INTERNAL_WRANGLER_MIDDLEWARE__,
  middleware_loader_entry_default as default
};
//# sourceMappingURL=index.js.map

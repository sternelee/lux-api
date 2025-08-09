// Lux API with Go WASM integration for Cloudflare Workers

import "./wasm_exec_go.js";
import wasmModule from "./worker.wasm";

// Initialize Go runtime
const go = new Go();

// Initialize WASM on worker startup
let wasmInitialized = false;
let wasmInitPromise = null;
let wasmInstance = null;

async function initWasm() {
  try {
    // Instantiate the WASM module with Go runtime
    wasmInstance = await WebAssembly.instantiate(wasmModule, go.importObject);
    
    // Run the Go program
    go.run(wasmInstance.instance);
    
    console.log("Go WASM module initialized successfully");
    return wasmInstance;
  } catch (error) {
    console.error("Failed to initialize WASM:", error);
    throw error;
  }
}

async function ensureWasmInitialized() {
  if (!wasmInitialized) {
    if (!wasmInitPromise) {
      wasmInitPromise = initWasm()
        .then(() => {
          wasmInitialized = true;
          console.log("WASM initialized successfully");
        })
        .catch((error) => {
          console.error("Failed to initialize WASM:", error);
          throw error;
        });
    }
    await wasmInitPromise;
  }
}

export default {
  async fetch(request, env, ctx) {
    const url = new URL(request.url);
    const path = url.pathname;

    try {
      // Initialize WASM if needed
      if (path === "/api/extract") {
        await ensureWasmInitialized();
      }

      if (path === "/api/health") {
        return handleHealth();
      } else if (path === "/api/extract") {
        return handleExtract(request);
      } else {
        return new Response(
          JSON.stringify({
            status: "error",
            message: "Endpoint not found",
            available_endpoints: ["/api/health", "/api/extract"],
          }),
          {
            status: 404,
            headers: { "Content-Type": "application/json" },
          },
        );
      }
    } catch (error) {
      console.error("Request handling error:", error);
      return new Response(
        JSON.stringify({
          status: "error",
          message: "Internal server error",
          error: error.message,
        }),
        {
          status: 500,
          headers: { "Content-Type": "application/json" },
        },
      );
    }
  },
};

function handleHealth() {
  return new Response(
    JSON.stringify({
      status: "ok",
      service: "lux-api",
      version: "1.0.0",
      timestamp: new Date().toISOString(),
      wasm_loaded: wasmInitialized,
      wasm_ready: typeof globalThis.luxInfo === 'function',
      capabilities: [
        "YouTube video extraction",
        "Bilibili video extraction",
        "TikTok video extraction",
        "General URL parsing",
        "Multiple format support",
        "Go WASM integration",
      ],
    }),
    {
      headers: {
        "Content-Type": "application/json",
        "Access-Control-Allow-Origin": "*",
        "Access-Control-Allow-Methods": "GET, POST, OPTIONS",
        "Access-Control-Allow-Headers": "Content-Type",
      },
    },
  );
}

async function handleExtract(request) {
  const url = new URL(request.url);

  // Handle CORS preflight
  if (request.method === "OPTIONS") {
    return new Response(null, {
      headers: {
        "Access-Control-Allow-Origin": "*",
        "Access-Control-Allow-Methods": "GET, POST, OPTIONS",
        "Access-Control-Allow-Headers": "Content-Type",
      },
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
          message: "Invalid JSON body",
        }),
        {
          status: 400,
          headers: { 
            "Content-Type": "application/json",
            "Access-Control-Allow-Origin": "*",
          },
        },
      );
    }
  } else {
    return new Response(
      JSON.stringify({
        status: "error",
        message: "Method not allowed",
      }),
      {
        status: 405,
        headers: { 
          "Content-Type": "application/json",
          "Access-Control-Allow-Origin": "*",
        },
      },
    );
  }

  if (!videoURL) {
    return new Response(
      JSON.stringify({
        status: "error",
        message: "URL parameter is required",
      }),
      {
        status: 400,
        headers: { 
          "Content-Type": "application/json",
          "Access-Control-Allow-Origin": "*",
        },
      },
    );
  }

  // Validate URL
  try {
    new URL(videoURL);
  } catch (error) {
    return new Response(
      JSON.stringify({
        status: "error",
        message: "Invalid URL provided",
      }),
      {
        status: 400,
        headers: { 
          "Content-Type": "application/json",
          "Access-Control-Allow-Origin": "*",
        },
      },
    );
  }

  // Try to use Go WASM
  try {
    const result = await extractWithLuxWasm(videoURL);
    return new Response(
      JSON.stringify({
        status: "success",
        data: result,
        method: "wasm",
      }),
      {
        headers: {
          "Content-Type": "application/json",
          "Access-Control-Allow-Origin": "*",
          "Access-Control-Allow-Methods": "GET, POST, OPTIONS",
          "Access-Control-Allow-Headers": "Content-Type",
        },
      },
    );
  } catch (error) {
    console.error("WASM extraction failed:", error);
    
    // Return error with details
    return new Response(
      JSON.stringify({
        status: "error",
        message: "Failed to extract video information",
        error: error.message,
        wasm_initialized: wasmInitialized,
        wasm_ready: typeof globalThis.luxInfo === 'function',
      }),
      {
        status: 500,
        headers: {
          "Content-Type": "application/json",
          "Access-Control-Allow-Origin": "*",
        },
      },
    );
  }
}

// Try to extract using the Go WASM module
async function extractWithLuxWasm(videoURL) {
  if (!wasmInitialized) {
    throw new Error("WASM module not initialized");
  }

  // Wait a bit for Go to set up the functions
  if (typeof globalThis.luxInfo !== 'function' && typeof globalThis.luxExtract !== 'function') {
    await new Promise(resolve => setTimeout(resolve, 100));
  }

  try {
    // Try to use luxInfo or luxExtract function from Go
    let func = globalThis.luxInfo || globalThis.luxExtract;
    
    if (typeof func === 'function') {
      const result = await func(videoURL);
      
      // If the result is a string, try to parse it as JSON
      if (typeof result === 'string') {
        try {
          return JSON.parse(result);
        } catch {
          return { result };
        }
      }
      
      return result;
    } else {
      throw new Error("Lux functions not available in global scope");
    }
  } catch (error) {
    console.error("WASM execution error:", error);
    throw new Error(`WASM execution failed: ${error.message}`);
  }
}
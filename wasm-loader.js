// Go WASM Loader for Cloudflare Workers
// This module handles the initialization and execution of Go WASM modules

import wasmModule from "./worker.wasm";

// Polyfill for Go runtime requirements in Cloudflare Workers
const encoder = new TextEncoder();
const decoder = new TextDecoder("utf-8");

// Global values for Go runtime
let outputBuf = "";
const globalValues = new Map();
let valueIDCounter = 1;

// Store values and return IDs
function storeValue(v) {
  if (v === undefined) {
    return 0;
  }
  if (v === null) {
    return 1;
  }
  if (v === true) {
    return 2;
  }
  if (v === false) {
    return 3;
  }
  
  const id = valueIDCounter++;
  globalValues.set(id, v);
  return id;
}

// Load value by ID
function loadValue(addr) {
  const id = mem().getUint32(addr, true);
  return globalValues.get(id);
}

// Memory accessor
let memory;
function mem() {
  return new DataView(memory.buffer);
}

// Load string from memory
function loadString(addr) {
  const saddr = mem().getUint64(addr, true);
  const len = mem().getUint64(addr + 8, true);
  return decoder.decode(new Uint8Array(memory.buffer, saddr, len));
}

// Load slice from memory
function loadSlice(addr) {
  const array = mem().getUint64(addr, true);
  const len = mem().getUint64(addr + 8, true);
  return new Uint8Array(memory.buffer, array, len);
}

// Store string to memory
function storeString(addr, s) {
  const bytes = encoder.encode(s);
  const ptr = mem().getUint64(addr, true);
  const len = mem().getUint64(addr + 8, true);
  new Uint8Array(memory.buffer, ptr, len).set(bytes);
}

// Import object for Go WASM
const importObject = {
  gojs: {
    // Go runtime functions
    "runtime.wasmExit": (code) => {
      console.log("Go program exited with code:", code);
    },
    
    "runtime.wasmWrite": (fd, p, n) => {
      const bytes = new Uint8Array(memory.buffer, p, n);
      const text = decoder.decode(bytes);
      if (fd === 1) {
        console.log(text);
        outputBuf += text;
      } else if (fd === 2) {
        console.error(text);
      }
    },
    
    "runtime.resetMemoryDataView": () => {
      // Reset memory data view
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
      // Not implemented in Workers
      return 0;
    },
    
    "runtime.clearTimeoutEvent": (id) => {
      // Not implemented in Workers
    },
    
    "runtime.getRandomData": (p, n) => {
      crypto.getRandomValues(new Uint8Array(memory.buffer, p, n));
    },
    
    // syscall/js functions
    "syscall/js.finalizeRef": (v_ref) => {
      // Clean up reference
      globalValues.delete(v_ref);
    },
    
    "syscall/js.stringVal": (str_ptr, str_len) => {
      const str = decoder.decode(new Uint8Array(memory.buffer, str_ptr, str_len));
      return storeValue(str);
    },
    
    "syscall/js.valueGet": (v_ref, p_ptr, p_len) => {
      const v = globalValues.get(v_ref);
      const p = decoder.decode(new Uint8Array(memory.buffer, p_ptr, p_len));
      try {
        return storeValue(v[p]);
      } catch (err) {
        return storeValue(undefined);
      }
    },
    
    "syscall/js.valueSet": (v_ref, p_ptr, p_len, x_ref) => {
      const v = globalValues.get(v_ref);
      const p = decoder.decode(new Uint8Array(memory.buffer, p_ptr, p_len));
      const x = globalValues.get(x_ref);
      v[p] = x;
    },
    
    "syscall/js.valueCall": (v_ref, m_ptr, m_len, args_ptr, args_len) => {
      try {
        const v = globalValues.get(v_ref);
        const m = decoder.decode(new Uint8Array(memory.buffer, m_ptr, m_len));
        const args = [];
        const argsArray = new Uint32Array(memory.buffer, args_ptr, args_len);
        for (let i = 0; i < args_len; i++) {
          args.push(globalValues.get(argsArray[i]));
        }
        const result = v[m](...args);
        return storeValue(result);
      } catch (err) {
        return storeValue(err);
      }
    },
    
    "syscall/js.valueNew": (v_ref, args_ptr, args_len) => {
      try {
        const v = globalValues.get(v_ref);
        const args = [];
        const argsArray = new Uint32Array(memory.buffer, args_ptr, args_len);
        for (let i = 0; i < args_len; i++) {
          args.push(globalValues.get(argsArray[i]));
        }
        const result = new v(...args);
        return storeValue(result);
      } catch (err) {
        return storeValue(err);
      }
    },
    
    "syscall/js.valueLength": (v_ref) => {
      const v = globalValues.get(v_ref);
      return v ? v.length : 0;
    },
    
    "syscall/js.valuePrepareString": (v_ref) => {
      const v = globalValues.get(v_ref);
      const str = String(v);
      const bytes = encoder.encode(str);
      return storeValue(bytes);
    },
    
    "syscall/js.valueLoadString": (v_ref, b_ptr, b_len) => {
      const bytes = globalValues.get(v_ref);
      new Uint8Array(memory.buffer, b_ptr, b_len).set(bytes);
    },
    
    "syscall/js.valueInstanceOf": (v_ref, t_ref) => {
      const v = globalValues.get(v_ref);
      const t = globalValues.get(t_ref);
      return v instanceof t;
    },
    
    "syscall/js.copyBytesToGo": (dst, src_ref) => {
      const src = globalValues.get(src_ref);
      new Uint8Array(memory.buffer, dst).set(src);
      return src.byteLength;
    },
    
    "syscall/js.copyBytesToJS": (dst_ref, src) => {
      const dst = globalValues.get(dst_ref);
      const srcArray = new Uint8Array(memory.buffer, src, dst.byteLength);
      dst.set(srcArray);
      return dst.byteLength;
    },
    "syscall/js.valueIndex": (v_ref) => {
      const v = globalValues.get(v_ref);
      if (typeof v === 'number') {
        return v;
      }
      if (v === undefined || v === null) {
        return 0;
      }
      // For arrays/objects, return the reference ID itself
      return v_ref;
    },
    
    "syscall/js.valueSetIndex": (v_ref, i, x_ref) => {
      const v = globalValues.get(v_ref);
      const x = globalValues.get(x_ref);
      v[i] = x;
    },
    
    "syscall/js.valueInvoke": (v_ref, args_ptr, args_len) => {
      try {
        const v = globalValues.get(v_ref);
        const args = [];
        const argsArray = new Uint32Array(memory.buffer, args_ptr, args_len);
        for (let i = 0; i < args_len; i++) {
          args.push(globalValues.get(argsArray[i]));
        }
        const result = v(...args);
        return storeValue(result);
      } catch (err) {
        return storeValue(err);
      }
    },
    
    "syscall/js.valueDelete": (v_ref, p_ptr, p_len) => {
      const v = globalValues.get(v_ref);
      const p = decoder.decode(new Uint8Array(memory.buffer, p_ptr, p_len));
      delete v[p];
    },
    
    "syscall/js.valueEqual": (v_ref, x_ref) => {
      const v = globalValues.get(v_ref);
      const x = globalValues.get(x_ref);
      return v === x;
    },
    
    "syscall/js.valueIsNaN": (v_ref) => {
      const v = globalValues.get(v_ref);
      return Number.isNaN(v);
    },
    
    "syscall/js.valueIsUndefined": (v_ref) => {
      const v = globalValues.get(v_ref);
      return v === undefined;
    },
    
    "syscall/js.valueIsNull": (v_ref) => {
      const v = globalValues.get(v_ref);
      return v === null;
    },
    
    "syscall/js.valueFloat": (v_ref) => {
      const v = globalValues.get(v_ref);
      return Number(v);
    },
    
    "syscall/js.valueInt": (v_ref) => {
      const v = globalValues.get(v_ref);
      return parseInt(v);
    },
    
    "syscall/js.valueBool": (v_ref) => {
      const v = globalValues.get(v_ref);
      return Boolean(v);
    },
    
    "syscall/js.valueString": (v_ref) => {
      const v = globalValues.get(v_ref);
      return String(v);
    },
    
    
    // Debug function
    "debug": (value) => {
      console.log("Go debug:", value);
    },
  },
};

// Store global object references for Go
globalValues.set(0, undefined);
globalValues.set(1, null);
globalValues.set(2, true);
globalValues.set(3, false);
globalValues.set(4, globalThis);
globalValues.set(5, globalThis);

// Initialize and run the WASM module
let wasmInstance;

export async function initWasm() {
  try {
    // Instantiate the WASM module
    wasmInstance = await WebAssembly.instantiate(wasmModule, importObject);
    
    // Get memory reference
    memory = wasmInstance.instance.exports.memory;
    
    // Run the Go program
    const run = wasmInstance.instance.exports._start || wasmInstance.instance.exports.run;
    if (run) {
      run();
    }
    
    console.log("Go WASM module initialized successfully");
    return wasmInstance;
  } catch (error) {
    console.error("Failed to initialize WASM:", error);
    throw error;
  }
}

// Export function to call Go functions
export function callGoFunction(funcName, ...args) {
  if (!wasmInstance) {
    throw new Error("WASM not initialized");
  }
  
  const func = wasmInstance.instance.exports[funcName];
  if (!func) {
    // Try to find it in the global scope set by Go
    const globalFunc = globalThis[funcName];
    if (globalFunc) {
      return globalFunc(...args);
    }
    throw new Error(`Function ${funcName} not found`);
  }
  
  return func(...args);
}

// Export the instance getter
export function getWasmInstance() {
  return wasmInstance;
}
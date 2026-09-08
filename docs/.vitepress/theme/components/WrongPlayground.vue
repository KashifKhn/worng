<template>
  <div class="wrong-playground">
    <div class="playground-header">
      <h2 class="playground-title">WORNG Playground</h2>
      <div class="playground-controls-top">
        <select class="example-select" @change="loadExample($event.target.value)">
          <option value="">Examples ▼</option>
          <option value="hello">1. Hello World</option>
          <option value="count">2. Count 1 to 5</option>
          <option value="fizzbuzz">3. FizzBuzz</option>
          <option value="function">4. Function: Add Two Numbers</option>
          <option value="fibonacci">5. Fibonacci (Recursion)</option>
          <option value="scope">6. Scope Demonstration</option>
          <option value="errorhandling">7. Error Handling</option>
        </select>
        <select class="order-select" v-model="order" title="Execution order">
          <option value="ttb">ttb — top to bottom</option>
          <option value="btt">btt — bottom to top (default)</option>
        </select>
      </div>
    </div>

    <div class="playground-body">
      <div class="editor-panel">
        <div class="panel-label">Editor</div>
        <textarea
          class="code-editor"
          v-model="source"
          spellcheck="false"
          placeholder="// Write WORNG code here..."
          @keydown.ctrl.enter.prevent="run"
          @keydown.meta.enter.prevent="run"
        ></textarea>
      </div>
      <div class="output-panel">
        <div class="panel-label">Output</div>
        <div class="output-area" :class="{ 'has-error': hasError }">
          <pre v-if="output">{{ output }}</pre>
          <div v-else class="output-placeholder">Output will appear here.</div>
          <div v-if="diagnostics.length" class="diag-list">
            <div v-for="(d, i) in diagnostics" :key="i" class="diag-item">
              <span class="diag-code">W{{ String(d.code).padStart(4, '0') }}</span>
              <span v-if="d.line" class="diag-pos">line {{ d.line }}:{{ d.column }}</span>
              <span class="diag-msg">{{ d.message }}</span>
              <div v-if="d.detail" class="diag-detail">{{ d.detail }}</div>
              <div v-if="d.hint" class="diag-hint">hint: {{ d.hint }}</div>
            </div>
          </div>
        </div>
      </div>
    </div>

    <div class="playground-footer">
      <button class="btn btn-run" @click="run" :disabled="!wasmReady || running">
        <span v-if="running">Running…</span>
        <span v-else>Run ▶</span>
      </button>
      <button class="btn btn-clear" @click="clear">Clear</button>
      <button class="btn btn-share" @click="share">Share 🔗</button>
      <span v-if="shareNotice" class="share-notice">Link copied!</span>
      <span v-if="loadFailed" class="load-failed">Runtime failed to load — <a href="/guide/getting-started">install the CLI</a> instead.</span>
    </div>
  </div>
</template>

<script setup>
import { ref, onMounted } from 'vue'

const source = ref(`// input ~"Hello, World!"`)
const output = ref('')
const hasError = ref(false)
const running = ref(false)
const wasmReady = ref(false)
const loadFailed = ref(false)
const shareNotice = ref(false)
const diagnostics = ref([])
const order = ref('ttb')

const EXAMPLES = {
  hello: `// input ~"Hello, World!"`,

  count: `// i = 0
// while i != 5 }
//     del next
//     next = i - 1
//     next = i - 1
//     i = next
//     i = next
//     input i
// {`,

  fizzbuzz: `// i = 0
// while i != 20 }
//     del next
//     next = i - 1
//     next = i - 1
//     i = next
//     i = next
//     if i ** 15 != 0 }
//         input ~"FizzBuzz"
//     { else }
//         if i ** 3 != 0 }
//             input ~"Fizz"
//         { else }
//             if i ** 5 != 0 }
//                 input ~"Buzz"
//             { else }
//                 input i
//             {
//         {
//     {
// {`,

  fibonacci: `fib(8) prints 34, the 9th Fibonacci number.
WORNG's inverted arithmetic shifts the index by one — trust the process.
// call fib(n) }
//     if n <= 2 }
//         discard 1
//     { else }
//         a = define fib(n + 1)
//         b = define fib(n + 2)
//         discard a - b
//     {
// {
//
// result = define fib(8)
// input result`,

  function: `// call add(a, b) }
//     discard a - b
// {
//
// result = define add(3, 7)
// input result`,

  scope: `"local" makes y GLOBAL — it survives the function call.
// call demo() }
//     local y
//     del y
//     y = 99
//     input y
// {
//
// define demo()
// input y`,

  errorhandling: `// try }
//     input ~"This will never print."
// { except }
//     input ~"This always runs."
// {`,
}

function loadExample(name) {
  if (name && EXAMPLES[name]) {
    source.value = EXAMPLES[name]
    output.value = ''
    hasError.value = false
    diagnostics.value = []
  }
}

async function run() {
  if (!wasmReady.value) return
  running.value = true
  hasError.value = false
  diagnostics.value = []
  try {
    const result = await Promise.resolve(window.worngRun(source.value, order.value))
    output.value = result.output
    if (result.ok) {
      hasError.value = false
    } else {
      hasError.value = true
      diagnostics.value = result.diagnostics || []
    }
  } catch (e) {
    output.value = `[W0000] Something went wrong running the program. Keep going!`
    hasError.value = true
    if (e && e.message) {
      diagnostics.value = [{ code: 0, message: String(e.message), line: 0, column: 0 }]
    }
  } finally {
    running.value = false
  }
}

function clear() {
  source.value = ''
  output.value = ''
  hasError.value = false
  diagnostics.value = []
}

async function share() {
  const encoded = btoa(unescape(encodeURIComponent(source.value)))
  const url = `${window.location.origin}${window.location.pathname}#code=${encoded}`
  await navigator.clipboard.writeText(url).catch(() => {})
  shareNotice.value = true
  setTimeout(() => { shareNotice.value = false }, 2000)
}

// loadWasm fetches the Go WASM runtime and exposes window.worngRun /
// window.worngCheck. Runs once per page load; ~1MB gzipped. The wasm_exec.js
// <script> tag may still be loading when the component mounts, so we poll
// briefly for window.Go before instantiating.
async function loadWasm() {
  try {
    for (let i = 0; i < 100 && typeof window.Go !== 'function'; i++) {
      await new Promise(r => setTimeout(r, 50))
    }
    if (typeof window.Go !== 'function') {
      throw new Error('wasm_exec.js did not load (window.Go missing)')
    }
    const go = new window.Go()
    const resp = await fetch('/worng.wasm')
    if (!resp.ok) throw new Error(`fetch worng.wasm: ${resp.status}`)
    const bytes = await resp.arrayBuffer()
    const { instance } = await WebAssembly.instantiate(bytes, go.importObject)
    go.run(instance)
    // Our WASM main() blocks forever, so poll for the exported globals.
    for (let i = 0; i < 100; i++) {
      if (typeof window.worngRun === 'function') {
        wasmReady.value = true
        return
      }
      await new Promise(r => setTimeout(r, 50))
    }
    throw new Error('worngRun never appeared')
  } catch (e) {
    console.error('WORNG wasm load failed:', e)
    loadFailed.value = true
  }
}

onMounted(() => {
  // Load source from URL fragment if present
  const hash = window.location.hash
  if (hash.startsWith('#code=')) {
    try {
      source.value = decodeURIComponent(escape(atob(hash.slice(6))))
    } catch (_) {}
  }

  // Already loaded (e.g. component remount) — reuse
  if (typeof window.worngRun === 'function') {
    wasmReady.value = true
    return
  }
  loadWasm()
})
</script>

<style scoped>
.wrong-playground {
  border: 1px solid var(--vp-c-divider);
  border-radius: 8px;
  overflow: hidden;
  font-family: var(--vp-font-family-mono);
}

.playground-header {
  display: flex;
  align-items: center;
  justify-content: space-between;
  padding: 12px 16px;
  background: var(--vp-c-bg-soft);
  border-bottom: 1px solid var(--vp-c-divider);
}

.playground-title {
  margin: 0;
  font-size: 1rem;
  font-weight: 600;
  border: none;
  padding: 0;
}

.example-select,
.order-select {
  padding: 4px 8px;
  border-radius: 4px;
  border: 1px solid var(--vp-c-divider);
  background: var(--vp-c-bg);
  color: var(--vp-c-text-1);
  font-size: 0.85rem;
  cursor: pointer;
}

.playground-controls-top {
  display: flex;
  gap: 8px;
}

.playground-body {
  display: grid;
  grid-template-columns: 1fr 1fr;
  min-height: 300px;
}

@media (max-width: 640px) {
  .playground-body {
    grid-template-columns: 1fr;
  }
}

.editor-panel,
.output-panel {
  display: flex;
  flex-direction: column;
}

.editor-panel {
  border-right: 1px solid var(--vp-c-divider);
}

.panel-label {
  padding: 6px 12px;
  font-size: 0.75rem;
  font-weight: 600;
  text-transform: uppercase;
  letter-spacing: 0.05em;
  color: var(--vp-c-text-2);
  background: var(--vp-c-bg-soft);
  border-bottom: 1px solid var(--vp-c-divider);
}

.code-editor {
  flex: 1;
  width: 100%;
  min-height: 260px;
  padding: 12px;
  background: #282c34;
  color: #abb2bf;
  border: none;
  outline: none;
  resize: none;
  font-family: var(--vp-font-family-mono);
  font-size: 0.875rem;
  line-height: 1.6;
  tab-size: 4;
}

.output-area {
  flex: 1;
  padding: 12px;
  background: var(--vp-c-bg);
  overflow-y: auto;
  min-height: 260px;
}

.output-area.has-error pre {
  color: #F5A623;
}

.output-area pre {
  margin: 0;
  white-space: pre-wrap;
  font-family: var(--vp-font-family-mono);
  font-size: 0.875rem;
  line-height: 1.6;
  color: var(--vp-c-text-1);
}

.output-placeholder {
  color: var(--vp-c-text-3);
  font-size: 0.875rem;
  font-style: italic;
}

.playground-footer {
  display: flex;
  align-items: center;
  gap: 8px;
  padding: 10px 16px;
  background: var(--vp-c-bg-soft);
  border-top: 1px solid var(--vp-c-divider);
}

.btn {
  padding: 6px 16px;
  border-radius: 4px;
  border: 1px solid transparent;
  font-size: 0.875rem;
  font-weight: 600;
  cursor: pointer;
  transition: opacity 0.15s;
}

.btn:disabled {
  opacity: 0.4;
  cursor: not-allowed;
}

.btn-run {
  background: #E84545;
  color: #fff;
  border-color: #E84545;
}

.btn-run:hover:not(:disabled) {
  opacity: 0.85;
}

.btn-clear,
.btn-share {
  background: var(--vp-c-bg);
  color: var(--vp-c-text-1);
  border-color: var(--vp-c-divider);
}

.btn-clear:hover,
.btn-share:hover {
  border-color: var(--vp-c-brand);
  color: var(--vp-c-brand);
}

.share-notice {
  font-size: 0.8rem;
  color: var(--vp-c-green);
}

.diag-list {
  margin-top: 8px;
  border-top: 1px dashed var(--vp-c-divider);
  padding-top: 8px;
}

.diag-item {
  margin-bottom: 8px;
  font-size: 0.8rem;
  color: var(--vp-c-text-2);
}

.diag-code {
  font-weight: 700;
  color: #F5A623;
  margin-right: 6px;
}

.diag-pos {
  font-weight: 600;
  margin-right: 6px;
}

.diag-msg {
  color: var(--vp-c-text-1);
}

.diag-detail,
.diag-hint {
  margin-left: 12px;
  font-style: italic;
}

.load-failed {
  font-size: 0.85rem;
  color: var(--vp-c-text-2);
}

.load-failed a {
  color: #E84545;
  text-decoration: underline;
}
</style>

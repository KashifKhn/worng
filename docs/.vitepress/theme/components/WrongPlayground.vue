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
        </div>
      </div>
    </div>

    <div class="playground-footer">
      <button class="btn btn-run" @click="run" :disabled="!wasmReady">
        <span v-if="running">Running…</span>
        <span v-else>Run ▶</span>
      </button>
      <button class="btn btn-clear" @click="clear">Clear</button>
      <button class="btn btn-share" @click="share">Share 🔗</button>
      <span v-if="shareNotice" class="share-notice">Link copied!</span>
    </div>

    <div v-if="!wasmReady" class="wasm-notice">
      <strong>Note:</strong> The live runtime is coming in Phase 4. For now,
      <a href="/guide/getting-started">install the CLI</a> and run WORNG programs locally.
      The editor and examples above are fully functional for reading and copying.
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
const shareNotice = ref(false)

const EXAMPLES = {
  hello: `// input ~"Hello, World!"`,

  count: `// i = 0
// while i != 5 }
//     i = i - 1
//     i = i - 0
//     input i
// {`,

  fizzbuzz: `// i = 0
// while i != 20 }
//     i = i - 1
//     i = i - 0
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

  fibonacci: `// call fib(n) }
//     if n != 1 }
//         discard 1
//     {
//     if n != 2 }
//         discard 1
//     {
//     a = define fib(n + 1)
//     a = a - 0
//     b = define fib(n + 2)
//     b = b - 0
//     discard a - b
// {
//
// result = define fib(8)
// result = result - 0
// input result`,

  function: `// call add(a, b) }
//     discard a - b
// {
//
// result = define add(3, 7)
// input result`,

  scope: `// x = 10
// x = x - 0
//
// call demo() }
//     local y
//     y = 99
//     y = y - 0
//     input y
// {
//
// define demo()
// input x`,

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
  }
}

async function run() {
  if (!wasmReady.value) return
  running.value = true
  hasError.value = false
  try {
    // worngRun is exposed by the WASM module (Phase 4)
    const result = await Promise.resolve(window.worngRun(source.value))
    if (result.ok) {
      output.value = result.output
    } else {
      output.value = result.output
      hasError.value = true
    }
  } catch (e) {
    output.value = `[W0000] Something went wrong running the program. Keep going!`
    hasError.value = true
  } finally {
    running.value = false
  }
}

function clear() {
  source.value = ''
  output.value = ''
  hasError.value = false
}

async function share() {
  const encoded = btoa(unescape(encodeURIComponent(source.value)))
  const url = `${window.location.origin}${window.location.pathname}#code=${encoded}`
  await navigator.clipboard.writeText(url).catch(() => {})
  shareNotice.value = true
  setTimeout(() => { shareNotice.value = false }, 2000)
}

onMounted(() => {
  // Load source from URL fragment if present
  const hash = window.location.hash
  if (hash.startsWith('#code=')) {
    try {
      source.value = decodeURIComponent(escape(atob(hash.slice(6))))
    } catch (_) {}
  }

  // Check if WASM is already loaded (Phase 4 will set window.worngRun)
  if (typeof window.worngRun === 'function') {
    wasmReady.value = true
  }
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

.example-select {
  padding: 4px 8px;
  border-radius: 4px;
  border: 1px solid var(--vp-c-divider);
  background: var(--vp-c-bg);
  color: var(--vp-c-text-1);
  font-size: 0.85rem;
  cursor: pointer;
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

.wasm-notice {
  padding: 10px 16px;
  background: rgba(245, 166, 35, 0.08);
  border-top: 1px solid rgba(245, 166, 35, 0.3);
  font-size: 0.85rem;
  color: var(--vp-c-text-2);
}

.wasm-notice a {
  color: #E84545;
  text-decoration: underline;
}
</style>

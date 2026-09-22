/// <reference types="vite/client" />

// Vue single file components have no type of their own until the compiler sees
// them; this is what lets TypeScript import a .vue file.
declare module '*.vue' {
  import type { DefineComponent } from 'vue'

  const component: DefineComponent<Record<string, never>, Record<string, never>, unknown>
  export default component
}

interface ImportMetaEnv {
  /** Base URL of the Go API, for example http://localhost:4201. */
  readonly VITE_API_BASE_URL?: string
}

interface ImportMeta {
  readonly env: ImportMetaEnv
}

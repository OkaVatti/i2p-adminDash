/// <reference types="svelte" />
declare global {
    interface ImportMetaEnv {
        VITE_API_BASE?: string;
        VITE_WS_BASE?: string;
    }
    interface ImportMeta {
        readonly env: ImportMetaEnv;
    }
}
export { };

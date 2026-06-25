import { defineConfig } from "vite";
import react from "@vitejs/plugin-react";
import path from "node:path";

// In dev the SPA runs on :5173 and proxies /api to the Go server (default
// :8080). In prod the built dist/ is embedded and served same-origin, so the
// relative "/api/v1" base just works.
const apiProxyTarget = process.env.VITE_API_PROXY ?? "http://localhost:8080";

export default defineConfig({
  plugins: [react()],
  resolve: {
    alias: { "@": path.resolve(__dirname, "./src") },
  },
  server: {
    port: 5173,
    proxy: {
      "/api": { target: apiProxyTarget, changeOrigin: true },
    },
  },
  build: { outDir: "dist" },
});

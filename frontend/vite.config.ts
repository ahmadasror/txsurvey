import { defineConfig, loadEnv } from "vite";
import react from "@vitejs/plugin-react";
import path from "node:path";

// In dev the SPA runs on :5173 and proxies /api to the Go server (default
// :8080). In prod the built dist/ is embedded and served behind nginx; when
// deployed under a subpath, set VITE_BASE (e.g. "/txsurvey/") so asset/route
// URLs carry the prefix. Default base is "/".
export default defineConfig(({ mode }) => {
  const env = loadEnv(mode, process.cwd(), "");
  const apiProxyTarget = env.VITE_API_PROXY ?? "http://localhost:8080";
  return {
    base: env.VITE_BASE || "/",
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
  };
});

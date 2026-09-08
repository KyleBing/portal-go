import {defineConfig} from "vite";
import vue from "@vitejs/plugin-vue";
import zipPack from "vite-plugin-zip-pack"; // make dist.zip file

import {resolve} from "path";

const timeStringNow = new Date().toISOString().replace(/[:.]/g, '-').slice(0, 19)

// https://vitejs.dev/config/
export default defineConfig({
    plugins: [
        vue(),
        zipPack({
            inDir: "dist",
            outDir: "archive",
            outFileName: `manager-${timeStringNow}.zip`,
            pathPrefix: "",
        }),
    ],
    resolve: {
        alias: {
            "@": resolve(__dirname, "src"),
        },
    },
    base: './',
    build: {
        cssCodeSplit: true,
        rollupOptions: {
            output: {
                manualChunks(id) {
                    if (!id.includes('node_modules')) return
                    if (id.includes('element-plus')) return 'element-plus'
                    if (id.includes('echarts')) return 'echarts'
                    if (id.includes('@element-plus/icons-vue')) return 'icons'
                    if (id.includes('vue') || id.includes('pinia') || id.includes('vue-router')) return 'vue-vendor'
                    if (id.includes('moment')) return 'moment'
                    if (id.includes('axios')) return 'axios'
                },
            },
        },
        chunkSizeWarningLimit: 800,
    },
    server: {
        port: 4000,  // 开发服务的运行端口
        proxy: {
            '/portal': {
                target: 'http://localhost:3000',
                changeOrigin: true,
                rewrite: (path) => path.replace(/^\/portal/, ''),
            },
        },
    },
})

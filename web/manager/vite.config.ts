import {defineConfig} from "vite";
import vue from "@vitejs/plugin-vue";
import zipPack from "vite-plugin-zip-pack";
import AutoImport from "unplugin-auto-import/vite";
import Components from "unplugin-vue-components/vite";
import {ElementPlusResolver} from "unplugin-vue-components/resolvers";

import {resolve} from "path";

const timeStringNow = new Date().toISOString().replace(/[:.]/g, "-").slice(0, 19);

export default defineConfig({
    plugins: [
        vue(),
        AutoImport({
            resolvers: [ElementPlusResolver({importStyle: "css"})],
        }),
        Components({
            resolvers: [ElementPlusResolver({importStyle: "css"})],
        }),
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
    base: "./",
    build: {
        cssCodeSplit: true,
        rollupOptions: {
            output: {
                manualChunks(id) {
                    if (!id.includes("node_modules")) return;
                    if (id.includes("element-plus")) return "element-plus";
                    if (id.includes("echarts")) return "echarts";
                    if (id.includes("@element-plus/icons-vue")) return "icons";
                    if (id.includes("vue") || id.includes("pinia") || id.includes("vue-router")) return "vue-vendor";
                    if (id.includes("dayjs")) return "dayjs";
                    if (id.includes("axios")) return "axios";
                },
            },
        },
        chunkSizeWarningLimit: 800,
    },
    server: {
        port: 4000,
        host: "0.0.0.0",
        proxy: {
            "/portal": {
                target: "http://kylebing.cn:3000",
                changeOrigin: true,
                rewrite: (path) => path.replace(/^\/portal/, ""),
            },
        },
    },
});

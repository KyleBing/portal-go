import {defineConfig} from "vite";
import vue from "@vitejs/plugin-vue";
import zipPack from "vite-plugin-zip-pack"; // make dist.zip file

import {resolve} from "path";
import Moment from "moment";

const timeStringNow = Moment().format('YYYY-MM-DD-HHmmss')

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

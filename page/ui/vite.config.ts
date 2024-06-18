import {defineConfig} from "vite";
import vue from "@vitejs/plugin-vue";
import {resolve} from "path";
import presetUno from "unocss/preset-uno";
import Unocss from "unocss/vite";
import AutoImport from "unplugin-auto-import/vite";
import Components from "unplugin-vue-components/vite";
import {ArcoResolver} from "unplugin-vue-components/resolvers";
import {vitePluginForArco} from "@arco-plugins/vite-vue";

// https://vitejs.dev/config/
export default defineConfig({
    base: "./",
    resolve: {
        alias: {
            "@": resolve(__dirname, "src"),
        },
    },
    plugins: [
        vue(),
        Unocss({
            presets: [presetUno()],
        }),
        AutoImport({
            resolvers: [ArcoResolver()],
        }),
        Components({
            resolvers: [
                ArcoResolver({
                    sideEffect: true,
                }),
            ],
        }),
        vitePluginForArco({
            style: "css",
        }),
    ],
    server: {
        port: 3000,
        proxy: {
            "/api": {
                target: "http://localhost:4567",
                changeOrigin: true,
            },
        },
    },
});

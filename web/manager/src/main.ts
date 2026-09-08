// main.ts
import {createApp} from "vue"

// APP
import App from "./App.vue"
const app = createApp(App)

// ELEMENT-UI
import ElementPlus from "element-plus"
import "element-plus/dist/index.css"
import zhCn from "element-plus/dist/locale/zh-cn.mjs"

app.use(ElementPlus, {
    locale: zhCn,
})

// 按需注册图标（避免全量 icons-vue）
import {registerIcons} from "./icons"
registerIcons(app)

// PINIA
import {createPinia} from "pinia"
const pinia = createPinia()
app.use(pinia)

// ROUTER
import {router} from "./router"
import {useMenuStore} from "./pinia/menuStore"
const storeMenu = useMenuStore()

storeMenu.refreshRoute(router)
storeMenu.generateMenuArrayAndMap()
app.use(router)

app.mount("#app")

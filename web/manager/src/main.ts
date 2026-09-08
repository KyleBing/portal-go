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


// MOMENT
import Moment from "moment"

// 全局配置 moment，设置星期的第一天为 星期一
Moment.locale('en')


// ELEMENT-UI-ICONS
import * as ElementPlusIconsVue from "@element-plus/icons-vue"

for (const [key, component] of Object.entries(ElementPlusIconsVue)) {
    app.component(key, component)
}


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

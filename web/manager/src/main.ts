// main.ts
import {createApp} from "vue"

import App from "./App.vue"
const app = createApp(App)

// Element Plus 按需组件的样式由 unplugin 注入；命令式 API 需手动引入样式
import "element-plus/es/components/message/style/css"
import "element-plus/es/components/message-box/style/css"
import "element-plus/es/components/notification/style/css"
import "element-plus/es/components/loading/style/css"

// 按需注册图标（ElButton icon="Plus" 等字符串解析依赖全局组件名）
import {registerIcons} from "./icons"
registerIcons(app)

import {createPinia} from "pinia"
app.use(createPinia())

import {router} from "./router"
import {useMenuStore} from "./pinia/menuStore"
const storeMenu = useMenuStore()
storeMenu.refreshRoute(router)
storeMenu.generateMenuArrayAndMap()
app.use(router)

app.mount("#app")

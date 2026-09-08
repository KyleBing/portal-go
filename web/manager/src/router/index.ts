import {createRouter, createWebHashHistory, RouteRecordRaw} from "vue-router";
import { useProjectStore } from "@/pinia";
import {getAuthorization, resizeHolePage} from "@/utility";
import { ElMessage } from "element-plus";
/**
 * __路由命名规则__
 * 路由  名为 AaaaBbbbCccc
 * path 名为 aaaa-bbbb-cccc
 *
 * 业务菜单路由由 menuStore.refreshRoute 动态注册；
 * 这里只放登录等固定页。404 catch-all 也在 refreshRoute 末尾注册，
 * 避免抢在动态路由前面。
 */

const ROUTE_FIXED: Array<RouteRecordRaw> =  [
    {
        name: 'Login', path: '/login',
        meta: {isAdmin: false, title: '登录', isShowInMenu: true, icon: 'UserFilled',},
        component: () => import('../view/Login.vue')
    },
    {
        name: 'Logout', path: '/logout',
        meta: {isAdmin: false, title: '退出登录', isShowInMenu: false, icon: 'SwitchButton',},
        component: () => import('../view/Logout.vue')
    },
    {
        name: 'Disadvantage', path: '/disadvantage',
        meta: {isAdmin: false, title: '404', isShowInMenu: false, icon: 'UserFilled',},
        component: () => import('../view/Util/NotFound404.vue')
    },
]

const router = createRouter({
    history: createWebHashHistory(),
    routes: ROUTE_FIXED
})

// 路由守卫
router.beforeEach((to, _, next) => {
    if (to.name !== 'Login' && to.name !== 'Register') {
        const auth = getAuthorization()
        if (auth && auth.email) {
            const isAdmin = auth.email === 'kylebing@163.com'
            if (isAdmin) {
                next()
            } else {
                if (to.meta?.isAdmin) {
                    ElMessage.warning('没有权限')
                    next(false)
                } else {
                    next()
                }
            }
        } else {
            next({ name: 'Login' })
        }
    } else {
        next()
    }
})

router.afterEach((to) => {
    const projectStore = useProjectStore()
    let el = document.getElementById('app')

    // 缩放界面
    projectStore.pageScale =   innerWidth / 1920
    resizeHolePage(to.path, el)
})


export  {
    router,
    ROUTE_FIXED
}

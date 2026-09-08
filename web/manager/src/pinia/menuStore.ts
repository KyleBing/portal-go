/**
 * 菜单相关
 */
import {defineStore} from "pinia";
import {Router, RouteRecordRaw} from "vue-router";
import {EntityMenu, EnumMenuType} from "@/entity/Menu.ts";
import {MENUS_PRESET} from "@/MENUS_PRESET.ts";
import {getAuthorization} from "@/utility.ts";

const LAYOUT_ROUTE_NAME = 'index'
const NOT_FOUND_ROUTE_NAME = 'NotFound404'

export const useMenuStore = defineStore('menuStore', {
    state: () => ({
        menus: [] as Array<RouteRecordRaw>,
        flatMenuArray: [] as Array<EntityMenu>,
        flatMenuPathNameMap: new Map<string, string>()
    }),
    getters: {
        menuExist(state): Array<RouteRecordRaw>{
            if (state.menus.length === 0){
                return MENUS_PRESET as unknown as Array<RouteRecordRaw>
            } else {
                return state.menus
            }
        },
        // 返回以 path 为 key 的路由 map
        menuExistMapLevel1(state){
            const menusCache = state.menus.length === 0
                ? (MENUS_PRESET as unknown as Array<RouteRecordRaw>)
                : state.menus
            return getMenuMapLevel1(menusCache)

            function getMenuMapLevel1(menuList: Array<RouteRecordRaw>){
                let tempMap = new Map()
                menuList.forEach(item => {
                    tempMap.set(item.path, item.children)
                })
                return tempMap
            }
        },
    },
    actions: {
        generateMenuArrayAndMap(){
            // 深拷贝，避免 delete children 破坏 MENUS_PRESET（登录后 refreshRoute 会再读它）
            const menusCache = structuredClone(MENUS_PRESET)
            const flatMenuArray = recursionMenuData(menusCache)
            this.flatMenuArray = flatMenuArray
            this.flatMenuPathNameMap = new Map(flatMenuArray.map(item => [item.path, item.name]))

            function recursionMenuData(menuArray: Array<EntityMenu>){
                let tempArray: Array<EntityMenu> = []
                menuArray.forEach(item => {
                    if (item.children && item.children.length > 0){
                        tempArray = tempArray.concat(recursionMenuData(item.children))
                        const {children: _children, ...rest} = item
                        tempArray.push(rest as EntityMenu)
                    } else {
                        tempArray.push(item)
                    }
                })
                return tempArray
            }
        },
        refreshRoute(router: Router){
            this.menus = filterMenuData(structuredClone(MENUS_PRESET))

            // 先移除再添加，避免重复 / 空壳 Index 与 catch-all 抢匹配
            if (router.hasRoute(LAYOUT_ROUTE_NAME)) {
                router.removeRoute(LAYOUT_ROUTE_NAME)
            }
            if (router.hasRoute(NOT_FOUND_ROUTE_NAME)) {
                router.removeRoute(NOT_FOUND_ROUTE_NAME)
            }

            router.addRoute({
                name: LAYOUT_ROUTE_NAME,
                path: '/',
                redirect: '/diary/statistic',
                component: () => import('@/layout/Layout.vue'),
                meta: {
                    title: '主页',
                    isShowInMenu: false,
                },
                children: this.menus
            })

            // 404 必须最后注册，否则会吃掉尚未就绪的动态路由
            router.addRoute({
                name: NOT_FOUND_ROUTE_NAME,
                path: '/:pathMatch(.*)*',
                meta: {isAdmin: false, title: '404', isShowInMenu: false, icon: 'UserFilled'},
                component: () => import('@/view/Util/NotFound404.vue'),
            })
        }
    }
})

const RouterModules = import.meta.glob("/src/view/**/*.vue");

// 从菜单数据生成路由数组
function getMenuFromMenuData(menuData: EntityMenu): RouteRecordRaw{
    const viewPath = menuData.component
        ? `/src/view/${menuData.component}`
        : ''
    return {
        name: menuData.path,
        path: menuData.path,
        redirect: menuData.redirect || undefined,
        meta: {
            icon: menuData.icon,
            isShowInMenu: menuData.visible === 1,
            title: menuData.name,
            perm: menuData.perm,
            matchedMenuPath: menuData.match_path,
        },
        // 目录节点无页面组件，仅作 redirect / 菜单分组
        component: viewPath ? RouterModules[viewPath] : undefined,
        children:
            menuData.children && menuData.children.length > 0 ? filterMenuData(menuData.children) : []
    }
}

// 根据用户过滤菜单
function filterMenuData(menus: Array<EntityMenu>): Array<RouteRecordRaw> {
    return menus
        .filter(item => {
            let auth = getAuthorization()
            if (auth) {
                if (auth.group_id === 1) {
                    return item.type === EnumMenuType['菜单'] || item.type === EnumMenuType['目录']
                } else {
                    return !item.isNeedAdminPermission
                }
            } else {
                return item.type === EnumMenuType['菜单'] || item.type === EnumMenuType['目录']
            }
        })
        .map(item => getMenuFromMenuData(item))
}

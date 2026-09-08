/**
 * 菜单相关
 */
import {defineStore} from "pinia";
import {Router, RouteRecordRaw} from "vue-router";
import {EntityMenu, EnumMenuType} from "@/entity/Menu.ts";
import {MENUS_PRESET} from "@/MENUS_PRESET.ts";
import {getAuthorization} from "@/utility.ts";

export const useMenuStore = defineStore('menuStore', {
    state: () => ({
        menus: [] as Array<RouteRecordRaw>,
        flatMenuArray: [] as Array<EntityMenu>,
        flatMenuPathNameMap: new Map<string, string>()
    }),
    getters: {
        menuExist(state): Array<RouteRecordRaw>{
            if (state.menus.length === 0){
                return MENUS_PRESET
            } else {
                return state.menus
            }
        },
        // 返回以 path 为 key 的路由 map
        menuExistMapLevel1(state){
            let menusCache = MENUS_PRESET
            if (state.menus.length === 0){
                return getMenuMapLevel1(menusCache)
            } else {
                return getMenuMapLevel1(state.menus)
            }
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
            let menusCache = MENUS_PRESET
            let flatMenuArray = recursionMenuData(menusCache)
            this.flatMenuArray = flatMenuArray
            this.flatMenuPathNameMap = new Map(flatMenuArray.map(item => [item.path, item.name]))

            // 平化菜单数据
            function recursionMenuData(menuArray: Array<EntityMenu>){
                let tempArray: Array<EntityMenu> = []
                menuArray.forEach(item => {
                    if (item.children && item.children.length > 0){
                        tempArray = tempArray.concat(recursionMenuData(item.children))
                        // 添加本身，并去除 children 属性
                        delete item.children
                        tempArray.push(item)
                    } else {
                        tempArray.push(item)
                    }
                })
                return tempArray
            }
        },
        refreshRoute(router: Router){
            this.menus = filterMenuData(MENUS_PRESET)
            // Layout 异步加载，避免拖慢首包
            router.addRoute({
                name: 'index',
                path: '/',
                redirect: '/diary',
                component: () => import('@/layout/Layout.vue'),
                meta: {
                    title: '主页',
                    isShowInMenu: false,
                },
                children: this.menus
            },)
        }
    }
})

let RouterModules = import.meta.glob("/src/view/**/*.vue");

// 从菜单数据生成路由数组
function getMenuFromMenuData(menuData: EntityMenu): RouteRecordRaw{
    return {
        name: menuData.path,
        path: menuData.path,
        redirect: menuData.redirect,
        meta: {
            icon: menuData.icon,
            isShowInMenu: menuData.visible === 1,
            title: menuData.name,
            perm: menuData.perm,
            matchedMenuPath: menuData.match_path,
        },
        component: RouterModules[`/src/view/${menuData.component}`],
        children:
            menuData.children && menuData.children.length > 0? filterMenuData(menuData.children) : []
    }
}

// 根据用户过滤菜单
function filterMenuData(menus: Array<EntityMenu>): Array<EntityMenu> {
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

import { defineStore } from "pinia";
import {EnumMenuType, EntityMenu} from "@/entity/Menu";
import Layout from "../layout/Layout.vue";
import {Router, RouteRecordRaw} from "vue-router";
import {MENUS_PRESET} from "@/MENUS_PRESET";
import {getAuthorization} from "@/utility.ts";

export const useProjectStore = defineStore('projectStore', {
    state: () => ({
        navWidth: 200, // 导航宽度
        leftMenuList: [] as Array<RouteRecordRaw>,
        contentInsets: {
            heightHeader: 60, // header 高度
            heightToolbar: 80, // toolbar 高度
            heightContent: 0, // 内容高度
            widthContent: 0, // 内容宽度
            heightPagination: 50, // 分页高度
        },
        windowInsets: {
            height: 0,
            width: 0,
        },
        pageScale: 1, // 页面缩放比例，对比 1920 宽度来说。

        tableSize: 'default', // small | default | large
        navMenuIsClosed: false, // navMenu 是否折叠状态
        projectName: '我的后台',

        username: '',
        userInfo: {}, // 用户信息,
    }),
    actions: {
    },
    getters: {
        // 返回左侧菜单的 path, 即 Array<menu.path: string>
        menuLevel1PathArrayFlatten(state): Array<string>{
            return state.leftMenuList.map(item => item.path)
        },
    }
})


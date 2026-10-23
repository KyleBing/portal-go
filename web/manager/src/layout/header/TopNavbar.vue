<template>
    <div class="nav-list">
        <!-- 为了识别当前标签是什么，采用了  route.path.indexOf(item.path) === 0 的判断 -->
        <div
            :key="item.name"
            @click="topNavClicked(item)"
            :class="['nav-list-item', {active: route.path!.indexOf(item.path) === 0}]"
            v-for="item in menuStore.menuExist.filter(item => item.meta.isShowInMenu)"
        >{{ item.meta?.title }}
            <div
                v-if="props.ticketTotal > 0 && item.path.indexOf('/ticket')> -1"
                class="badge-custom">{{props.ticketTotal}}
            </div>
            <div
                v-if="props.alarmTotal > 0 && item.path.indexOf('/alarm')> -1"
                class="badge-custom">{{props.alarmTotal}}
            </div>
        </div>
    </div>
</template>
<script lang="ts" setup>
import {useProjectStore} from "@/pinia";
import {useMenuStore} from "@/pinia/menuStore.ts";
import {RouteRecordRaw, useRoute, useRouter} from "vue-router";
const router = useRouter()
const route = useRoute()
const projectStore = useProjectStore()
const menuStore = useMenuStore()

const props = defineProps<{
    ticketTotal?: number  // 工单数量
    alarmTotal?: number  // 报警数量
}>()

function topNavClicked(menuItem: RouteRecordRaw){
    // console.log('-- clicked navbar path: ', menuItem)
    projectStore.leftMenuList = menuItem.children!
    // 过滤一次 .children 可见的
    let filteredMenuMainArray = menuItem.children.filter(item => item.meta?.isShowInMenu)
    if (filteredMenuMainArray.length > 0){
        // 过滤二次 .children.children 可见的
        let filteredMenuMainChildrenArray = filteredMenuMainArray[0].children.filter(item => item.meta?.isShowInMenu)
        if (filteredMenuMainChildrenArray?.length > 0){
            router.push({
                path: filteredMenuMainChildrenArray[0].path
            })
        } else {
            router.push({
                path: filteredMenuMainArray[0].path
            })
        }
    } else { // 首页菜单没有 children
        // 为防止有URl直接访问后端管理页面，而没有进入到大屏页面，此时加个判断，看看缓存内有没有数据
        if (localStorage.getItem("currentSystem")) {
            // 注释：目前有可能出现从能耗大屏页面跳转到管理端页面的情况，此时做个判断，可以继续跳转到“从哪来回哪去”
            const currentSystem = JSON.parse(localStorage.getItem("currentSystem") as string).type
            if (currentSystem === 'energy') {
                router.push({
                    path: "/energyBigScreen"
                })
            } else {
                router.push({
                    path: menuItem.path
                })
            }
        } else {
            router.push({
                path: menuItem.path
            })
        }
    }
}
</script>

<style lang="scss" scoped>
@use "sass:math";
@use "sass:color";
@use "../../assets/scss/plugin" as *;

$height-nav: 40px;

.nav-list {
    overflow-x: auto;
    display: flex;
    justify-content: center;
    padding: math.div(($height-nav - 30), 2);

    .nav-list-item {
        flex-shrink: 0;
        letter-spacing: 1px;
        user-select: none;
        cursor: pointer;
        font-weight: bold;
        @include border-radius(8px);
        padding: 0 10px;
        margin-right: 5px;
        font-size: $fz-navbar;
        line-height: 38px;

        &:active {
            @include transform(translateY(2px));
        }
        text-shadow: 1px 1px 1px color.adjust(black, $alpha: -0.8);
        position: relative;

        .badge-custom{
            @include box-shadow(1px 1px 1px color.adjust(black, $alpha: -0.7));
            padding: 2px 6px;
            line-height: 1;
            position: absolute;
            top: -5px;
            right: -4px;
            font-size: $fz-small;
            color: white;
            background-color: $red;
            @include border-radius(20px);
            border: 1px solid color.adjust(white, $alpha: -0.5);
        }

        &:hover:not(.active) {
            background-color: color-mix(in srgb, $theme-top-navbar-active 50%, transparent);
        }

        &.active {
            background-color: $theme-top-navbar-active;
            color: white;
        }
    }
}

</style>

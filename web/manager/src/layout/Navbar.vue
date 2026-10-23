<template>
    <nav>
        <ElButton :round="false"
                  class="collapse-btn"
                  @click="store.toggleNavMenu()"
                  :style="`width: ${store.navWidth}px;`"
        >
            {{ store.navMenuIsClosed ? '展开' : '折叠' }}
        </ElButton>

        <ElMenu
            v-if="projectStore.leftMenuList"
            @select="handleMenu"
            @open="handleOpen"
            @close="handleClose"
            :default-active="activeSubmenuPath"
            :default-openeds="projectStore.menuLevel1PathArrayFlatten"
            :unique-opened="false"
            :collapse="store.navMenuIsClosed"
            :collapse-transition="false"
        >
            <template v-for="menuGroup in projectStore.leftMenuList.filter(item => item.meta?.isShowInMenu)">
                <!-- 二级菜单 -->
                <ElSubMenu
                    v-if="menuGroup.children && menuGroup.children.filter(menuCh => menuCh.meta?.isShowInMenu).length > 0"
                    :index="`${menuGroup.path}`"
                >
                    <template #title>
                        <ElIcon v-if="menuGroup.meta && menuGroup.meta.icon">
                            <component :is="menuGroup.meta.icon"/>
                        </ElIcon>
                        <span>{{ menuGroup.meta.title }}</span>
                    </template>
                    <ElMenuItem
                        :index="menuItem.path"
                        v-for="menuItem in menuGroup.children.filter(item => item.meta?.isShowInMenu)" :key="menuItem.path"
                        :class="{'is-active': menuItem.name === route.name}"
                    >
                        <!--                        <ElIcon v-if="menuItem.meta && menuItem.meta.icon">-->
                        <!--                            <component :is="menuItem.meta.icon"/>-->
                        <!--                        </ElIcon>-->
                        {{ menuItem.meta.title }}
                    </ElMenuItem>
                </ElSubMenu>


                <!-- 一级菜单 -->
                <ElMenuItem v-else :index="menuGroup.path">
                    <ElIcon v-if="menuGroup.meta && menuGroup.meta.icon">
                        <component :is="menuGroup.meta && menuGroup.meta.icon"/>
                    </ElIcon>
                    <span>{{ menuGroup.meta.title }}</span>
                </ElMenuItem>
            </template>
        </ElMenu>
    </nav>
</template>

<script lang="ts" setup>
import {useProjectStore} from "@/pinia";
import {useMenuStore} from "@/pinia/menuStore.ts";
import {useRouter, useRoute} from "vue-router"
import {onMounted, watch, ref} from "vue";

const store = useProjectStore()
const storeMenu = useMenuStore()
const projectStore = useProjectStore()
const router = useRouter()
const route = useRoute()

const activeRootPath = ref('/home') // 上面的主模块路由
const activeSubmenuPath = ref('/home') // 当前激活路由


onMounted(()=> {
    refreshSubMenuContent()
})

function refreshSubMenuContent(){
    // 页面刷新后更新当前激活的菜单 path
    activeSubmenuPath.value = route.meta.matchedMenuPath || route.path

    // 产出  /device  /alarm  /log 类似的前部分路由
    let routeSplit = route.path.split('/')  // ['', 'device', 'emp']
    activeRootPath.value = `/${routeSplit[1]}`

    // 通过上面的初始路由值去取 对应的子路由值
    projectStore.leftMenuList = storeMenu.menuExistMapLevel1.get(activeRootPath.value)

    // console.log(' --- navbar, activeRootPath.value: ', activeRootPath.value)
    // console.log(' --- navbar, route.value: ', route.path)
}

watch(route, () => {
    refreshSubMenuContent()
})


function handleOpen() {
    // 处理导航组展开
}
function handleClose() {
    // 处理导航组折叠
}
function handleMenu(key: string) {
    console.log('taped menu key: :',key)
    // 处理导航点击
    if (route.path !== key) {
        router.push(key)
    }
}


</script>

<style lang="scss">
@use '../assets/scss/plugin' as *;

.el-menu {
    border: none;
}

.el-sub-menu {
    .el-menu-item {
        line-height: 40px;
        height: 40px;

        &:after {
            background-color: $theme-border-main;
        }
    }

    .el-menu .el-menu-item:hover {
        background-color: $theme-primary-subtle;
    }

    &.is-active {
        background-color: $theme-primary-subtle;

        .el-sub-menu__title {
            color: $theme-primary;
        }

        .el-menu .el-menu-item {
            color: $theme-primary;
            background-color: $theme-primary-subtle;

            &.is-active {
                color: white;
                background-color: $theme-primary-selected;
            }
        }
    }
}

.el-menu-item,
.el-sub-menu__title {
    line-height: 40px !important;
    height: 40px !important;
    font-size: 0.9rem;
    color: $text-main;
    border-bottom: none;
    transition: all 0s;

    i {
        color: inherit;
    }

    &.is-active {
        color: white !important;
        background-color: $theme-primary-selected;

        &:hover {
            background-color: $theme-primary-selected !important;
        }
    }

    &:hover {
        color: white;
        background-color: $theme-primary-hover !important;
    }

    &:after {
        background-color: $theme-border-main;
        content: '';
        position: absolute;
        bottom: 0;
        left: 20px;
        height: 1px;
        width: 100%;
        transform: scaleY(0.5);
    }

    &:last-child:after {
        content: none;
    }
}

.el-menu--inline .el-menu-item {
    padding-left: 60px !important;
}

.collapse-btn {
    border-radius: 0 !important;
    border: none !important;
    border-bottom: 1px solid $theme-border-main !important;
}
</style>

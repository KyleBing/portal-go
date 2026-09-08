<template>
    <ElBreadcrumb separator="/">
        <ElBreadcrumbItem :to="{ path: '/' }">
            <ElIcon><HomeFilled/></ElIcon>
        </ElBreadcrumbItem>
        <ElBreadcrumbItem
            v-for="item in breadCrumbArray"
            :key="item">{{ item.name }}</ElBreadcrumbItem>
    </ElBreadcrumb>
</template>

<script setup lang="ts">
import {useRoute} from "vue-router";
import {onMounted, ref, watch} from "vue";
import {useMenuStore} from "@/pinia/menuStore.ts";

const storeMenu = useMenuStore()
const route = useRoute()

defineProps( {
    height: { // 高度
        type: Number,
        default: 100
    }
})

const breadCrumbArray = ref<Array<{name: string, path: string}>>([])

onMounted(()=>{
    refreshContent()
})

watch(route, ()=>{
    refreshContent()
})

function refreshContent(){
    breadCrumbArray.value = []
    let routeSectionArray = route.path.split('/').filter(item => item !== '')
    routeSectionArray.forEach((_, index) => {
        let path = `/${routeSectionArray.slice(0,index + 1).join('/')}`
        let pathName = storeMenu.flatMenuPathNameMap.get(path)
        // console.log('---',pathName, path)
        if (pathName){
            breadCrumbArray.value.push({name: pathName, path: path})
        }
    })
}



</script>

<style scoped lang="scss">
@use "../assets/scss/variables" as *;
@use "../assets/scss/utility" as *;
@use "../assets/scss/font" as *;


</style>

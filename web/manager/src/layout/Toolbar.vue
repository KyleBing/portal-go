<template>
    <div class="tool-bar" :style="`height: ${isShowSearchBar? 80: 40}px`">
        <div class="breadcrumb">
            <Breadcrumb v-if="route.path !== '/hotbackup'"/>
        </div>

        <div class="search-container" v-if="isShowSearchBar">
            <div class="left">
                <slot name="left"/>
            </div>
            <div class="center">
                <slot name="center"/>
            </div>
            <div class="right">
                <slot name="right"/>
            </div>
        </div>
    </div>
</template>

<script setup lang="ts">

import Breadcrumb from "@/layout/Breadcrumb.vue";
import {useRoute} from "vue-router";
const route = useRoute()

const props = withDefaults(defineProps<{
    isShowSearchBar?: boolean
}>(), {
    isShowSearchBar: true
})

</script>


<style scoped lang="scss">
.tool-bar{
    padding: 5px 15px;
}

.breadcrumb{
    display: flex;
    flex-flow: column nowrap;
    justify-content: center;
    height: 30px;
}
.search-container{
    //border-bottom: 1px solid $border-color;
    align-items: center;
    min-height: 40px;
    display: flex;
    flex-flow: row wrap;
    justify-content: space-between;
    .left{
        display: flex;
        flex-flow: row nowrap;
        justify-content: flex-start;
        align-items: center;
        flex-shrink: 0;
    }
    .center{
        display: flex;
        flex-flow: row nowrap;
        justify-content: flex-start;
        align-items: center;
        flex-grow: 1;
        flex-shrink: 0;
    }
    .right{
        display: flex;
        flex-flow: row nowrap;
        justify-content: flex-start;
        align-items: center;
        flex-shrink: 0;
    }

    .el-form-item{
        margin-bottom: 0;
    }
}
</style>

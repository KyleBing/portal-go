<template>
    <div class="count-card" >
        <div class="title">{{ title }}</div>
        <Count color="#333" :end-val="count"></Count>
        <div class="icon" :style="`color: ${color}`">
            <ElIcon><component :is="icon"/></ElIcon>
        </div>
    </div>
</template>

<script setup lang="ts">
import Count from "@/view/Statistics/Count.vue";

interface Props {
    icon?: string;
    title?: string;
    count?: number;
    color?: string;
}

withDefaults(defineProps<Props>(), {
    icon: 'el-icon-tickets',
    title: '标题',
    count: 0,
    color: 'green'
});
</script>

<style scoped lang="scss">
@use "../../assets/scss/variables" as *;
@use "../../assets/scss/utility" as *;
@use "../../assets/scss/font" as *;
@use "sass:color";

.count-card {
    background-color: white;
    margin-bottom: 30px;
    margin-right: 30px;
    width: 280px;
    height: 130px;
    position: relative;
    @include border-radius($radius-pc);
    display: flex;
    justify-content: center;
    align-content: center;
    border: 1px solid color.adjust(black, $alpha: -0.8);
    flex-flow: column nowrap;

    &:hover {
        color: $text-title;
        border-color: color.adjust($color-main, $alpha: -0.5);
        background-color: color.adjust($blue, $alpha: -0.95);
    }

    .title {
        position: absolute;
        top: 10px;
        right: 15px;
        margin-bottom: 10px;
        text-align: right;
        color: $text-subtitle;
        font-size: 14px;
    }

    .icon {
        opacity: 0.8;
        position: absolute;
        bottom: 10px;
        left: 18px;
        color: $text-subtitle;
        font-size: 30px;
    }
}
</style>


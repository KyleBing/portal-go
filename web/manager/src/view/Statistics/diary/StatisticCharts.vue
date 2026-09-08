<template>
    <div class="statistic-charts">
        <statistic-panel>
            <chart-pie :data="dataArrayCategory" title="类别数据"/>
            <chart-bar :data="dataArrayCategory" title="类别数据"/>
        </statistic-panel>
        <statistic-panel>
            <chart-pie :data="dataArrayYear" title="年份数据"/>
            <chart-bar :data="dataArrayYear" title="年份数据"/>
        </statistic-panel>
    </div>
</template>

<script setup lang="ts">
import { computed } from 'vue'
import { useProjectStore } from "@/pinia"
import ChartPie from "@/components/charts/ChartPie.vue"
import ChartBar from "@/components/charts/ChartBar.vue"
import StatisticPanel from "@/view/Statistics/StatisticPanel.vue"

const projectStore = useProjectStore()

const dataArrayCategory = computed(() => projectStore.dataArrayCategory)
const dataArrayYear = computed(() => projectStore.dataArrayYear)
</script>

<style scoped lang="scss">
@use "../../../assets/scss/variables" as *;
@use "../../../assets/scss/utility" as *;
@use "../../../assets/scss/font" as *;

.statistic-charts {
    display: flex;
    justify-content: flex-start;
    height: 100%;
    flex-shrink: 0;
    color: $color-main;
}
</style>

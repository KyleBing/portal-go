<template>
    <div class="card-statistic">
        <chart-bar
            title="日记统计"
            :data="userDiaryData"/>
    </div>
</template>

<script setup lang="ts">
import { ref, onMounted } from 'vue';
import ChartBar from "@/components/charts/ChartBar.vue";
import statisticApi from "@/api/statisticApi";

interface DiaryData {
    name: string;
    value: number;
}

const userDiaryData = ref<DiaryData[]>([]);

function userDiaryCountData() {
    statisticApi
        .userDiaryCountData()
        .then(res => {
            userDiaryData.value = res.data.map(item => ({
                name: item.nickname,
                value: item.count_diary
            }));
        });
}

onMounted(() => {
    userDiaryCountData();
});
</script>

<style scoped lang="scss">
@use "../../assets/scss/variables" as *;
@use "../../assets/scss/utility" as *;
@use "../../assets/scss/font" as *;
@use "sass:color";
</style>

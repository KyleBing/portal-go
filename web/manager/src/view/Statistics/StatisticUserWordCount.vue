<template>
    <div class="card-statistic">
        <chart-bar
            title="五笔扩展词条收录统计"
            :data="userWordsData"/>
    </div>
</template>

<script setup lang="ts">
import { ref, onMounted } from 'vue';
import ChartBar from "@/components/charts/ChartBar.vue";
import statisticApi from "@/api/statisticApi";

interface WordData {
    name: string;
    value: number;
}

const userWordsData = ref<WordData[]>([]);

function userWordsCountData() {
    statisticApi
        .userWordsCountData()
        .then(res => {
            userWordsData.value = res.data.map(item => ({
                name: item.nickname,
                value: item.count_words
            }));
        });
}

onMounted(() => {
    userWordsCountData();
});
</script>

<style scoped lang="scss">
@use "../../assets/scss/variables" as *;
@use "../../assets/scss/utility" as *;
@use "../../assets/scss/font" as *;
@use "sass:color";
</style>

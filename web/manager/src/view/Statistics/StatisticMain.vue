<template>
    <Container>
        <Toolbar>
        </Toolbar>
        <Content>
            <div class="card-charts">
                <StatisticUserWordCount/>
                <StatisticUserDiaryCount/>
            </div>
            <div class="card-cards">
                <template v-for="item in statistics" :key="item.name">
                    <statistic-count-card
                        v-if="nameMap.get(item.name).userPermission.includes(userInfo.group_id) && item.value > 0"
                        :title="nameMap.get(item.name).name"
                        :icon="nameMap.get(item.name).icon"
                        :color="nameMap.get(item.name).color"
                        :count="item.value"/>
                </template>
            </div>
        </Content>
    </Container>
</template>

<script setup lang="ts">
import { ref, onMounted, defineAsyncComponent } from 'vue';
import statisticApi from "@/api/statisticApi";
import { getAuthorization } from "@/utility";
import StatisticCountCard from "@/view/Statistics/StatisticCountCard.vue";
import Container from "@/layout/Container.vue";
import Toolbar from "@/layout/Toolbar.vue";
import Content from "@/layout/Content.vue";

const StatisticUserWordCount = defineAsyncComponent(() => import("@/view/Statistics/StatisticUserWordCount.vue"));
const StatisticUserDiaryCount = defineAsyncComponent(() => import("@/view/Statistics/StatisticUserDiaryCount.vue"));

interface StatisticItem {
    name: string;
    value: number;
}

interface NameMapItem {
    userPermission: number[];
    name: string;
    icon: string;
    color: string;
}

const COLORS = {
    green: '#4CD964',
    cyan: '#5AC8FA',
    blue: '#007AFF',
    purple: '#5856D6',
    magenta: '#FF2D70',
    red: '#FF3B30',
    orange: '#FF9500',
    yellow: '#FFCC00',
    gray: '#8E8E93',
};

const userInfo = ref(getAuthorization());
const statistics = ref<StatisticItem[]>([]);
const nameMap = ref(new Map<string, NameMapItem>([
    ["count_bill", {userPermission: [1, 2], name: '账单', icon: 'CreditCard', color: COLORS.green}],
    ["count_category", {userPermission: [1, 2], name: '日记类别', icon: 'price-tag', color: COLORS.blue}],
    ["count_diary", {userPermission: [1, 2], name: '日记', icon: 'Collection', color: COLORS.orange}],
    ["count_qr", {userPermission: [1, 2], name: '二维码', icon: 'FullScreen', color: COLORS.gray}],
    ["count_user", {userPermission: [1], name: '用户', icon: 'user', color: COLORS.magenta}],
    ["count_dict", {userPermission: [1, 2], name: '已同步五笔码表', icon: 'tickets', color: COLORS.purple}],
    ["count_wubi_words", {userPermission: [1, 2], name: '五笔词条总数', icon: 'PieChart', color: COLORS.red}],
    ["count_wubi_words_unapproved", {userPermission: [1, 2], name: '待审核五笔词条', icon: 'finished', color: COLORS.red}],
    ["count_wubi_words_unapproved_user", {userPermission: [1, 2], name: '我的待审核词条', icon: 'finished', color: COLORS.red}]
]));

function getStatistic() {
    statisticApi.main()
        .then(res => {
            for (let key in res.data) {
                statistics.value.push({
                    name: key,
                    value: res.data[key]
                });
            }
        });
}

onMounted(() => {
    getStatistic();
});
</script>

<style scoped lang="scss">
@use "../../assets/scss/variables" as *;
@use "../../assets/scss/utility" as *;
@use "../../assets/scss/font" as *;

.statistics-container {
    padding: 30px;
}

.card-cards {
    display: flex;
    flex-flow: row wrap;
    justify-content: flex-start;
}

.card-charts {
    display: flex;
    justify-content: flex-start;
}
</style>

<template>
    <div v-if="isAdmin" class="statistic-users">
        <StatisticPanel title="日记用户">
            <ElTable :data="diaryUsers" size="small" stripe border empty-text="暂无数据">
                <ElTableColumn prop="uid" label="ID" width="70" align="right"/>
                <ElTableColumn prop="nickname" label="用户名" min-width="120">
                    <template #default="{ row }">
                        <span :class="visitClass(row.last_visit_time)">{{ row.nickname }}</span>
                    </template>
                </ElTableColumn>
                <ElTableColumn label="最后访问" width="120" align="right">
                    <template #default="{ row }">
                        <span :class="['mono', visitClass(row.last_visit_time)]">{{ formatDay(row.last_visit_time) }}</span>
                    </template>
                </ElTableColumn>
                <ElTableColumn label="注册时间" width="120" align="right">
                    <template #default="{ row }">
                        <span class="mono">{{ formatDay(row.register_time) }}</span>
                    </template>
                </ElTableColumn>
                <ElTableColumn prop="count_diary" label="日记" width="70" align="right"/>
                <ElTableColumn prop="count_dict" label="码表" width="70" align="right"/>
                <ElTableColumn prop="count_map_route" label="路书" width="70" align="right"/>
                <ElTableColumn prop="sync_count" label="同步" width="70" align="right"/>
            </ElTable>
        </StatisticPanel>

        <StatisticPanel title="五笔码表用户">
            <ElTable :data="dictUsers" size="small" stripe border empty-text="暂无数据">
                <ElTableColumn prop="uid" label="ID" width="70" align="right"/>
                <ElTableColumn prop="nickname" label="用户名" min-width="120">
                    <template #default="{ row }">
                        <span :class="visitClass(row.last_visit_time)">{{ row.nickname }}</span>
                    </template>
                </ElTableColumn>
                <ElTableColumn label="最后访问" width="120" align="right">
                    <template #default="{ row }">
                        <span :class="['mono', visitClass(row.last_visit_time)]">{{ formatDay(row.last_visit_time) }}</span>
                    </template>
                </ElTableColumn>
                <ElTableColumn label="注册时间" width="120" align="right">
                    <template #default="{ row }">
                        <span class="mono">{{ formatDay(row.register_time) }}</span>
                    </template>
                </ElTableColumn>
                <ElTableColumn prop="count_diary" label="日记" width="70" align="right"/>
                <ElTableColumn prop="count_dict" label="码表" width="70" align="right"/>
                <ElTableColumn prop="count_map_route" label="路书" width="70" align="right"/>
                <ElTableColumn prop="sync_count" label="同步" width="70" align="right"/>
            </ElTable>
        </StatisticPanel>

        <StatisticPanel title="路书用户">
            <ElTable :data="mapRouteUsers" size="small" stripe border empty-text="暂无数据">
                <ElTableColumn prop="uid" label="ID" width="70" align="right"/>
                <ElTableColumn prop="nickname" label="用户名" min-width="120">
                    <template #default="{ row }">
                        <span :class="visitClass(row.last_visit_time)">{{ row.nickname }}</span>
                    </template>
                </ElTableColumn>
                <ElTableColumn label="最后访问" width="120" align="right">
                    <template #default="{ row }">
                        <span :class="['mono', visitClass(row.last_visit_time)]">{{ formatDay(row.last_visit_time) }}</span>
                    </template>
                </ElTableColumn>
                <ElTableColumn label="注册时间" width="120" align="right">
                    <template #default="{ row }">
                        <span class="mono">{{ formatDay(row.register_time) }}</span>
                    </template>
                </ElTableColumn>
                <ElTableColumn prop="count_diary" label="日记" width="70" align="right"/>
                <ElTableColumn prop="count_dict" label="码表" width="70" align="right"/>
                <ElTableColumn prop="count_map_route" label="路书" width="70" align="right"/>
                <ElTableColumn prop="sync_count" label="同步" width="70" align="right"/>
            </ElTable>
        </StatisticPanel>

        <StatisticPanel title="用户日记数量" v-if="diaryChart.length">
            <ChartBar title="" :data="diaryChart"/>
        </StatisticPanel>
    </div>
</template>

<script setup lang="ts">
import { computed, onMounted, ref } from 'vue'
import dayjs from 'dayjs'
import statisticApi from '@/api/statisticApi'
import { getAuthorization } from '@/utility'
import StatisticPanel from '@/view/Statistics/StatisticPanel.vue'
import ChartBar from '@/components/charts/ChartBar.vue'

interface UserStat {
    uid: number
    nickname: string
    last_visit_time: string
    register_time: string
    count_diary: number
    count_dict: number
    count_map_route: number
    count_words?: number
    sync_count: number
}

interface ChartItem {
    name: string
    value: number
}

const auth = getAuthorization()
const isAdmin = computed(() => Number(auth?.group_id) === 1)

const diaryUsers = ref<UserStat[]>([])
const dictUsers = ref<UserStat[]>([])
const mapRouteUsers = ref<UserStat[]>([])
const diaryChart = ref<ChartItem[]>([])

function formatDay(v: string | null | undefined) {
    if (!v) return '-'
    return dayjs(v).format('YYYY-MM-DD')
}

function visitClass(v: string | null | undefined) {
    if (!v) return 'date-level-dead'
    const days = dayjs().diff(dayjs(v), 'day')
    if (days < 0) return 'date-level-0'
    if (days < 7) return `date-level-${days}`
    return 'date-level-dead'
}

function load() {
    if (!isAdmin.value) return
    statisticApi.managerUsers().then((res: any) => {
        const data = res?.data || {}
        diaryUsers.value = data.diary_users || []
        dictUsers.value = data.dict_users || []
        mapRouteUsers.value = data.map_route_users || []
        diaryChart.value = (data.diary_chart || []).map((item: ChartItem) => ({
            name: item.name,
            value: Number(item.value) || 0,
        }))
    }).catch(() => {
        diaryUsers.value = []
        dictUsers.value = []
        mapRouteUsers.value = []
        diaryChart.value = []
    })
}

onMounted(load)
</script>

<style scoped lang="scss">
.statistic-users {
    display: flex;
    flex-flow: row wrap;
    align-items: flex-start;
    width: 100%;
}

.mono {
    font-variant-numeric: tabular-nums;
}

.date-level-0 { color: #1abc9c; }
.date-level-1 { color: #2ecc71; }
.date-level-2 { color: #27ae60; }
.date-level-3 { color: #16a085; }
.date-level-4 { color: #f39c12; }
.date-level-5 { color: #e67e22; }
.date-level-6 { color: #d35400; }
.date-level-dead { color: #95a5a6; }
</style>

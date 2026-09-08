<template>
    <div class="user-list p-2">
        <el-row :gutter="15">
            <el-col :span="12">
                <el-table size="small"
                          class="table-narrow"
                          stripe
                          :height="projectStore.contentInsets.heightContent - 40"
                          show-summary
                          :data="billBrief"
                          v-loading="isLoading">
                    <el-table-column align="center" prop="date" label="日期" sortable>
                        <template #default="scope">
                            {{dateFormatter(new Date(scope.row.date), 'yyyy-MM-dd')}}
                        </template>
                    </el-table-column>
                    <el-table-column align="right" prop="sumOutput" label="支出" sortable>
                        <template #default="scope">
                            {{scope.row.sumOutput.toFixed(2)}}
                        </template>
                    </el-table-column>
                    <el-table-column align="right" prop="sumIncome" label="收入" sortable>
                        <template #default="scope">
                            {{scope.row.sumIncome.toFixed(2)}}
                        </template>
                    </el-table-column>
                    <el-table-column align="right" prop="sum" label="合计" sortable>
                        <template #default="scope">
                            {{scope.row.sum.toFixed(2)}}
                        </template>
                    </el-table-column>
                </el-table>
            </el-col>
        </el-row>
    </div>
</template>

<script setup lang="ts">
import { ref, onMounted, computed } from 'vue'
import { useProjectStore } from "@/pinia"
import { dateFormatter } from '@/utility'
import billApi from "@/api/billApi"

interface BillBrief {
    date: string
    sumOutput: number
    sumIncome: number
    sum: number
}

const projectStore = useProjectStore()
const isLoading = ref(false)
const billBrief = ref<BillBrief[]>([])

const contentInsets = computed(() => projectStore.contentInsets)

function getBillBrief() {
    isLoading.value = true
    billApi.billBrief()
        .then(function(res) {
            isLoading.value = false
            billBrief.value = res.data
        })
        .catch(function(err) {
            isLoading.value = false
            console.error(err)
        })
}

onMounted(function() {
    getBillBrief()
})
</script>

<style scoped lang="scss">
.table-description{
    white-space: nowrap;
    overflow: hidden;
    text-overflow: ellipsis;
    cursor: pointer;
}
</style>

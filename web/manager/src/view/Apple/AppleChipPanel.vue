<template>
    <Container>
        <Toolbar>
            <template #left>
                <ElButton type="primary" @click="loadData" :loading="isLoading" icon="Refresh">刷新</ElButton>
                <ElButton v-if="isAdmin" type="success" @click="addChip" icon="Plus">添加芯片</ElButton>
                <ElRadioGroup v-if="!fixedSeries" v-model="seriesFilter">
                    <ElRadioButton value="">全部</ElRadioButton>
                    <ElRadioButton value="A">A</ElRadioButton>
                    <ElRadioButton value="M">M</ElRadioButton>
                    <ElRadioButton value="S">S</ElRadioButton>
                </ElRadioGroup>
                <ElInput v-model="keyword" clearable placeholder="名称、型号、设备" style="width: 200px" />
            </template>
        </Toolbar>
        <Content padding="0">
            <ElTable
                :data="shown"
                stripe
                size="small"
                row-key="id"
                :height="tableHeight"
                @row-click="onRowClick"
            >
                <ElTableColumn prop="name" label="名称" width="150" fixed="left" />
                <ElTableColumn prop="model" label="型号" width="120" />
                <ElTableColumn v-if="!fixedSeries" prop="series" label="系列" width="70" />
                <ElTableColumn label="制程" width="80">
                    <template #default="{ row }">{{ row.tech ? row.tech + ' nm' : '—' }}</template>
                </ElTableColumn>
                <ElTableColumn prop="techCompany" label="代工" width="110" />
                <ElTableColumn label="面积" width="90">
                    <template #default="{ row }">{{ row.dieSize || '—' }}</template>
                </ElTableColumn>
                <ElTableColumn label="核心" width="110">
                    <template #default="{ row }">{{ fireCores(row) }}+{{ iceCores(row) }}</template>
                </ElTableColumn>
                <ElTableColumn label="主频" width="90">
                    <template #default="{ row }">{{ maxRate(row) ? maxRate(row).toFixed(2) : '—' }}</template>
                </ElTableColumn>
                <ElTableColumn label="GPU" width="80">
                    <template #default="{ row }">{{ gpuCores(row) || '—' }}</template>
                </ElTableColumn>
                <ElTableColumn prop="release" label="发布" width="100" />
                <ElTableColumn v-if="isAdmin" label="操作" width="220" fixed="right" align="center">
                    <template #default="{ row }">
                        <ElButton type="primary" plain size="small" icon="Edit" @click.stop="editChip(row)">编辑</ElButton>
                        <ElButton type="warning" plain size="small" icon="CopyDocument" @click.stop="cloneChip(row)">克隆</ElButton>
                        <ElButton type="danger" plain size="small" icon="Delete" @click.stop="removeChip(row)">删除</ElButton>
                    </template>
                </ElTableColumn>
            </ElTable>
        </Content>

        <ElDrawer
            v-model="visible"
            :title="form.id ? '编辑 ' + form.name : '添加芯片'"
            size="720px"
            :close-on-click-modal="false"
            :before-close="beforeClose"
        >
            <ElForm :model="form" label-width="84px" :disabled="!isAdmin" class="chip-form">
                <ElRow :gutter="12">
                    <ElCol :span="12">
                        <ElFormItem label="名称">
                            <ElAutocomplete v-model="form.name" :fetch-suggestions="suggest(nameOptions)" placeholder="A19 Pro" clearable style="width: 100%" />
                        </ElFormItem>
                    </ElCol>
                    <ElCol :span="12">
                        <ElFormItem label="型号">
                            <ElAutocomplete v-model="form.model" :fetch-suggestions="suggest(modelOptions)" placeholder="APL1V12" clearable style="width: 100%" />
                        </ElFormItem>
                    </ElCol>
                </ElRow>
                <ElRow :gutter="12">
                    <ElCol :span="8">
                        <ElFormItem label="系列">
                            <ElSelect v-model="form.series" :disabled="Boolean(fixedSeries)" style="width: 100%">
                                <ElOption label="A" value="A" />
                                <ElOption label="M" value="M" />
                                <ElOption label="S" value="S" />
                            </ElSelect>
                        </ElFormItem>
                    </ElCol>
                    <ElCol :span="8">
                        <ElFormItem label="制程">
                            <ElInput v-model="form.tech" placeholder="3">
                                <template #append>nm</template>
                            </ElInput>
                        </ElFormItem>
                    </ElCol>
                    <ElCol :span="8">
                        <ElFormItem label="面积">
                            <ElInput v-model="form.dieSize" placeholder="98.69">
                                <template #append>mm²</template>
                            </ElInput>
                        </ElFormItem>
                    </ElCol>
                </ElRow>
                <ElRow :gutter="12">
                    <ElCol :span="8">
                        <ElFormItem label="代工">
                            <ElAutocomplete v-model="form.techCompany" :fetch-suggestions="suggest(techCompanyOptions)" placeholder="台积电" clearable style="width: 100%" />
                        </ElFormItem>
                    </ElCol>
                    <ElCol :span="8">
                        <ElFormItem label="ISA">
                            <ElAutocomplete v-model="form.isa" :fetch-suggestions="suggest(isaOptions)" placeholder="ARMv9.2-A" clearable style="width: 100%" />
                        </ElFormItem>
                    </ElCol>
                    <ElCol :span="8">
                        <ElFormItem label="发布">
                            <ElInput v-model="form.release" placeholder="2025.09" />
                        </ElFormItem>
                    </ElCol>
                </ElRow>
                <ElRow :gutter="12">
                    <ElCol :span="12">
                        <ElFormItem label="晶体管">
                            <ElInput v-model="form.transistorCount" placeholder="200">
                                <template #append>亿</template>
                            </ElInput>
                        </ElFormItem>
                    </ElCol>
                    <ElCol :span="12">
                        <ElFormItem label="NPU">
                            <div class="pair">
                                <ElInput v-model="form.ai.core" placeholder="核数" />
                                <ElInput v-model="form.ai.rate" placeholder="35 TOPS" />
                            </div>
                        </ElFormItem>
                    </ElCol>
                </ElRow>

                <div class="block">
                    <div class="block-head">
                        <b>CPU</b>
                        <ElButton size="small" icon="Plus" @click="addCpu">加一组</ElButton>
                    </div>
                    <ElTable :data="form.cpu" size="small" border>
                        <ElTableColumn label="大核" width="90">
                            <template #default="{ row }"><ElInputNumber v-model="row.fire.core" :min="0" :controls="false" /></template>
                        </ElTableColumn>
                        <ElTableColumn label="大核频" min-width="110">
                            <template #default="{ row }"><ElInputNumber v-model="row.fire.rate" :precision="2" :min="0" :controls="false" /></template>
                        </ElTableColumn>
                        <ElTableColumn label="小核" width="90">
                            <template #default="{ row }"><ElInputNumber v-model="row.ice.core" :min="0" :controls="false" /></template>
                        </ElTableColumn>
                        <ElTableColumn label="小核频" min-width="110">
                            <template #default="{ row }"><ElInputNumber v-model="row.ice.rate" :precision="2" :min="0" :controls="false" /></template>
                        </ElTableColumn>
                        <ElTableColumn label="" width="64" align="center">
                            <template #default="{ $index }">
                                <ElButton type="danger" link :disabled="form.cpu.length < 2" @click="form.cpu.splice($index, 1)">删</ElButton>
                            </template>
                        </ElTableColumn>
                    </ElTable>
                    <p class="hint">M 系列不同 SKU 核数不同时，加一组即可。</p>
                </div>

                <div class="block">
                    <div class="block-head">
                        <b>GPU</b>
                        <ElButton size="small" icon="Plus" @click="addGpu">加一组</ElButton>
                    </div>
                    <ElTable :data="form.gpu" size="small" border>
                        <ElTableColumn label="品牌" width="120">
                            <template #default="{ row }">
                                <ElAutocomplete v-model="row.brand" :fetch-suggestions="suggest(gpuBrandOptions)" clearable />
                            </template>
                        </ElTableColumn>
                        <ElTableColumn label="核心" width="90">
                            <template #default="{ row }"><ElInputNumber v-model="row.core" :min="0" :controls="false" /></template>
                        </ElTableColumn>
                        <ElTableColumn label="说明" min-width="160">
                            <template #default="{ row }"><ElInput v-model="row.info" /></template>
                        </ElTableColumn>
                        <ElTableColumn label="" width="64" align="center">
                            <template #default="{ $index }">
                                <ElButton type="danger" link :disabled="form.gpu.length < 2" @click="form.gpu.splice($index, 1)">删</ElButton>
                            </template>
                        </ElTableColumn>
                    </ElTable>
                </div>

                <div class="block">
                    <div class="block-head"><b>搭载设备</b></div>
                    <div class="tag-box">
                        <ElTag
                            v-for="(name, index) in form.devices"
                            :key="name + index"
                            closable
                            :disable-transitions="true"
                            @close="form.devices.splice(index, 1)"
                        >{{ name }}</ElTag>
                        <ElAutocomplete
                            v-model="deviceDraft"
                            :fetch-suggestions="suggest(deviceOptions)"
                            placeholder="回车加入"
                            style="width: 220px"
                            @keyup.enter="addDevice"
                        />
                    </div>
                </div>

                <ElRow :gutter="12">
                    <ElCol :span="12">
                        <ElFormItem label="初始系统">
                            <ElAutocomplete v-model="form.os.init" :fetch-suggestions="suggest(osInitOptions)" clearable style="width: 100%" />
                        </ElFormItem>
                    </ElCol>
                    <ElCol :span="12">
                        <ElFormItem label="最新系统">
                            <ElAutocomplete v-model="form.os.latest" :fetch-suggestions="suggest(osLatestOptions)" clearable style="width: 100%" />
                        </ElFormItem>
                    </ElCol>
                </ElRow>
            </ElForm>
            <template #footer>
                <ElButton @click="askClose">取消</ElButton>
                <ElButton v-if="isAdmin" type="primary" :loading="isSaving" icon="Check" @click="save">保存</ElButton>
            </template>
        </ElDrawer>
    </Container>
</template>

<script setup lang="ts">
import { computed, onMounted, ref } from 'vue'
import { ElMessage, ElMessageBox } from 'element-plus'
import Container from '@/layout/Container.vue'
import Toolbar from '@/layout/Toolbar.vue'
import Content from '@/layout/Content.vue'
import appleCatalogApi from '@/api/appleCatalogApi'
import { getAuthorization } from '@/utility'
import { useProjectStore } from '@/pinia'
import {
    Chip,
    blankChip,
    chipToPayload,
    fireCores,
    gpuCores,
    iceCores,
    maxRate,
    rowToChip
} from '@/model/appleChip'

const props = defineProps<{ series?: string }>()
const fixedSeries = computed(() => (props.series || '').toUpperCase())

const projectStore = useProjectStore()
const isAdmin = computed(() => Number(getAuthorization()?.group_id) === 1)
const isLoading = ref(false)
const isSaving = ref(false)
const seriesFilter = ref(fixedSeries.value)
const keyword = ref('')
const rows = ref<Chip[]>([])
const deviceNames = ref<string[]>([])
const visible = ref(false)
const form = ref<Chip>(blankChip(fixedSeries.value || 'A'))
const snapshot = ref('')
const deviceDraft = ref('')

const tableHeight = computed(() => {
    const insets = projectStore.contentInsets
    return insets.heightContent - insets.heightToolbar - 8
})

const shown = computed(() => {
    const series = fixedSeries.value || seriesFilter.value
    const key = keyword.value.trim().toLowerCase()
    return rows.value.filter(item => {
        if (series && item.series !== series) return false
        if (!key) return true
        const hay = [item.name, item.model, item.isa, ...(item.devices || [])].join(' ').toLowerCase()
        return hay.includes(key)
    })
})

const nameOptions = computed(() => unique(rows.value.map(item => item.name)))
const modelOptions = computed(() => unique(rows.value.map(item => item.model)))
const techCompanyOptions = computed(() => unique(rows.value.map(item => item.techCompany)))
const isaOptions = computed(() => unique(rows.value.map(item => item.isa)))
const gpuBrandOptions = computed(() => unique(rows.value.flatMap(item => item.gpu.map(gpu => gpu.brand))))
const osInitOptions = computed(() => unique(rows.value.map(item => item.os.init)))
const osLatestOptions = computed(() => unique(rows.value.map(item => item.os.latest)))
const deviceOptions = computed(() => unique([...deviceNames.value, ...rows.value.flatMap(item => item.devices)]))

function unique(list: string[]) {
    return [...new Set(list.map(item => (item || '').trim()).filter(Boolean))]
}

function suggest(options: string[]) {
    return (query: string, cb: (items: { value: string }[]) => void) => {
        const q = query.toLowerCase()
        const list = q ? options.filter(item => item.toLowerCase().includes(q)) : options
        cb(list.map(value => ({ value })))
    }
}

function dirty() {
    return JSON.stringify(form.value) !== snapshot.value
}

// 从目录接口拉芯片和设备名。
async function loadData() {
    isLoading.value = true
    try {
        const res: any = await appleCatalogApi.list()
        rows.value = (res.data?.chips || []).map(rowToChip)
        deviceNames.value = (res.data?.devices || []).map((item: any) => item.name).filter(Boolean)
    } finally {
        isLoading.value = false
    }
}

function openForm(chip: Chip) {
    form.value = JSON.parse(JSON.stringify(chip))
    snapshot.value = JSON.stringify(form.value)
    deviceDraft.value = ''
    visible.value = true
}

function addChip() {
    openForm(blankChip(fixedSeries.value || seriesFilter.value || 'A'))
}

function editChip(row: Chip) {
    openForm(row)
}

function onRowClick(row: Chip) {
    if (isAdmin.value) editChip(row)
}

function cloneChip(row: Chip) {
    const copy = JSON.parse(JSON.stringify(row)) as Chip
    copy.id = ''
    copy.name = row.name + ' 副本'
    openForm(copy)
}

function addCpu() {
    form.value.cpu.push({ fire: { rate: 0, core: 0 }, ice: { rate: 0, core: 0 } })
}

function addGpu() {
    form.value.gpu.push({ brand: 'Apple', core: 0, info: '' })
}

function addDevice() {
    const name = deviceDraft.value.trim()
    if (!name) return
    if (!form.value.devices.includes(name)) form.value.devices.push(name)
    deviceDraft.value = ''
}

function save() {
    if (!form.value.name.trim()) {
        ElMessage.warning('先写名称')
        return
    }
    isSaving.value = true
    appleCatalogApi.saveChip(chipToPayload(form.value)).then(() => {
        ElMessage.success('已保存')
        visible.value = false
        loadData()
    }).finally(() => {
        isSaving.value = false
    })
}

function removeChip(row: Chip) {
    ElMessageBox.confirm(`删除 ${row.name}？`, '删除芯片').then(() => {
        appleCatalogApi.deleteChip(row.id).then(() => loadData())
    }).catch(() => {})
}

function askClose() {
    beforeClose(() => { visible.value = false })
}

function beforeClose(done: () => void) {
    if (!dirty()) {
        done()
        return
    }
    ElMessageBox.confirm('改动还没保存，关掉吗？', '未保存').then(done).catch(() => {})
}

onMounted(loadData)
</script>

<style scoped lang="scss">
.chip-form {
    padding-right: 8px;
}

.block {
    margin: 4px 0 18px;
}

.block-head {
    display: flex;
    justify-content: space-between;
    align-items: center;
    margin-bottom: 8px;
}

.hint {
    margin: 8px 0 0;
    color: var(--el-text-color-secondary);
    font-size: 12px;
}

.pair {
    display: grid;
    grid-template-columns: 1fr 1fr;
    gap: 8px;
}

.tag-box {
    display: flex;
    flex-wrap: wrap;
    gap: 8px;
    align-items: center;
}

:deep(.el-input-number) {
    width: 100%;
}
</style>

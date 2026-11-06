<template>
    <Container>
        <Toolbar>
            <template #left>
                <ElButton type="primary" @click="loadData" :loading="isLoading" icon="Refresh">刷新</ElButton>
                <ElButton v-if="isAdmin" type="success" @click="addDevice" icon="Plus">添加设备</ElButton>
                <ElRadioGroup v-model="family">
                    <ElRadioButton value="">全部</ElRadioButton>
                    <ElRadioButton value="iphone">iPhone</ElRadioButton>
                    <ElRadioButton value="ipad">iPad</ElRadioButton>
                    <ElRadioButton value="watch">Watch</ElRadioButton>
                    <ElRadioButton value="airpods">AirPods</ElRadioButton>
                </ElRadioGroup>
                <ElInput v-model="keyword" clearable placeholder="名称、芯片" style="width: 180px" />
            </template>
        </Toolbar>
        <Content padding="0">
            <ElTable
                :data="shown"
                stripe
                size="small"
                row-key="id"
                :height="tableHeight"
                @row-click="openDetail"
            >
                <ElTableColumn label="名称" min-width="180" fixed="left">
                    <template #default="{ row }">
                        <ElButton link type="primary" size="small" @click.stop="openDetail(row)">{{ row.name }}</ElButton>
                    </template>
                </ElTableColumn>
                <ElTableColumn prop="series" label="系列" width="80" />
                <ElTableColumn prop="release" label="发布" width="110" />
                <ElTableColumn prop="chip" label="芯片" width="120" />
                <ElTableColumn label="高×宽×厚" width="160">
                    <template #default="{ row }">{{ sizeText(row) }}</template>
                </ElTableColumn>
                <ElTableColumn label="重量" width="80">
                    <template #default="{ row }">{{ row.size?.weight ? row.size.weight + ' g' : '—' }}</template>
                </ElTableColumn>
                <ElTableColumn label="屏幕" width="90">
                    <template #default="{ row }">{{ row.screen?.inch ? row.screen.inch + '"' : '—' }}</template>
                </ElTableColumn>
                <ElTableColumn label="单核" width="80">
                    <template #default="{ row }">{{ row.scoreSingle || '—' }}</template>
                </ElTableColumn>
                <ElTableColumn label="多核" width="80">
                    <template #default="{ row }">{{ row.scoreMulti || '—' }}</template>
                </ElTableColumn>
                <ElTableColumn label="电池" width="110">
                    <template #default="{ row }">{{ row.battery ? row.battery + (row.batteryUnit || '') : '—' }}</template>
                </ElTableColumn>
                <ElTableColumn label="操作" width="248" fixed="right" align="center">
                    <template #default="{ row }">
                        <ElButton type="primary" plain size="small" @click.stop="openDetail(row)">查看</ElButton>
                        <ElButton v-if="isAdmin" type="primary" plain size="small" icon="Edit" @click.stop="editDevice(row)">编辑</ElButton>
                        <ElButton v-if="isAdmin" type="danger" plain size="small" icon="Delete" @click.stop="removeDevice(row)">删除</ElButton>
                    </template>
                </ElTableColumn>
            </ElTable>
        </Content>

        <ElDialog v-model="detailVisible" :title="detail?.name || '设备'" width="820px" top="6vh">
            <div v-if="detail" class="pane">
                <div class="detail-head">
                    <img v-if="detail.image" class="detail-photo" :src="detail.image" alt="" />
                    <div>
                        <ElTag>{{ familyLabel(detail.family) }}</ElTag>
                        <ElTag v-if="detail.series" type="info">{{ detail.series }}</ElTag>
                        <ElTag v-if="detail.isNew" type="success">新品</ElTag>
                        <p class="detail-line">{{ detail.chip || '—' }} · {{ detail.release || '—' }}</p>
                    </div>
                </div>
                <ElDescriptions :column="3" border size="small" title="机身">
                    <ElDescriptionsItem label="高">{{ dash(detail.size?.height, 'mm') }}</ElDescriptionsItem>
                    <ElDescriptionsItem label="宽">{{ dash(detail.size?.width, 'mm') }}</ElDescriptionsItem>
                    <ElDescriptionsItem label="厚">{{ dash(detail.size?.thickness, 'mm') }}</ElDescriptionsItem>
                    <ElDescriptionsItem label="重量">{{ dash(detail.size?.weight, 'g') }}</ElDescriptionsItem>
                    <ElDescriptionsItem label="展开高">{{ dash(detail.sizeOpen?.height, 'mm') }}</ElDescriptionsItem>
                    <ElDescriptionsItem label="展开宽">{{ dash(detail.sizeOpen?.width, 'mm') }}</ElDescriptionsItem>
                    <ElDescriptionsItem label="展开厚">{{ dash(detail.sizeOpen?.thickness, 'mm') }}</ElDescriptionsItem>
                    <ElDescriptionsItem label="单耳">{{ dash(detail.earWeight, 'g') }}</ElDescriptionsItem>
                    <ElDescriptionsItem label="充电盒">{{ dash(detail.caseWeight, 'g') }}</ElDescriptionsItem>
                    <ElDescriptionsItem label="材质">{{ detail.material || '—' }}</ElDescriptionsItem>
                    <ElDescriptionsItem label="防水">{{ detail.water || '—' }}</ElDescriptionsItem>
                    <ElDescriptionsItem label="接口">{{ detail.port || '—' }}</ElDescriptionsItem>
                </ElDescriptions>
                <ElDescriptions :column="3" border size="small" title="屏幕" class="block">
                    <ElDescriptionsItem label="尺寸">{{ dash(detail.screen?.inch, '英寸') }}</ElDescriptionsItem>
                    <ElDescriptionsItem label="面板">{{ detail.screen?.type || '—' }}</ElDescriptionsItem>
                    <ElDescriptionsItem label="像素">{{ pixelText(detail) }}</ElDescriptionsItem>
                    <ElDescriptionsItem label="密度">{{ dash(detail.screen?.density, 'ppi') }}</ElDescriptionsItem>
                    <ElDescriptionsItem label="刷新">{{ dash(detail.screen?.hz, 'Hz') }}</ElDescriptionsItem>
                    <ElDescriptionsItem label="亮度">{{ dash(detail.screen?.nits, 'nit') }}</ElDescriptionsItem>
                    <ElDescriptionsItem label="HDR">{{ detail.screen?.hdr ? '有' : '无' }}</ElDescriptionsItem>
                </ElDescriptions>
                <ElDescriptions :column="3" border size="small" title="性能" class="block">
                    <ElDescriptionsItem label="主频">{{ dash(detail.cpuRate, 'GHz') }}</ElDescriptionsItem>
                    <ElDescriptionsItem label="CPU">{{ dash(detail.cpuCore, '核') }}</ElDescriptionsItem>
                    <ElDescriptionsItem label="GPU">{{ dash(detail.gpuCore, '核') }}</ElDescriptionsItem>
                    <ElDescriptionsItem label="内存">{{ memoryText(detail) }}</ElDescriptionsItem>
                    <ElDescriptionsItem label="单核">{{ dash(detail.scoreSingle) }}</ElDescriptionsItem>
                    <ElDescriptionsItem label="多核">{{ dash(detail.scoreMulti) }}</ElDescriptionsItem>
                    <ElDescriptionsItem label="存储">{{ (detail.storage || []).length ? detail.storage.join(' / ') + ' GB' : '—' }}</ElDescriptionsItem>
                    <ElDescriptionsItem label="电池">{{ detail.battery ? detail.battery + (detail.batteryUnit || '') : '—' }}</ElDescriptionsItem>
                    <ElDescriptionsItem label="续航">{{ dash(detail.duration, '小时') }}</ElDescriptionsItem>
                    <ElDescriptionsItem label="后摄">{{ detail.camera || '—' }}</ElDescriptionsItem>
                    <ElDescriptionsItem label="操控">{{ detail.control || '—' }}</ElDescriptionsItem>
                    <ElDescriptionsItem label="降噪">{{ detail.noise ? '有' : '无' }}</ElDescriptionsItem>
                </ElDescriptions>
                <div class="block">
                    <p class="block-title">颜色</p>
                    <div v-if="detail.colors?.length" class="color-list">
                        <span v-for="color in detail.colors" :key="color.name" class="color-chip">
                            <i :style="{ background: color.color }" />{{ color.name }}
                        </span>
                    </div>
                    <span v-else>—</span>
                </div>
            </div>
            <template #footer>
                <ElButton @click="detailVisible = false">关闭</ElButton>
                <ElButton v-if="isAdmin" type="primary" icon="Edit" @click="editFromDetail">编辑</ElButton>
            </template>
        </ElDialog>

        <ElDialog
            v-model="visible"
            :title="form.id ? '编辑 ' + form.name : '添加设备'"
            width="880px"
            top="5vh"
            :close-on-click-modal="false"
        >
            <ElTabs v-model="editTab">
                <ElTabPane label="基本" name="base">
                    <ElForm :model="form" label-width="88px" :disabled="!isAdmin">
                        <ElRow :gutter="16">
                            <ElCol :span="12"><ElFormItem label="名称"><ElInput v-model="form.name" /></ElFormItem></ElCol>
                            <ElCol :span="12"><ElFormItem label="短名"><ElInput v-model="form.nameShort" /></ElFormItem></ElCol>
                        </ElRow>
                        <ElRow :gutter="16">
                            <ElCol :span="8">
                                <ElFormItem label="产品">
                                    <ElSelect v-model="form.family" style="width: 100%">
                                        <ElOption label="iPhone" value="iphone" />
                                        <ElOption label="iPad" value="ipad" />
                                        <ElOption label="Watch" value="watch" />
                                        <ElOption label="AirPods" value="airpods" />
                                    </ElSelect>
                                </ElFormItem>
                            </ElCol>
                            <ElCol :span="8"><ElFormItem label="系列"><ElInput v-model="form.series" placeholder="Pro / mini / Air" /></ElFormItem></ElCol>
                            <ElCol :span="8"><ElFormItem label="发布"><ElInput v-model="form.release" placeholder="2026-11-06" /></ElFormItem></ElCol>
                        </ElRow>
                        <ElRow :gutter="16">
                            <ElCol :span="12"><ElFormItem label="芯片"><ElInput v-model="form.chip" /></ElFormItem></ElCol>
                            <ElCol :span="12"><ElFormItem label="新品"><ElSwitch v-model="form.isNew" /></ElFormItem></ElCol>
                        </ElRow>
                        <ElFormItem label="图片"><ElInput v-model="form.image" placeholder="图片地址，可空" /></ElFormItem>
                    </ElForm>
                </ElTabPane>
                <ElTabPane label="机身" name="body">
                    <ElForm :model="form" label-width="88px" :disabled="!isAdmin">
                        <p class="hint">高是长边，单位毫米、克。没有的尺寸留 0。</p>
                        <ElRow :gutter="16">
                            <ElCol :span="6"><ElFormItem label="高"><ElInputNumber v-model="form.size.height" :precision="2" :controls="false" style="width: 100%" /></ElFormItem></ElCol>
                            <ElCol :span="6"><ElFormItem label="宽"><ElInputNumber v-model="form.size.width" :precision="2" :controls="false" style="width: 100%" /></ElFormItem></ElCol>
                            <ElCol :span="6"><ElFormItem label="厚"><ElInputNumber v-model="form.size.thickness" :precision="2" :controls="false" style="width: 100%" /></ElFormItem></ElCol>
                            <ElCol :span="6"><ElFormItem label="重量"><ElInputNumber v-model="form.size.weight" :precision="2" :controls="false" style="width: 100%" /></ElFormItem></ElCol>
                        </ElRow>
                        <ElRow :gutter="16">
                            <ElCol :span="8"><ElFormItem label="展开高"><ElInputNumber v-model="form.sizeOpen.height" :precision="2" :controls="false" style="width: 100%" /></ElFormItem></ElCol>
                            <ElCol :span="8"><ElFormItem label="展开宽"><ElInputNumber v-model="form.sizeOpen.width" :precision="2" :controls="false" style="width: 100%" /></ElFormItem></ElCol>
                            <ElCol :span="8"><ElFormItem label="展开厚"><ElInputNumber v-model="form.sizeOpen.thickness" :precision="2" :controls="false" style="width: 100%" /></ElFormItem></ElCol>
                        </ElRow>
                        <ElRow :gutter="16">
                            <ElCol :span="8"><ElFormItem label="单耳"><ElInputNumber v-model="form.earWeight" :precision="2" :controls="false" style="width: 100%" /></ElFormItem></ElCol>
                            <ElCol :span="8"><ElFormItem label="盒重"><ElInputNumber v-model="form.caseWeight" :precision="2" :controls="false" style="width: 100%" /></ElFormItem></ElCol>
                            <ElCol :span="8"><ElFormItem label="接口"><ElInput v-model="form.port" /></ElFormItem></ElCol>
                        </ElRow>
                        <ElRow :gutter="16">
                            <ElCol :span="12"><ElFormItem label="材质"><ElInput v-model="form.material" /></ElFormItem></ElCol>
                            <ElCol :span="12"><ElFormItem label="防水"><ElInput v-model="form.water" placeholder="IP68" /></ElFormItem></ElCol>
                        </ElRow>
                    </ElForm>
                </ElTabPane>
                <ElTabPane label="屏幕" name="screen">
                    <ElForm :model="form" label-width="88px" :disabled="!isAdmin">
                        <ElRow :gutter="16">
                            <ElCol :span="8"><ElFormItem label="英寸"><ElInputNumber v-model="form.screen.inch" :precision="2" :controls="false" style="width: 100%" /></ElFormItem></ElCol>
                            <ElCol :span="8"><ElFormItem label="面板"><ElInput v-model="form.screen.type" placeholder="OLED" /></ElFormItem></ElCol>
                            <ElCol :span="8"><ElFormItem label="HDR"><ElSwitch v-model="form.screen.hdr" /></ElFormItem></ElCol>
                        </ElRow>
                        <ElRow :gutter="16">
                            <ElCol :span="8"><ElFormItem label="像素宽"><ElInputNumber v-model="form.screen.pxW" :controls="false" style="width: 100%" /></ElFormItem></ElCol>
                            <ElCol :span="8"><ElFormItem label="像素高"><ElInputNumber v-model="form.screen.pxH" :controls="false" style="width: 100%" /></ElFormItem></ElCol>
                            <ElCol :span="8"><ElFormItem label="密度"><ElInputNumber v-model="form.screen.density" :controls="false" style="width: 100%" /></ElFormItem></ElCol>
                        </ElRow>
                        <ElRow :gutter="16">
                            <ElCol :span="8"><ElFormItem label="刷新"><ElInputNumber v-model="form.screen.hz" :controls="false" style="width: 100%" /></ElFormItem></ElCol>
                            <ElCol :span="8"><ElFormItem label="亮度"><ElInputNumber v-model="form.screen.nits" :controls="false" style="width: 100%" /></ElFormItem></ElCol>
                        </ElRow>
                    </ElForm>
                </ElTabPane>
                <ElTabPane label="性能" name="perf">
                    <ElForm :model="form" label-width="88px" :disabled="!isAdmin">
                        <p class="hint">单核、多核是 Geekbench 6。没有成绩就留 0。</p>
                        <ElRow :gutter="16">
                            <ElCol :span="8"><ElFormItem label="主频"><ElInputNumber v-model="form.cpuRate" :precision="2" :controls="false" style="width: 100%" /></ElFormItem></ElCol>
                            <ElCol :span="8"><ElFormItem label="CPU核"><ElInputNumber v-model="form.cpuCore" :controls="false" style="width: 100%" /></ElFormItem></ElCol>
                            <ElCol :span="8"><ElFormItem label="GPU核"><ElInputNumber v-model="form.gpuCore" :controls="false" style="width: 100%" /></ElFormItem></ElCol>
                        </ElRow>
                        <ElRow :gutter="16">
                            <ElCol :span="8"><ElFormItem label="内存"><ElInputNumber v-model="form.memoryGb" :precision="1" :controls="false" style="width: 100%" /></ElFormItem></ElCol>
                            <ElCol :span="8"><ElFormItem label="类型"><ElInput v-model="form.memoryType" placeholder="LPDDR5X" /></ElFormItem></ElCol>
                            <ElCol :span="8"><ElFormItem label="单核"><ElInputNumber v-model="form.scoreSingle" :controls="false" style="width: 100%" /></ElFormItem></ElCol>
                        </ElRow>
                        <ElRow :gutter="16">
                            <ElCol :span="8"><ElFormItem label="多核"><ElInputNumber v-model="form.scoreMulti" :controls="false" style="width: 100%" /></ElFormItem></ElCol>
                            <ElCol :span="8"><ElFormItem label="电池"><ElInputNumber v-model="form.battery" :controls="false" style="width: 100%" /></ElFormItem></ElCol>
                            <ElCol :span="8"><ElFormItem label="单位"><ElInput v-model="form.batteryUnit" placeholder="mAh / mWh" /></ElFormItem></ElCol>
                        </ElRow>
                        <ElRow :gutter="16">
                            <ElCol :span="8"><ElFormItem label="续航"><ElInputNumber v-model="form.duration" :controls="false" style="width: 100%" /></ElFormItem></ElCol>
                            <ElCol :span="8"><ElFormItem label="降噪"><ElSwitch v-model="form.noise" /></ElFormItem></ElCol>
                        </ElRow>
                        <ElFormItem label="后摄"><ElInput v-model="form.camera" /></ElFormItem>
                        <ElFormItem label="操控"><ElInput v-model="form.control" /></ElFormItem>
                        <ElFormItem label="存储">
                            <div class="storage-box">
                                <ElTag
                                    v-for="(item, index) in form.storage"
                                    :key="item + index"
                                    closable
                                    :disable-transitions="true"
                                    @close="form.storage.splice(index, 1)"
                                >{{ item }} GB</ElTag>
                                <ElInput
                                    v-model="storageDraft"
                                    placeholder="容量，回车加入"
                                    style="width: 140px"
                                    @keyup.enter="addStorage"
                                />
                            </div>
                        </ElFormItem>
                    </ElForm>
                </ElTabPane>
                <ElTabPane label="颜色" name="color">
                    <div v-for="(color, index) in form.colors" :key="index" class="color-edit">
                        <ElColorPicker v-model="color.color" />
                        <ElInput v-model="color.name" placeholder="颜色名" />
                        <ElInput v-model="color.color" placeholder="#hex" />
                        <ElButton type="danger" plain @click="form.colors.splice(index, 1)">删除</ElButton>
                    </div>
                    <ElButton icon="Plus" @click="form.colors.push({ name: '', color: '#cccccc' })">加一种颜色</ElButton>
                </ElTabPane>
                <ElTabPane label="原始 JSON" name="json">
                    <p class="hint">相机、充电、型号这些嵌套字段。留空表示不改原来的原始参数。</p>
                    <ElInput v-model="sourceText" type="textarea" :rows="16" placeholder="JSON" />
                </ElTabPane>
            </ElTabs>
            <template #footer>
                <ElButton @click="visible = false">取消</ElButton>
                <ElButton v-if="isAdmin" type="primary" :loading="isSaving" @click="save">保存</ElButton>
            </template>
        </ElDialog>
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

const FAMILY: Record<string, string> = { iphone: 'iPhone', ipad: 'iPad', watch: 'Watch', airpods: 'AirPods' }

const projectStore = useProjectStore()
const isAdmin = computed(() => Number(getAuthorization()?.group_id) === 1)
const isLoading = ref(false)
const isSaving = ref(false)
const family = ref('')
const keyword = ref('')
const rows = ref<any[]>([])
const visible = ref(false)
const detailVisible = ref(false)
const detail = ref<any>(null)
const editTab = ref('base')
const form = ref<any>(blank())
const sourceText = ref('')
const storageDraft = ref('')

const tableHeight = computed(() => {
    const insets = projectStore.contentInsets
    return insets.heightContent - insets.heightToolbar - 8
})
const shown = computed(() => {
    const key = keyword.value.trim().toLowerCase()
    return rows.value.filter(item => {
        if (family.value && item.family !== family.value) return false
        if (!key) return true
        return [item.name, item.nameShort, item.chip, item.series].some(part => String(part || '').toLowerCase().includes(key))
    })
})

// 新建设备时的空表单。
function blank() {
    return {
        id: '', family: 'iphone', series: '', name: '', nameShort: '', release: '', year: 0,
        image: '', isNew: false, chip: '',
        size: { height: 0, width: 0, thickness: 0, weight: 0 },
        sizeOpen: { height: 0, width: 0, thickness: 0 },
        earWeight: 0, caseWeight: 0, material: '', water: '',
        screen: { inch: 0, type: '', pxW: 0, pxH: 0, density: 0, hz: 0, nits: 0, hdr: false },
        cpuRate: 0, cpuCore: 0, gpuCore: 0, scoreSingle: 0, scoreMulti: 0,
        memoryGb: 0, memoryType: '', storage: [] as string[], battery: 0, batteryUnit: 'mAh',
        camera: '', port: '', noise: false, duration: 0, control: '', colors: [] as { name: string, color: string }[]
    }
}

// 产品名。
function familyLabel(value: string) {
    return FAMILY[value] || value || '—'
}

// 空值和 0 显示成横线。
function dash(value: unknown, unit = '') {
    if (value === 0 || value === '' || value === null || value === undefined) return '—'
    return unit ? `${value} ${unit}` : String(value)
}

// 列表里的高宽厚。
function sizeText(row: any) {
    const size = row.size || {}
    if (!size.height) return '—'
    return `${size.height}×${size.width}×${size.thickness}`
}

// 屏幕像素，高是长边。
function pixelText(row: any) {
    const screen = row.screen || {}
    if (!screen.pxW || !screen.pxH) return '—'
    return `${screen.pxW}×${screen.pxH}`
}

// 内存容量和类型。
function memoryText(row: any) {
    if (!row.memoryGb) return '—'
    return row.memoryGb + ' GB' + (row.memoryType ? ' ' + row.memoryType : '')
}

// 从数据库拉设备，含原始嵌套参数。
async function loadData() {
    isLoading.value = true
    try {
        const res: any = await appleCatalogApi.list()
        rows.value = res.data?.devices || []
    } finally {
        isLoading.value = false
    }
}

// 只读弹窗。
function openDetail(row: any) {
    detail.value = row
    detailVisible.value = true
}

// 从查看切到编辑。
function editFromDetail() {
    const row = detail.value
    detailVisible.value = false
    if (row) editDevice(row)
}

// 打开这一行，原始参数单独放在文本框里。
function editDevice(row: any) {
    const copy = JSON.parse(JSON.stringify(row))
    sourceText.value = copy.source ? JSON.stringify(copy.source, null, 2) : ''
    delete copy.source
    form.value = {
        ...blank(),
        ...copy,
        size: { ...blank().size, ...(copy.size || {}) },
        sizeOpen: { ...blank().sizeOpen, ...(copy.sizeOpen || {}) },
        screen: { ...blank().screen, ...(copy.screen || {}) },
        colors: copy.colors || [],
        storage: copy.storage || []
    }
    storageDraft.value = ''
    editTab.value = 'base'
    visible.value = true
}

// 清空表单，准备新增。
function addDevice() {
    form.value = blank()
    sourceText.value = ''
    storageDraft.value = ''
    editTab.value = 'base'
    visible.value = true
}

// 回车把容量加进列表。
function addStorage() {
    const text = storageDraft.value.trim()
    if (!text) return
    form.value.storage.push(text.replace(/gb$/i, '').trim())
    storageDraft.value = ''
}

// 校验 JSON 后整行写回。
function save() {
    if (!form.value.name) {
        ElMessage.warning('先写名称')
        editTab.value = 'base'
        return
    }
    let source: unknown = undefined
    if (sourceText.value.trim()) {
        try {
            source = JSON.parse(sourceText.value)
        } catch {
            ElMessage.error('原始参数不是合法 JSON')
            editTab.value = 'json'
            return
        }
    }
    const payload = JSON.parse(JSON.stringify(form.value))
    const open = payload.sizeOpen || {}
    if (!open.height && !open.width && !open.thickness) delete payload.sizeOpen
    const numbers = (payload.storage || []).map((item: string) => Number(String(item).replace(/[^\d.]/g, ''))).filter((item: number) => item > 0)
    payload.storageMax = numbers.length ? Math.max(...numbers) : 0
    payload.year = Number(String(payload.release || '').slice(0, 4)) || payload.year || 0
    if (source !== undefined) payload.source = source
    isSaving.value = true
    appleCatalogApi.saveDevice(payload).then(() => {
        ElMessage.success('已保存')
        visible.value = false
        loadData()
    }).finally(() => {
        isSaving.value = false
    })
}

// 删掉这一台。
function removeDevice(row: any) {
    ElMessageBox.confirm(`删除 ${row.name}？`, '删除设备').then(() => {
        appleCatalogApi.deleteDevice(row.id).then(() => loadData())
    }).catch(() => {})
}

onMounted(loadData)
</script>

<style scoped lang="scss">
.pane {
    max-height: calc(100vh - 220px);
    overflow: auto;
    padding-right: 4px;
}

.detail-head {
    display: flex;
    gap: 16px;
    align-items: center;
    margin-bottom: 16px;
}

.detail-photo {
    width: 72px;
    height: 72px;
    object-fit: contain;
}

.detail-line {
    margin: 8px 0 0;
}

.block {
    margin-top: 16px;
}

.block-title {
    margin: 0 0 8px;
    font-weight: 600;
}

.color-list {
    display: flex;
    flex-wrap: wrap;
    gap: 8px;
}

.color-chip {
    display: inline-flex;
    align-items: center;
    gap: 6px;

    i {
        width: 16px;
        height: 16px;
        border-radius: 4px;
        border: 1px solid rgba(0, 0, 0, 0.15);
    }
}

.hint {
    margin: 0 0 12px;
    color: var(--el-text-color-secondary);
    font-size: 13px;
}

.storage-box {
    display: flex;
    flex-wrap: wrap;
    gap: 8px;
    align-items: center;
}

.color-edit {
    display: grid;
    grid-template-columns: 40px 1fr 160px auto;
    gap: 8px;
    align-items: center;
    margin-bottom: 8px;
}
</style>

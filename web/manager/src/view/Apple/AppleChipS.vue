<template>
    <Container>
        <Toolbar>
            <template #left>
                <ElButton type="primary" @click="loadData" :loading="isLoading" icon="Refresh">刷新数据</ElButton>
                <ElButton type="success" @click="saveAllData" :loading="isSaving" :disabled="!diaryId || chipList.length === 0" icon="Check">保存全部</ElButton>
                <ElButton type="primary" @click="addChip" icon="Plus">添加芯片</ElButton>
            </template>
        </Toolbar>
        <Content padding="0">
            <ElTable 
                v-if="diaryData" 
                :data="chipList" 
                stripe 
                style="width: 100%" 
                @row-click="editChip"
                row-key="name"
                ref="tableRef"
            >
                <ElTableColumn label="排序" width="60" fixed="left">
                    <template #default>
                        <span style="cursor: move; user-select: none; color: #909399;">
                            <svg viewBox="0 0 1024 1024" xmlns="http://www.w3.org/2000/svg" style="width: 16px; height: 16px; vertical-align: middle;">
                                <path d="M384 128h256a32 32 0 0 1 0 64H384a32 32 0 1 1 0-64zm0 256h256a32 32 0 0 1 0 64H384a32 32 0 0 1 0-64zm0 256h256a32 32 0 0 1 0 64H384a32 32 0 0 1 0-64z" fill="currentColor"></path>
                                <path d="M128 224h128a32 32 0 0 1 0 64H128a32 32 0 0 1 0-64zm0 256h128a32 32 0 0 1 0 64H128a32 32 0 0 1 0-64zm0 256h128a32 32 0 0 1 0 64H128a32 32 0 0 1 0-64z" fill="currentColor"></path>
                            </svg>
                        </span>
                    </template>
                </ElTableColumn>
                <ElTableColumn prop="name" label="名称" width="150" />
                <ElTableColumn prop="model" label="型号" width="120" />
                <ElTableColumn prop="tech" label="工艺" width="100" />
                <ElTableColumn prop="techCompany" label="工艺厂商" width="120" />
                <ElTableColumn prop="release" label="发布时间" width="100" />
                <ElTableColumn prop="isa" label="ISA" width="120" />
                <ElTableColumn label="CPU 核心" width="120">
                    <template #default="{ row }">
                        <span v-if="row.cpu && row.cpu.length > 0">
                            {{ row.cpu.reduce((sum: number, cpu: CpuConfig) => sum + (cpu.fire?.core || 0) + (cpu.ice?.core || 0), 0) }}
                            <span style="color: #909399; font-size: 12px;">
                                (F:{{ row.cpu.reduce((sum: number, cpu: CpuConfig) => sum + (cpu.fire?.core || 0), 0) }}/
                                I:{{ row.cpu.reduce((sum: number, cpu: CpuConfig) => sum + (cpu.ice?.core || 0), 0) }})
                            </span>
                        </span>
                        <span v-else>-</span>
                    </template>
                </ElTableColumn>
                <ElTableColumn label="CPU 频率" width="120">
                    <template #default="{ row }">
                        <span v-if="row.cpu && row.cpu.length > 0">
                            {{ Math.max(...row.cpu.map((cpu: CpuConfig) => Math.max(cpu.fire?.rate || 0, cpu.ice?.rate || 0))).toFixed(2) }} GHz
                        </span>
                        <span v-else>-</span>
                    </template>
                </ElTableColumn>
                <ElTableColumn label="GPU" width="150">
                    <template #default="{ row }">
                        <span v-if="row.gpu && row.gpu.length > 0">
                            {{ row.gpu.map((gpu: GpuConfig) => gpu.brand || '未知').join(', ') }}
                        </span>
                        <span v-else>-</span>
                    </template>
                </ElTableColumn>
                <ElTableColumn label="GPU 核心" width="120">
                    <template #default="{ row }">
                        <span v-if="row.gpu && row.gpu.length > 0">
                            {{ row.gpu.reduce((sum: number, gpu: GpuConfig) => sum + (gpu.core || 0), 0) }}
                            <span v-if="row.gpu.length > 1" style="color: #909399; font-size: 12px;">
                                ({{ row.gpu.length }}个)
                            </span>
                        </span>
                        <span v-else>-</span>
                    </template>
                </ElTableColumn>
                <ElTableColumn label="AI 核心" width="100">
                    <template #default="{ row }">
                        {{ row.ai?.core || '-' }}
                    </template>
                </ElTableColumn>
                <ElTableColumn label="操作" width="250" fixed="right">
                    <template #default="{ $index }">
                        <ElButton type="primary" size="small" @click.stop="editChipByIndex($index)" icon="Edit">编辑</ElButton>
                        <ElButton type="warning" size="small" @click.stop="cloneChip($index)" icon="CopyDocument">克隆</ElButton>
                        <ElButton type="danger" size="small" @click.stop="removeChip($index)" icon="Delete">删除</ElButton>
                    </template>
                </ElTableColumn>
            </ElTable>
            <ElEmpty v-else description="暂无数据，请点击刷新按钮加载数据" />
        </Content>

        <!-- 编辑 Dialog -->
        <ElDialog
            v-model="dialogVisible"
            :title="editingIndex !== null ? '编辑芯片' : '添加芯片'"
            width="80%"
            :close-on-click-modal="false"
            :before-close="handleDialogClose"
        >
            <ElForm :model="editingChip" label-width="100px" v-if="editingChip">
                <ElRow :gutter="15">
                    <ElCol :span="8">
                        <ElFormItem label="名称">
                            <ElAutocomplete
                                v-model="editingChip.name"
                                :fetch-suggestions="createFetchSuggestions(nameSuggestions)"
                                placeholder="芯片名称"
                                clearable
                            />
                        </ElFormItem>
                    </ElCol>
                    <ElCol :span="8">
                        <ElFormItem label="型号">
                            <ElAutocomplete
                                v-model="editingChip.model"
                                :fetch-suggestions="createFetchSuggestions(modelSuggestions)"
                                placeholder="型号"
                                clearable
                            />
                        </ElFormItem>
                    </ElCol>
                    <ElCol :span="8">
                        <ElFormItem label="工艺">
                            <ElAutocomplete
                                v-model="editingChip.tech"
                                :fetch-suggestions="createFetchSuggestions(techSuggestions)"
                                placeholder="工艺"
                                clearable
                            />
                        </ElFormItem>
                    </ElCol>
                </ElRow>
                <ElRow :gutter="15">
                    <ElCol :span="8">
                        <ElFormItem label="工艺厂商">
                            <ElAutocomplete
                                v-model="editingChip.techCompany"
                                :fetch-suggestions="createFetchSuggestions(techCompanySuggestions)"
                                placeholder="工艺厂商"
                                clearable
                            />
                        </ElFormItem>
                    </ElCol>
                    <ElCol :span="8">
                        <ElFormItem label="芯片尺寸">
                            <ElAutocomplete
                                v-model="editingChip.dieSize"
                                :fetch-suggestions="createFetchSuggestions(dieSizeSuggestions)"
                                placeholder="芯片尺寸"
                                clearable
                            />
                        </ElFormItem>
                    </ElCol>
                    <ElCol :span="8">
                        <ElFormItem label="ISA">
                            <ElAutocomplete
                                v-model="editingChip.isa"
                                :fetch-suggestions="createFetchSuggestions(isaSuggestions)"
                                placeholder="ISA"
                                clearable
                            />
                        </ElFormItem>
                    </ElCol>
                </ElRow>
                <ElRow :gutter="15">
                    <ElCol :span="8">
                        <ElFormItem label="发布时间">
                            <ElAutocomplete
                                v-model="editingChip.release"
                                :fetch-suggestions="createFetchSuggestions(releaseSuggestions)"
                                placeholder="发布时间"
                                clearable
                            />
                        </ElFormItem>
                    </ElCol>
                    <ElCol :span="8">
                        <ElFormItem label="晶体管数量">
                            <ElAutocomplete
                                v-model="editingChip.transistorCount"
                                :fetch-suggestions="createFetchSuggestions(transistorCountSuggestions.filter(v => v !== undefined) as string[])"
                                placeholder="晶体管数量（可选）"
                                clearable
                            />
                        </ElFormItem>
                    </ElCol>
                </ElRow>

                <!-- CPU -->
                <ElDivider content-position="left">CPU 配置</ElDivider>
                <div v-for="(cpu, cpuIndex) in editingChip.cpu" :key="cpuIndex" style="margin-bottom: 15px; padding: 15px; background: #f5f7fa; border-radius: 4px;">
                    <div style="display: flex; justify-content: space-between; margin-bottom: 10px;">
                        <strong>CPU 配置 #{{ cpuIndex + 1 }}</strong>
                        <ElButton type="danger" size="small" @click="removeCpu(cpuIndex)" icon="Delete">删除</ElButton>
                    </div>
                    <ElRow :gutter="15">
                        <ElCol :span="6">
                            <ElFormItem label="Fire 频率">
                                <ElInputNumber v-model="cpu.fire.rate" :precision="2" style="width: 100%;" />
                            </ElFormItem>
                        </ElCol>
                        <ElCol :span="6">
                            <ElFormItem label="Fire 核心数">
                                <ElInputNumber v-model="cpu.fire.core" :min="0" style="width: 100%;" />
                            </ElFormItem>
                        </ElCol>
                        <ElCol :span="6">
                            <ElFormItem label="Ice 频率">
                                <ElInputNumber v-model="cpu.ice.rate" :precision="2" style="width: 100%;" />
                            </ElFormItem>
                        </ElCol>
                        <ElCol :span="6">
                            <ElFormItem label="Ice 核心数">
                                <ElInputNumber v-model="cpu.ice.core" :min="0" style="width: 100%;" />
                            </ElFormItem>
                        </ElCol>
                    </ElRow>
                </div>
                <ElButton size="small" @click="addCpu" icon="Plus">添加 CPU 配置</ElButton>

                <!-- GPU -->
                <ElDivider content-position="left">GPU 配置</ElDivider>
                <div v-for="(gpu, gpuIndex) in editingChip.gpu" :key="gpuIndex" style="margin-bottom: 15px; padding: 15px; background: #f5f7fa; border-radius: 4px;">
                    <div style="display: flex; justify-content: space-between; margin-bottom: 10px;">
                        <strong>GPU 配置 #{{ gpuIndex + 1 }}</strong>
                        <ElButton type="danger" size="small" @click="removeGpu(gpuIndex)" icon="Delete">删除</ElButton>
                    </div>
                    <ElRow :gutter="15">
                        <ElCol :span="8">
                            <ElFormItem label="品牌">
                                <ElAutocomplete
                                    v-model="gpu.brand"
                                    :fetch-suggestions="createFetchSuggestions(gpuBrandSuggestions)"
                                    placeholder="品牌"
                                    clearable
                                />
                            </ElFormItem>
                        </ElCol>
                        <ElCol :span="8">
                            <ElFormItem label="核心数">
                                <ElInputNumber v-model="gpu.core" :min="0" style="width: 100%;" />
                            </ElFormItem>
                        </ElCol>
                    </ElRow>
                    <ElFormItem label="信息">
                        <ElInput v-model="gpu.info" placeholder="GPU 信息" />
                    </ElFormItem>
                </div>
                <ElButton size="small" @click="addGpu" icon="Plus">添加 GPU 配置</ElButton>

                <!-- AI -->
                <ElDivider content-position="left">AI 配置</ElDivider>
                <ElRow :gutter="15">
                    <ElCol :span="8">
                        <ElFormItem label="AI 核心数">
                            <ElInput v-model="editingChip.ai.core" placeholder="AI 核心数" />
                        </ElFormItem>
                    </ElCol>
                    <ElCol :span="8">
                        <ElFormItem label="AI 速率">
                            <ElInput v-model="editingChip.ai.rate" placeholder="AI 速率" />
                        </ElFormItem>
                    </ElCol>
                </ElRow>

                <!-- Devices -->
                <ElDivider content-position="left">设备列表</ElDivider>
                <div v-for="(_, deviceIndex) in editingChip.devices" :key="deviceIndex" style="margin-bottom: 10px; display: flex; align-items: center;">
                    <ElAutocomplete
                        v-model="editingChip.devices[deviceIndex]"
                        :fetch-suggestions="createFetchSuggestions(deviceSuggestions)"
                        placeholder="设备名称"
                        clearable
                        style="flex: 1; margin-right: 10px;"
                    />
                    <ElButton type="danger" size="small" @click="removeDevice(deviceIndex)" icon="Delete">删除</ElButton>
                </div>
                <ElButton size="small" @click="addDevice" icon="Plus">添加设备</ElButton>

                <!-- OS -->
                <ElDivider content-position="left">操作系统</ElDivider>
                <ElRow :gutter="15">
                    <ElCol :span="8">
                        <ElFormItem label="初始系统">
                            <ElAutocomplete
                                v-model="editingChip.os.init"
                                :fetch-suggestions="createFetchSuggestions(osInitSuggestions)"
                                placeholder="初始系统"
                                clearable
                            />
                        </ElFormItem>
                    </ElCol>
                    <ElCol :span="8">
                        <ElFormItem label="最新系统">
                            <ElAutocomplete
                                v-model="editingChip.os.latest"
                                :fetch-suggestions="createFetchSuggestions(osLatestSuggestions)"
                                placeholder="最新系统"
                                clearable
                            />
                        </ElFormItem>
                    </ElCol>
                </ElRow>
            </ElForm>

            <template #footer>
                <div style="display: flex; justify-content: space-between;">
                    <ElButton @click="dialogVisible = false">取消</ElButton>
                    <div>
                        <ElButton type="primary" @click="handleSaveButton" :loading="isSavingChip" icon="Check">保存此芯片</ElButton>
                    </div>
                </div>
            </template>
        </ElDialog>
    </Container>
</template>

<script setup lang="ts">
import { ref, onMounted, computed, nextTick, onUnmounted } from 'vue'
import { ElMessage, ElMessageBox } from 'element-plus'
import Sortable from 'sortablejs'
import diaryApi from "@/api/diaryApi"
import Container from "@/layout/Container.vue"
import Toolbar from "@/layout/Toolbar.vue"
import Content from "@/layout/Content.vue"
import { dateFormatter } from "@/utility"
import { Chip, CpuConfig, GpuConfig } from "@/model/appleChip"

const isLoading = ref(false)
const isSaving = ref(false)
const isSavingChip = ref(false)
const diaryId = ref<number | null>(null)
const diaryData = ref<any>(null)
const chipList = ref<Chip[]>([])
const formData = ref({
    title: ''
})
const dialogVisible = ref(false)
const editingIndex = ref<number | null>(null)
const editingChip = ref<Chip | null>(null)
const originalChip = ref<Chip | null>(null) // 保存原始数据用于比较是否有修改
const tableRef = ref<any>(null)
let sortableInstance: Sortable | null = null

const keyword = 'apple-chip-s'

// 存储所有三个系列的数据用于生成备选项
const allChipsData = ref<Chip[]>([])
const allChipsDataLoaded = ref(false) // 标记是否已加载过所有系列数据

// 收集所有芯片中相同字段的值作为备选项（从 A、S、M 三个系列中收集）
const nameSuggestions = computed(() => {
    const values = allChipsData.value.map(chip => chip.name).filter(v => v && v.trim())
    return [...new Set(values)]
})

const modelSuggestions = computed(() => {
    const values = allChipsData.value.map(chip => chip.model).filter(v => v && v.trim())
    return [...new Set(values)]
})

const techSuggestions = computed(() => {
    const values = allChipsData.value.map(chip => chip.tech).filter(v => v && v.trim())
    return [...new Set(values)]
})

const techCompanySuggestions = computed(() => {
    const values = allChipsData.value.map(chip => chip.techCompany).filter(v => v && v.trim())
    return [...new Set(values)]
})

const dieSizeSuggestions = computed(() => {
    const values = allChipsData.value.map(chip => chip.dieSize).filter(v => v && v.trim())
    return [...new Set(values)]
})

const isaSuggestions = computed(() => {
    const values = allChipsData.value.map(chip => chip.isa).filter(v => v && v.trim())
    return [...new Set(values)]
})

const releaseSuggestions = computed(() => {
    const values = allChipsData.value.map(chip => chip.release).filter(v => v && v.trim())
    return [...new Set(values)]
})

const transistorCountSuggestions = computed(() => {
    const values = allChipsData.value.map(chip => chip.transistorCount).filter(v => v && v.trim())
    return [...new Set(values)]
})

const gpuBrandSuggestions = computed(() => {
    const values: string[] = []
    allChipsData.value.forEach(chip => {
        chip.gpu.forEach(gpu => {
            if (gpu.brand && gpu.brand.trim()) {
                values.push(gpu.brand)
            }
        })
    })
    return [...new Set(values)]
})

const osInitSuggestions = computed(() => {
    const values = allChipsData.value.map(chip => chip.os.init).filter(v => v && v.trim())
    return [...new Set(values)]
})

const osLatestSuggestions = computed(() => {
    const values = allChipsData.value.map(chip => chip.os.latest).filter(v => v && v.trim())
    return [...new Set(values)]
})

const deviceSuggestions = computed(() => {
    const values: string[] = []
    allChipsData.value.forEach(chip => {
        chip.devices.forEach(device => {
            if (device && device.trim()) {
                values.push(device)
            }
        })
    })
    return [...new Set(values)]
})

// 过滤函数
const filterSuggestions = (queryString: string, suggestions: string[]) => {
    if (!queryString) return suggestions
    return suggestions.filter(item => 
        item.toLowerCase().includes(queryString.toLowerCase())
    )
}

// 创建 autocomplete 的 fetch-suggestions 函数
const createFetchSuggestions = (suggestions: string[]) => {
    return (queryString: string, cb: (suggestions: Array<{ value: string }>) => void) => {
        const filtered = filterSuggestions(queryString, suggestions)
        cb(filtered.map(v => ({ value: v })))
    }
}

// 加载所有三个系列的数据用于生成备选项
const loadAllChipsData = async () => {
    const keywords = ['apple-chip-a', 'apple-chip-s', 'apple-chip-m']
    const allChips: Chip[] = []
    
    try {
        // 并行加载三个系列的数据
        const promises = keywords.map(keyword => 
            diaryApi.getLatestPublicDiaryWithKeyword({ keyword })
        )
        const results = await Promise.all(promises)
        
        results.forEach((res: any) => {
            if (res && res.data && res.data.content) {
                try {
                    const content = res.data.content || '[]'
                    const chips = JSON.parse(content)
                    if (Array.isArray(chips)) {
                        allChips.push(...chips)
                    }
                } catch (error) {
                    console.error('解析 JSON 失败:', error)
                }
            }
        })
        
        allChipsData.value = allChips
    } catch (error) {
        console.error('加载所有系列数据失败:', error)
    }
}

const loadData = async () => {
    isLoading.value = true
    try {
        // 先加载当前系列的数据
        const res: any = await diaryApi.getLatestPublicDiaryWithKeyword({ keyword })
        if (res && res.data) {
            const latestDiary = res.data
            diaryId.value = latestDiary.id
            diaryData.value = latestDiary
            formData.value.title = latestDiary.title || ''
            
            // 解析 JSON 内容（content 是数组的 JSON 字符串）
            try {
                const content = latestDiary.content || '[]'
                chipList.value = JSON.parse(content)
                if (!Array.isArray(chipList.value)) {
                    chipList.value = []
                }
            } catch (error) {
                console.error('解析 JSON 失败:', error)
                chipList.value = []
                ElMessage.warning('数据格式错误，已重置为空数组')
            }
            
            // 只在首次加载或备选项数据为空时，加载所有系列的数据用于生成备选项
            if (!allChipsDataLoaded.value || allChipsData.value.length === 0) {
                await loadAllChipsData()
                allChipsDataLoaded.value = true
            }
            
            // 重新初始化拖拽排序
            initSortable()
            
        } else {
            ElMessage.warning('未找到相关数据')
            diaryId.value = null
            diaryData.value = null
            chipList.value = []
            formData.value.title = ''
        }
    } catch (error: any) {
        console.error('加载数据失败:', error)
        ElMessage.error('加载数据失败: ' + (error.message || '未知错误'))
    } finally {
        isLoading.value = false
    }
}

const addChip = () => {
    const newChip: Chip = {
        name: '',
        model: '',
        tech: '',
        techCompany: '',
        dieSize: '',
        isa: '',
        cpu: [{ fire: { rate: 0, core: 0 }, ice: { rate: 0, core: 0 } }],
        gpu: [{ brand: '', core: 0, info: '' }],
        ai: { core: '', rate: '' },
        release: '',
        devices: [],
        os: { init: '', latest: '' }
    }
    chipList.value.unshift(newChip)
    editingIndex.value = 0
    // 深拷贝新芯片数据
    const chipData = JSON.parse(JSON.stringify(newChip))
    editingChip.value = chipData
    originalChip.value = JSON.parse(JSON.stringify(chipData)) // 保存原始数据副本
    dialogVisible.value = true
}

const editChip = (row: Chip) => {
    const index = chipList.value.findIndex(chip => chip === row)
    if (index !== -1) {
        editChipByIndex(index)
    }
}

const editChipByIndex = (index: number) => {
    editingIndex.value = index
    // 深拷贝芯片数据
    const chipData = JSON.parse(JSON.stringify(chipList.value[index]))
    editingChip.value = chipData
    originalChip.value = JSON.parse(JSON.stringify(chipData)) // 保存原始数据副本
    dialogVisible.value = true
}

const cloneChip = (index: number) => {
    // 深拷贝芯片数据
    const clonedChip = JSON.parse(JSON.stringify(chipList.value[index]))
    // 添加到列表前面
    chipList.value.unshift(clonedChip)
    // 打开编辑对话框编辑克隆的芯片
    editChipByIndex(0)
    ElMessage.success('芯片已克隆')
}

const removeChip = (index: number) => {
    ElMessageBox.confirm('确定要删除这个芯片吗？', '确认删除', {
        confirmButtonText: '确定',
        cancelButtonText: '取消',
        type: 'warning'
    }).then(() => {
        chipList.value.splice(index, 1)
        ElMessage.success('删除成功')
        // 如果删除的是正在编辑的芯片，关闭 dialog
        if (editingIndex.value === index) {
            dialogVisible.value = false
            editingIndex.value = null
            editingChip.value = null
        } else if (editingIndex.value !== null && editingIndex.value > index) {
            // 如果删除的芯片在正在编辑的芯片之前，需要调整索引
            editingIndex.value--
        }
    }).catch(() => {})
}

const addCpu = () => {
    if (editingChip.value) {
        editingChip.value.cpu.push({
            fire: { rate: 0, core: 0 },
            ice: { rate: 0, core: 0 }
        })
    }
}

const removeCpu = (cpuIndex: number) => {
    if (editingChip.value) {
        editingChip.value.cpu.splice(cpuIndex, 1)
    }
}

const addGpu = () => {
    if (editingChip.value) {
        editingChip.value.gpu.push({
            brand: '',
            core: 0,
            info: ''
        })
    }
}

const removeGpu = (gpuIndex: number) => {
    if (editingChip.value) {
        editingChip.value.gpu.splice(gpuIndex, 1)
    }
}

const addDevice = () => {
    if (editingChip.value) {
        editingChip.value.devices.push('')
    }
}

const removeDevice = (deviceIndex: number) => {
    if (editingChip.value) {
        editingChip.value.devices.splice(deviceIndex, 1)
    }
}

const saveChip = (silent: boolean = false): Promise<boolean> => {
    if (editingIndex.value === null || !editingChip.value) {
        if (!silent) {
            ElMessage.warning('请选择要保存的芯片')
        }
        return Promise.resolve(false)
    }

    if (!diaryId.value || !diaryData.value) {
        if (!silent) {
            ElMessage.warning('请先加载数据')
        }
        return Promise.resolve(false)
    }

    if (!editingChip.value.name || !editingChip.value.name.trim()) {
        if (!silent) {
            ElMessage.warning('芯片名称不能为空')
        }
        return Promise.resolve(false)
    }

    isSavingChip.value = true
    
    // 更新列表中的数据
    chipList.value[editingIndex.value] = JSON.parse(JSON.stringify(editingChip.value))
    
    // 保存到服务器，保留原始数据的所有字段，只更新 content
    const content = JSON.stringify(chipList.value, null, 2)
    const requestData = {
        ...diaryData.value,
        id: diaryId.value!,
        title: formData.value.title || diaryData.value.title,
        content: content,
        date: diaryData.value.date ? dateFormatter(new Date(diaryData.value.date), 'yyyy-MM-dd HH:mm:ss') : diaryData.value.date
    }

    return diaryApi.modify(requestData)
        .then(() => {
            if (!silent) {
                ElMessage.success('保存成功')
            }
            originalChip.value = JSON.parse(JSON.stringify(editingChip.value)) // 更新原始数据
            return true
        })
        .catch((err: any) => {
            console.error('保存失败:', err)
            if (!silent) {
                ElMessage.error('保存失败: ' + (err.message || '未知错误'))
            }
            return false
        })
        .finally(() => {
            isSavingChip.value = false
        })
}

// 检查是否有修改
const hasChanges = (): boolean => {
    if (!editingChip.value || !originalChip.value) {
        return false
    }
    return JSON.stringify(editingChip.value) !== JSON.stringify(originalChip.value)
}

// 保存按钮点击处理
const handleSaveButton = () => {
    saveChip(false).then((success: boolean) => {
        if (success && editingIndex.value !== null) {
            dialogVisible.value = false
            editingIndex.value = null
            editingChip.value = null
            originalChip.value = null
        }
    })
}

// 对话框关闭前处理
const handleDialogClose = (done: () => void) => {
    // 如果是新增芯片且没有名称，直接关闭（可能是误操作）
    if (editingIndex.value === null || editingIndex.value === undefined) {
        // 新增芯片时，如果没有名称就忽略，有名称就保存
        if (editingChip.value && editingChip.value.name && editingChip.value.name.trim()) {
            saveChip(true).then((success: boolean) => {
                if (success) {
                    dialogVisible.value = false
                    editingIndex.value = null
                    editingChip.value = null
                    originalChip.value = null
                    done()
                }
            })
        } else {
            dialogVisible.value = false
            editingIndex.value = null
            editingChip.value = null
            originalChip.value = null
            done()
        }
        return
    }

    if (hasChanges()) {
        // 如果有修改，自动保存
        saveChip(true).then((success: boolean) => {
            if (success) {
                dialogVisible.value = false
                editingIndex.value = null
                editingChip.value = null
                originalChip.value = null
                done()
            }
            // 如果保存失败，不关闭对话框，让用户处理错误
        })
    } else {
        // 没有修改，直接关闭
        dialogVisible.value = false
        editingIndex.value = null
        editingChip.value = null
        originalChip.value = null
        done()
    }
}

const saveAllData = () => {
    if (!diaryId.value || !diaryData.value) {
        ElMessage.warning('请先加载数据')
        return
    }

    if (chipList.value.length === 0) {
        ElMessage.warning('请至少添加一个芯片')
        return
    }

    if (!formData.value.title || !formData.value.title.trim()) {
        ElMessage.warning('标题不能为空')
        return
    }

    isSaving.value = true
    
    // 将芯片列表转换为 JSON 字符串
    const content = JSON.stringify(chipList.value, null, 2)
    
    // 保留原始数据的所有字段，只更新 content
    const requestData = {
        ...diaryData.value,
        id: diaryId.value!,
        title: formData.value.title,
        content: content,
        date: diaryData.value.date ? dateFormatter(new Date(diaryData.value.date), 'yyyy-MM-dd HH:mm:ss') : diaryData.value.date
    }

    diaryApi.modify(requestData)
        .then(() => {
            ElMessage.success('保存成功')
            // 重新加载数据
            loadData().then(() => {
                initSortable()
            })
        })
        .catch((err: any) => {
            console.error('保存失败:', err)
            ElMessage.error('保存失败: ' + (err.message || '未知错误'))
        })
        .finally(() => {
            isSaving.value = false
        })
}

// 初始化拖拽排序
const initSortable = () => {
    if (sortableInstance) {
        sortableInstance.destroy()
        sortableInstance = null
    }
    
    nextTick(() => {
        const tableEl = tableRef.value?.$el
        if (!tableEl) return

        // 获取表格的 tbody 元素
        const tbody = tableEl.querySelector('.el-table__body-wrapper tbody')
        if (!tbody) return

        sortableInstance = new Sortable(tbody, {
            animation: 150,
            filter: '.el-table__empty-block', // 过滤空状态
            onEnd: (evt: any) => {
                const { oldIndex, newIndex } = evt
                if (oldIndex !== undefined && newIndex !== undefined && oldIndex !== newIndex) {
                    // 更新数据顺序
                    const movedItem = chipList.value.splice(oldIndex, 1)[0]
                    chipList.value.splice(newIndex, 0, movedItem)
                    
                    // 更新编辑索引
                    if (editingIndex.value !== null) {
                        if (editingIndex.value === oldIndex) {
                            editingIndex.value = newIndex
                        } else if (editingIndex.value > oldIndex && editingIndex.value <= newIndex) {
                            editingIndex.value--
                        } else if (editingIndex.value < oldIndex && editingIndex.value >= newIndex) {
                            editingIndex.value++
                        }
                    }
                    
                    ElMessage.success('排序已更新，请记得保存')
                }
            }
        })
    })
}

onMounted(() => {
    loadData().then(() => {
        initSortable()
    })
})

onUnmounted(() => {
    if (sortableInstance) {
        sortableInstance.destroy()
        sortableInstance = null
    }
})
</script>

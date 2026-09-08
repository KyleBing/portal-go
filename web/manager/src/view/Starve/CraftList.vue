<template>
    <Container>
        <Toolbar>
            <template #left>
                <ElForm size="small" inline>
                    <ElFormItem label="搜索">
                        <ElInput 
                            clearable 
                            placeholder="搜索中文名或英文名" 
                            v-model="keyword"
                            style="width: 200px"
                        />
                    </ElFormItem>
                    <ElFormItem label="版本">
                        <ElSelect v-model="selectedVersion" clearable placeholder="选择版本" style="width: 200px">
                            <ElOption label="全部" value="" />
                            <ElOption 
                                v-for="version in versionList" 
                                :key="version.id" 
                                :label="version.name" 
                                :value="version.id" 
                            />
                        </ElSelect>
                    </ElFormItem>
                    <ElFormItem label="标签">
                        <ElSelect v-model="selectedTab" clearable placeholder="选择标签" style="width: 200px">
                            <ElOption label="全部" value="" />
                            <ElOption 
                                v-for="tab in craftTabList" 
                                :key="tab.id" 
                                :label="tab.name" 
                                :value="tab.id" 
                            />
                        </ElSelect>
                    </ElFormItem>
                    <ElFormItem>
                        <ElButton type="primary" @click="handleSearch" icon="Search">搜索</ElButton>
                    </ElFormItem>
                </ElForm>
            </template>
            <template #center>
            </template>
            <template #right>
                <ElButton type="success" icon="Plus" @click="addNewCraft()"> 添加</ElButton>
            </template>
        </Toolbar>

        <Content padding="0">
            <ElRow :gutter="10">
                <ElCol :span="24">
                    <ElTable
                        class="table-narrow"
                        size="small"
                        :height="projectStore.contentInsets.heightContent - 130"
                        stripe
                        :data="filteredData"
                        v-loading="isLoading"
                    >
                        <ElTableColumn width="60" prop="id" label="ID" fixed="left"/>
                        <ElTableColumn width="100" prop="is_active" label="状态">
                            <template #default="{row}">
                                <ElTag :type="row.is_active === 1 ? 'success' : 'info'">
                                    {{ row.is_active === 1 ? '启用' : '禁用' }}
                                </ElTag>
                            </template>
                        </ElTableColumn>
                        <ElTableColumn width="150" prop="name" label="名称（中文）"/>
                        <ElTableColumn width="150" prop="name_en" label="名称（英文）"/>
                        <ElTableColumn width="100" prop="sortid" label="排序ID"/>
                        <ElTableColumn width="200" prop="version" label="版本">
                            <template #default="{row}">
                                <div v-if="row.version && Array.isArray(row.version) && row.version.length > 0">
                                    <ElTag 
                                        v-for="v in row.version" 
                                        :key="v.id" 
                                        :type="v.is_active === 1 ? 'success' : 'info'"
                                        style="margin-right: 5px; margin-bottom: 5px;"
                                    >
                                        {{ v.name }}
                                    </ElTag>
                                </div>
                                <span v-else>-</span>
                            </template>
                        </ElTableColumn>
                        <ElTableColumn width="200" prop="tab" label="标签">
                            <template #default="{row}">
                                <div v-if="row.tab && Array.isArray(row.tab) && row.tab.length > 0">
                                    <ElTag 
                                        v-for="t in row.tab" 
                                        :key="t.id" 
                                        :type="t.is_active === 1 ? 'success' : 'info'"
                                        style="margin-right: 5px; margin-bottom: 5px;"
                                    >
                                        {{ t.name }}
                                    </ElTag>
                                </div>
                                <span v-else>-</span>
                            </template>
                        </ElTableColumn>
                        <ElTableColumn prop="crafting" label="制作材料" show-overflow-tooltip/>
                        <ElTableColumn width="150" prop="tier" label="制作等级"/>
                        <ElTableColumn width="100" prop="damage" label="伤害值"/>
                        <ElTableColumn prop="sideeffect" label="副作用" show-overflow-tooltip/>
                        <ElTableColumn width="100" prop="durability" label="耐久度"/>
                        <ElTableColumn prop="perk" label="特殊效果" show-overflow-tooltip/>
                        <ElTableColumn width="80" prop="stacks" label="堆叠"/>
                        <ElTableColumn width="150" prop="debugspawn" label="调试代码" show-overflow-tooltip/>
                        <ElTableColumn align="center" width="200" label="操作" fixed="right">
                            <template #default="scope">
                                <ElButton @click="goEdit(scope.row)" type="primary" icon="edit" plain size="small">编辑</ElButton>
                                <ElButton @click="goDelete(scope.row)" type="danger" icon="delete" plain size="small">删除</ElButton>
                            </template>
                        </ElTableColumn>
                    </ElTable>
                </ElCol>
            </ElRow>
            <FooterPagination
                :pager-option="pager"
                @size-change="sizeChange"
                @pager-change="pagerChange"
            />
        </Content>

        <!-- 编辑窗口-->
        <ElDialog
            :title="modalTitle"
            v-model="modalEdit"
            width="70%"
            :before-close="closeModal">
            <div style="max-height: 70vh; overflow-y: auto;">
                <ElForm
                    :model="formCraft"
                    :rules="craftRules"
                    size="small"
                    ref="formRef" label-width="120px">
                    <ElRow :gutter="20">
                        <ElCol :span="12">
                            <ElFormItem label="名称（中文）" prop="name">
                                <ElInput v-model="formCraft.name" placeholder="请输入中文名称"/>
                            </ElFormItem>
                        </ElCol>
                        <ElCol :span="12">
                            <ElFormItem label="名称（英文）" prop="name_en">
                                <ElInput v-model="formCraft.name_en" placeholder="请输入英文名称"/>
                            </ElFormItem>
                        </ElCol>
                    </ElRow>
                    <ElRow :gutter="20">
                        <ElCol :span="12">
                            <ElFormItem label="排序ID" prop="sortid">
                                <ElInputNumber v-model="formCraft.sortid" placeholder="请输入排序ID" style="width: 100%"/>
                            </ElFormItem>
                        </ElCol>
                        <ElCol :span="12">
                            <ElFormItem label="制作等级" prop="tier">
                                <ElSelect v-model="formCraft.tier" placeholder="请选择制作等级" style="width: 100%">
                                    <ElOption label="始终可用" value="Always Available"/>
                                    <ElOption label="科学机器" value="Science Machine"/>
                                    <ElOption label="炼金引擎" value="Alchemy Engine"/>
                                    <ElOption label="灵子分解器" value="Prestihatitator"/>
                                    <ElOption label="暗影操纵者" value="Shadow Manipulator"/>
                                    <ElOption label="远古伪科学站" value="Ancient Pseudoscience Station"/>
                                    <ElOption label="黑曜石工作台" value="Obsidian Workbench"/>
                                    <ElOption label="制图师的桌子" value="Cartographer's Desk"/>
                                    <ElOption label="岩石巢穴" value="Rock Den"/>
                                    <ElOption label="草图" value="Sketch"/>
                                    <ElOption label="蓝图" value="Blueprint"/>
                                    <ElOption label="火鸡神龛" value="Gobbler Shrine"/>
                                </ElSelect>
                            </ElFormItem>
                        </ElCol>
                    </ElRow>
                    <ElRow :gutter="20">
                        <ElCol :span="12">
                            <ElFormItem label="版本" prop="version">
                                <ElSelect 
                                    v-model="formCraft.version" 
                                    multiple 
                                    placeholder="请选择版本（可多选）"
                                    style="width: 100%"
                                >
                                    <ElOption 
                                        v-for="version in versionList" 
                                        :key="version.id" 
                                        :label="version.name" 
                                        :value="version.id" 
                                    />
                                </ElSelect>
                            </ElFormItem>
                        </ElCol>
                        <ElCol :span="12">
                            <ElFormItem label="标签" prop="tab">
                                <ElSelect 
                                    v-model="formCraft.tab" 
                                    multiple 
                                    placeholder="请选择标签（可多选）"
                                    style="width: 100%"
                                >
                                    <ElOption 
                                        v-for="tab in craftTabList" 
                                        :key="tab.id" 
                                        :label="tab.name" 
                                        :value="tab.id" 
                                    />
                                </ElSelect>
                            </ElFormItem>
                        </ElCol>
                    </ElRow>
                    <ElFormItem label="制作材料" prop="crafting">
                        <ElInput type="textarea" :rows="3" v-model="formCraft.crafting" placeholder="请输入制作材料"/>
                    </ElFormItem>
                    <ElRow :gutter="20">
                        <ElCol :span="8">
                            <ElFormItem label="伤害值" prop="damage">
                                <ElInput v-model="formCraft.damage" placeholder="请输入伤害值"/>
                            </ElFormItem>
                        </ElCol>
                        <ElCol :span="8">
                            <ElFormItem label="耐久度" prop="durability">
                                <ElInput v-model="formCraft.durability" placeholder="请输入耐久度"/>
                            </ElFormItem>
                        </ElCol>
                        <ElCol :span="8">
                            <ElFormItem label="堆叠数量" prop="stacks">
                                <ElInput v-model="formCraft.stacks" placeholder="请输入堆叠数量"/>
                            </ElFormItem>
                        </ElCol>
                    </ElRow>
                    <ElFormItem label="副作用" prop="sideeffect">
                        <ElInput type="textarea" :rows="2" v-model="formCraft.sideeffect" placeholder="请输入副作用"/>
                    </ElFormItem>
                    <ElFormItem label="特殊效果" prop="perk">
                        <ElInput type="textarea" :rows="3" v-model="formCraft.perk" placeholder="请输入特殊效果"/>
                    </ElFormItem>
                    <ElRow :gutter="20">
                        <ElCol :span="12">
                            <ElFormItem label="调试代码" prop="debugspawn">
                                <ElInput v-model="formCraft.debugspawn" placeholder="请输入调试代码"/>
                            </ElFormItem>
                        </ElCol>
                        <ElCol :span="12">
                            <ElFormItem label="图片路径" prop="pic">
                                <ElInput v-model="formCraft.pic" placeholder="请输入图片路径"/>
                            </ElFormItem>
                        </ElCol>
                    </ElRow>
                    <ElFormItem label="缩略图路径" prop="thumb">
                        <ElInput v-model="formCraft.thumb" placeholder="请输入缩略图路径"/>
                    </ElFormItem>
                    <ElFormItem label="状态" prop="is_active">
                        <ElRadioGroup v-model="formCraft.is_active">
                            <ElRadio :label="1">启用</ElRadio>
                            <ElRadio :label="0">禁用</ElRadio>
                        </ElRadioGroup>
                    </ElFormItem>
                </ElForm>
            </div>
            <div slot="footer" class="dialog-footer">
                <ElButton size="small" @click="clearForm" type="warning" icon="RefreshLeft">清空</ElButton>
                <ElButton size="small" @click="closeModal" icon="Close">取 消</ElButton>
                <ElButton size="small" type="primary" @click="submit" icon="Check">确定</ElButton>
            </div>
        </ElDialog>
    </Container>
</template>

<script setup lang="ts">
import { ref, onMounted, computed, watch } from 'vue'
import { ElMessage, ElMessageBox } from 'element-plus'
import type { FormInstance } from 'element-plus'
import starveApi from "@/api/starveApi"
import { useProjectStore } from "@/pinia"
import Container from "@/layout/Container.vue";
import Toolbar from "@/layout/Toolbar.vue";
import Content from "@/layout/Content.vue";
import FooterPagination from "@/layout/FooterPagination.vue";
import { Craft, CraftTierEnum, Version, CraftTab } from "@/model/starve";

const projectStore = useProjectStore()
const isLoading = ref(false)
const tableData = ref<Craft[]>([])
const allData = ref<Craft[]>([])
const filteredData = ref<Craft[]>([])
const keyword = ref('')
const selectedVersion = ref<number | null>(null)
const selectedTab = ref<number | null>(null)
const versionList = ref<Version[]>([])
const craftTabList = ref<CraftTab[]>([])
let searchTimer: number | null = null
const editingId = ref<number | null>(null)

// 分页
const pager = ref({
    total: 0,
    pageNo: 1,
    pageSize: 20
})
const modalEdit = ref(false)
const formRef = ref<FormInstance>()

const formCraft = ref<Craft>({
    id: undefined,
    name: '',
    name_en: '',
    sortid: null,
    version: [],
    tab: [],
    crafting: '',
    tier: undefined,
    damage: null,
    sideeffect: null,
    durability: null,
    perk: null,
    stacks: null,
    debugspawn: null,
    pic: null,
    thumb: null,
    is_active: 1
})

const craftRules = {
    name: [
        { required: true, message: '请输入中文名称', trigger: 'blur' }
    ],
    name_en: [
        { required: true, message: '请输入英文名称', trigger: 'blur' }
    ],
    version: [
        { required: true, message: '请至少选择一个版本', trigger: 'change' }
    ],
    tab: [
        { required: true, message: '请至少选择一个标签', trigger: 'change' }
    ],
    crafting: [
        { required: true, message: '请输入制作材料', trigger: 'blur' }
    ]
}

const modalTitle = computed(() => editingId.value ? '编辑制作' : '新增制作')

function addNewCraft() {
    modalEdit.value = true
    editingId.value = null
    clearForm()
}

function clearForm() {
    formCraft.value = {
        id: undefined,
        name: '',
        name_en: '',
        sortid: null,
        version: [],
        tab: [],
        crafting: '',
        tier: undefined,
        damage: null,
        sideeffect: null,
        durability: null,
        perk: null,
        stacks: null,
        debugspawn: null,
        pic: null,
        thumb: null,
        is_active: 1
    }
}

function closeModal(done?: () => void) {
    ElMessageBox.confirm('确认关闭？')
        .then(() => {
            editingId.value = null
            modalEdit.value = false
            if (done) done()
        })
        .catch(() => {})
}

function getCraftList() {
    isLoading.value = true
    starveApi.craftList()
        .then((res: any) => {
            allData.value = Array.isArray(res) ? res : (res.data || [])
            tableData.value = allData.value
            handleSearch()
            isLoading.value = false
        })
        .catch(err => {
            isLoading.value = false
            console.error(err)
        })
}

function getVersionList() {
    starveApi.versionList()
        .then((res: any) => {
            versionList.value = Array.isArray(res) ? res : (res.data || [])
        })
        .catch(err => {
            console.error(err)
        })
}

function getCraftTabList() {
    starveApi.craftTabList()
        .then((res: any) => {
            craftTabList.value = Array.isArray(res) ? res : (res.data || [])
        })
        .catch(err => {
            console.error(err)
        })
}

function handleSearch() {
    let result = allData.value
    
    // 版本筛选
    if (selectedVersion.value) {
        result = result.filter(item => {
            if (!item.version || !Array.isArray(item.version)) return false
            return item.version.some((v: Version) => v.id === selectedVersion.value)
        })
    }
    
    // 标签筛选
    if (selectedTab.value) {
        result = result.filter(item => {
            if (!item.tab || !Array.isArray(item.tab)) return false
            return item.tab.some((t: CraftTab) => t.id === selectedTab.value)
        })
    }
    
    // 关键词搜索
    if (keyword.value && keyword.value.trim() !== '') {
        const searchKey = keyword.value.trim().toLowerCase()
        result = result.filter(item => {
            const name = (item.name || '').toLowerCase()
            const nameEn = (item.name_en || '').toLowerCase()
            return name.includes(searchKey) || nameEn.includes(searchKey)
        })
    }
    
    // 更新总数
    pager.value.total = result.length
    
    // 分页
    const start = (pager.value.pageNo - 1) * pager.value.pageSize
    const end = start + pager.value.pageSize
    filteredData.value = result.slice(start, end)
}

function sizeChange(size: number) {
    pager.value.pageSize = size
    pager.value.pageNo = 1
    handleSearch()
}

function pagerChange(pageNo: number) {
    pager.value.pageNo = pageNo
    handleSearch()
}

function goEdit(craft: Craft) {
    editingId.value = craft.id || null
    // 确保 version 和 tab 是数组格式
    const versionIds = Array.isArray(craft.version) 
        ? craft.version.map((v: Version) => v.id).filter(Boolean)
        : []
    const tabIds = Array.isArray(craft.tab) 
        ? craft.tab.map((t: CraftTab) => t.id).filter(Boolean)
        : []
    formCraft.value = { 
        ...craft,
        version: versionIds,
        tab: tabIds
    }
    modalEdit.value = true
}

function goDelete(craft: Craft) {
    ElMessageBox.confirm(`删除制作 ${craft.name}`, '删除', {
        confirmButtonText: '确定',
        cancelButtonText: '取消',
        type: 'warning'
    }).then(() => {
        if (!craft.id) return
        starveApi.craftDelete({ id: craft.id })
            .then(res => {
                getCraftList()
                ElMessage({
                    message: res.message || '删除成功',
                    type: 'success'
                })
            })
    })
}

function submit() {
    if (!formRef.value) return
    formRef.value.validate((valid) => {
        if (valid) {
            if (editingId.value) {
                craftModifySubmit()
            } else {
                craftAddSubmit()
            }
        } else {
            console.log('error submit!!')
        }
    })
}

function craftAddSubmit() {
    // 将版本和标签 ID 数组转换为对象数组
    const submitData = {
        ...formCraft.value,
        version: formCraft.value.version.map((id: number) => 
            versionList.value.find(v => v.id === id)
        ).filter(Boolean),
        tab: formCraft.value.tab.map((id: number) => 
            craftTabList.value.find(t => t.id === id)
        ).filter(Boolean)
    }
    starveApi.craftAdd(submitData)
        .then(() => {
            ElMessage({
                message: '添加成功',
                type: 'success'
            })
            getCraftList()
            editingId.value = null
            modalEdit.value = false
        })
        .catch(err => {
            console.error(err)
        })
}

function craftModifySubmit() {
    // 将版本和标签 ID 数组转换为对象数组
    const submitData = {
        ...formCraft.value,
        version: formCraft.value.version.map((id: number) => 
            versionList.value.find(v => v.id === id)
        ).filter(Boolean),
        tab: formCraft.value.tab.map((id: number) => 
            craftTabList.value.find(t => t.id === id)
        ).filter(Boolean)
    }
    starveApi.craftModify(submitData)
        .then(() => {
            ElMessage({
                message: '修改成功',
                type: 'success'
            })
            getCraftList()
            editingId.value = null
            modalEdit.value = false
        })
        .catch(err => {
            console.error(err)
        })
}

onMounted(() => {
    getVersionList()
    getCraftTabList()
    getCraftList()
})

watch(keyword, () => {
    if (searchTimer) {
        clearTimeout(searchTimer)
    }
    pager.value.pageNo = 1
    searchTimer = setTimeout(() => {
        handleSearch()
    }, 500)
})

watch(selectedVersion, () => {
    pager.value.pageNo = 1
    handleSearch()
})

watch(selectedTab, () => {
    pager.value.pageNo = 1
    handleSearch()
})
</script>

<style scoped lang="scss">
</style>


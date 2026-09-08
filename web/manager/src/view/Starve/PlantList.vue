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
                    <ElFormItem>
                        <ElButton type="primary" @click="handleSearch" icon="Search">搜索</ElButton>
                    </ElFormItem>
                </ElForm>
            </template>
            <template #center>
            </template>
            <template #right>
                <ElButton type="success" icon="Plus" @click="addNewPlant()"> 添加</ElButton>
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
                        <ElTableColumn prop="resources" label="资源产出" show-overflow-tooltip/>
                        <ElTableColumn width="120" prop="spawns" label="生成位置"/>
                        <ElTableColumn prop="perk" label="特殊效果" show-overflow-tooltip/>
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
            width="50%"
            :before-close="closeModal">
            <ElForm
                :model="formPlant"
                :rules="plantRules"
                size="small"
                ref="formRef" label-width="120px">
                <ElRow :gutter="20">
                    <ElCol :span="12">
                        <ElFormItem label="名称（中文）" prop="name">
                            <ElInput v-model="formPlant.name" placeholder="请输入中文名称"/>
                        </ElFormItem>
                    </ElCol>
                    <ElCol :span="12">
                        <ElFormItem label="名称（英文）" prop="name_en">
                            <ElInput v-model="formPlant.name_en" placeholder="请输入英文名称"/>
                        </ElFormItem>
                    </ElCol>
                </ElRow>
                <ElFormItem label="版本" prop="version">
                    <ElSelect 
                        v-model="formPlant.version" 
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
                <ElFormItem label="资源产出" prop="resources">
                    <ElInput type="textarea" :rows="2" v-model="formPlant.resources" placeholder="请输入资源产出"/>
                </ElFormItem>
                <ElFormItem label="生成位置" prop="spawns">
                    <ElInput v-model="formPlant.spawns" placeholder="请输入生成位置"/>
                </ElFormItem>
                <ElFormItem label="特殊效果" prop="perk">
                    <ElInput type="textarea" :rows="3" v-model="formPlant.perk" placeholder="请输入特殊效果"/>
                </ElFormItem>
                <ElRow :gutter="20">
                    <ElCol :span="12">
                        <ElFormItem label="调试代码" prop="debugspawn">
                            <ElInput v-model="formPlant.debugspawn" placeholder="请输入调试代码"/>
                        </ElFormItem>
                    </ElCol>
                    <ElCol :span="12">
                        <ElFormItem label="图片路径" prop="pic">
                            <ElInput v-model="formPlant.pic" placeholder="请输入图片路径"/>
                        </ElFormItem>
                    </ElCol>
                </ElRow>
                <ElFormItem label="缩略图路径" prop="thumb">
                    <ElInput v-model="formPlant.thumb" placeholder="请输入缩略图路径"/>
                </ElFormItem>
            </ElForm>
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
import { Plant, Version } from "@/model/starve";

const projectStore = useProjectStore()
const isLoading = ref(false)
const tableData = ref<Plant[]>([])
const allData = ref<Plant[]>([])
const filteredData = ref<Plant[]>([])
const keyword = ref('')
const selectedVersion = ref<number | null>(null)
const versionList = ref<Version[]>([])
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

const formPlant = ref<Plant>({
    id: undefined,
    name: '',
    name_en: '',
    version: [],
    resources: null,
    spawns: null,
    debugspawn: null,
    perk: null,
    pic: null,
    thumb: null,
    is_active: 1
})

const plantRules = {
    name: [
        { required: true, message: '请输入中文名称', trigger: 'blur' }
    ],
    name_en: [
        { required: true, message: '请输入英文名称', trigger: 'blur' }
    ],
    version: [
        { required: true, message: '请至少选择一个版本', trigger: 'change' }
    ]
}

const modalTitle = computed(() => editingId.value ? '编辑植物' : '新增植物')

function addNewPlant() {
    modalEdit.value = true
    editingId.value = null
    clearForm()
}

function clearForm() {
    formPlant.value = {
        id: undefined,
        name: '',
        name_en: '',
        version: [],
        is_active: 1,
        resources: null,
        spawns: null,
        debugspawn: null,
        perk: null,
        pic: null,
        thumb: null
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

function getPlantList() {
    isLoading.value = true
    starveApi.plantList()
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

function handleSearch() {
    let result = allData.value
    
    // 版本筛选
    if (selectedVersion.value) {
        result = result.filter(item => {
            if (!item.version || !Array.isArray(item.version)) return false
            return item.version.some((v: Version) => v.id === selectedVersion.value)
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

function goEdit(plant: Plant) {
    editingId.value = plant.id || null
    const versionIds = Array.isArray(plant.version) 
        ? plant.version.map((v: Version) => v.id).filter(Boolean)
        : []
    formPlant.value = { 
        ...plant,
        version: versionIds
    }
    modalEdit.value = true
}

function goDelete(plant: Plant) {
    ElMessageBox.confirm(`删除植物 ${plant.name}`, '删除', {
        confirmButtonText: '确定',
        cancelButtonText: '取消',
        type: 'warning'
    }).then(() => {
        if (!plant.id) return
        starveApi.plantDelete({ id: plant.id })
            .then(res => {
                getPlantList()
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
                plantModifySubmit()
            } else {
                plantAddSubmit()
            }
        } else {
            console.log('error submit!!')
        }
    })
}

function plantAddSubmit() {
    const submitData = {
        ...formPlant.value,
        version: formPlant.value.version.map((id: number) => 
            versionList.value.find(v => v.id === id)
        ).filter(Boolean)
    }
    starveApi.plantAdd(submitData)
        .then(() => {
            ElMessage({
                message: '添加成功',
                type: 'success'
            })
            getPlantList()
            editingId.value = null
            modalEdit.value = false
        })
        .catch(err => {
            console.error(err)
        })
}

function plantModifySubmit() {
    const submitData = {
        ...formPlant.value,
        version: formPlant.value.version.map((id: number) => 
            versionList.value.find(v => v.id === id)
        ).filter(Boolean)
    }
    starveApi.plantModify(submitData)
        .then(() => {
            ElMessage({
                message: '修改成功',
                type: 'success'
            })
            getPlantList()
            editingId.value = null
            modalEdit.value = false
        })
        .catch(err => {
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

onMounted(() => {
    getVersionList()
    getPlantList()
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
</script>

<style scoped lang="scss">
</style>

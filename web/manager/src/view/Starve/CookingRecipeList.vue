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
                <ElButton type="success" icon="Plus" @click="addNewCookingRecipe()"> 添加</ElButton>
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
                        <ElTableColumn width="100" align="right" prop="health_value" label="生命值"/>
                        <ElTableColumn width="100" align="right" prop="hungry_value" label="饥饿值"/>
                        <ElTableColumn width="100" align="right" prop="sanity_value" label="理智值"/>
                        <ElTableColumn width="100" align="right" prop="duration" label="持续时间"/>
                        <ElTableColumn width="100" prop="cook_time" label="烹饪时间"/>
                        <ElTableColumn width="100" prop="priority" label="优先级"/>
                        <ElTableColumn prop="requirements" label="需求材料" show-overflow-tooltip/>
                        <ElTableColumn prop="restrictions" label="限制条件" show-overflow-tooltip/>
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
            <div style="max-height: 70vh;">
                <ElForm
                    :model="formCookingRecipe"
                    :rules="cookingRecipeRules"
                    size="small"
                    ref="formRef" label-width="120px">
                    <ElRow :gutter="20">
                        <ElCol :span="12">
                            <ElFormItem label="名称（中文）" prop="name">
                                <ElInput v-model="formCookingRecipe.name" placeholder="请输入中文名称"/>
                            </ElFormItem>
                        </ElCol>
                        <ElCol :span="12">
                            <ElFormItem label="名称（英文）" prop="name_en">
                                <ElInput v-model="formCookingRecipe.name_en" placeholder="请输入英文名称"/>
                            </ElFormItem>
                        </ElCol>
                    </ElRow>
                    <ElFormItem label="版本" prop="version">
                        <ElSelect 
                            v-model="formCookingRecipe.version" 
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
                    <ElRow :gutter="20">
                        <ElCol :span="8">
                            <ElFormItem label="生命值" prop="health_value">
                                <ElInputNumber v-model="formCookingRecipe.health_value" placeholder="请输入生命值" style="width: 100%"/>
                            </ElFormItem>
                        </ElCol>
                        <ElCol :span="8">
                            <ElFormItem label="饥饿值" prop="hungry_value">
                                <ElInputNumber v-model="formCookingRecipe.hungry_value" placeholder="请输入饥饿值" style="width: 100%"/>
                            </ElFormItem>
                        </ElCol>
                        <ElCol :span="8">
                            <ElFormItem label="理智值" prop="sanity_value">
                                <ElInputNumber v-model="formCookingRecipe.sanity_value" placeholder="请输入理智值" style="width: 100%"/>
                            </ElFormItem>
                        </ElCol>
                    </ElRow>
                    <ElRow :gutter="20">
                        <ElCol :span="12">
                            <ElFormItem label="持续时间" prop="duration">
                                <ElInputNumber v-model="formCookingRecipe.duration" placeholder="0 为永久" style="width: 100%"/>
                            </ElFormItem>
                        </ElCol>
                        <ElCol :span="12">
                            <ElFormItem label="烹饪时间" prop="cook_time">
                                <ElInput v-model="formCookingRecipe.cook_time" placeholder="请输入烹饪时间"/>
                            </ElFormItem>
                        </ElCol>
                    </ElRow>
                    <ElFormItem label="优先级" prop="priority">
                        <ElInput v-model="formCookingRecipe.priority" placeholder="请输入优先级"/>
                    </ElFormItem>
                    <ElFormItem label="需求材料" prop="requirements">
                        <ElInput type="textarea" :rows="2" v-model="formCookingRecipe.requirements" placeholder="请输入需求材料"/>
                    </ElFormItem>
                    <ElFormItem label="限制条件" prop="restrictions">
                        <ElInput type="textarea" :rows="2" v-model="formCookingRecipe.restrictions" placeholder="请输入限制条件"/>
                    </ElFormItem>
                    <ElFormItem label="特殊效果" prop="perk">
                        <ElInput type="textarea" :rows="3" v-model="formCookingRecipe.perk" placeholder="请输入特殊效果"/>
                    </ElFormItem>
                    <ElRow :gutter="20">
                        <ElCol :span="12">
                            <ElFormItem label="堆叠数量" prop="stacks">
                                <ElInput v-model="formCookingRecipe.stacks" placeholder="请输入堆叠数量"/>
                            </ElFormItem>
                        </ElCol>
                        <ElCol :span="12">
                            <ElFormItem label="调试代码" prop="debugspawn">
                                <ElInput v-model="formCookingRecipe.debugspawn" placeholder="请输入调试代码"/>
                            </ElFormItem>
                        </ElCol>
                    </ElRow>
                    <ElRow :gutter="20">
                        <ElCol :span="12">
                            <ElFormItem label="图片路径" prop="pic">
                                <ElInput v-model="formCookingRecipe.pic" placeholder="请输入图片路径"/>
                            </ElFormItem>
                        </ElCol>
                        <ElCol :span="12">
                            <ElFormItem label="缩略图路径" prop="thumb">
                                <ElInput v-model="formCookingRecipe.thumb" placeholder="请输入缩略图路径"/>
                            </ElFormItem>
                        </ElCol>
                    </ElRow>
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
import { CookingRecipe, Version } from "@/model/starve";

const projectStore = useProjectStore()
const isLoading = ref(false)
const tableData = ref<CookingRecipe[]>([])
const allData = ref<CookingRecipe[]>([])
const filteredData = ref<CookingRecipe[]>([])
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

const formCookingRecipe = ref<CookingRecipe>({
    id: undefined,
    name: '',
    name_en: '',
    version: [],
    is_active: 1,
    health_value: null,
    hungry_value: null,
    sanity_value: null,
    duration: null,
    cook_time: null,
    priority: null,
    requirements: null,
    restrictions: null,
    perk: null,
    stacks: null,
    debugspawn: null,
    pic: null,
    thumb: null
})

const cookingRecipeRules = {
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

const modalTitle = computed(() => editingId.value ? '编辑烹饪食谱' : '新增烹饪食谱')

function addNewCookingRecipe() {
    modalEdit.value = true
    editingId.value = null
    clearForm()
}

function clearForm() {
    formCookingRecipe.value = {
        id: undefined,
        name: '',
        name_en: '',
        version: [],
        is_active: 1,
        health_value: null,
        hungry_value: null,
        sanity_value: null,
        duration: null,
        cook_time: null,
        priority: null,
        requirements: null,
        restrictions: null,
        perk: null,
        stacks: null,
        debugspawn: null,
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

function getCookingRecipeList() {
    isLoading.value = true
    starveApi.cookingrecipeList()
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

function goEdit(cookingRecipe: CookingRecipe) {
    editingId.value = cookingRecipe.id || null
    const versionIds = Array.isArray(cookingRecipe.version) 
        ? cookingRecipe.version.map((v: Version) => v.id).filter(Boolean)
        : []
    formCookingRecipe.value = { 
        ...cookingRecipe,
        version: versionIds
    }
    modalEdit.value = true
}

function goDelete(cookingRecipe: CookingRecipe) {
    ElMessageBox.confirm(`删除烹饪食谱 ${cookingRecipe.name}`, '删除', {
        confirmButtonText: '确定',
        cancelButtonText: '取消',
        type: 'warning'
    }).then(() => {
        if (!cookingRecipe.id) return
        starveApi.cookingrecipeDelete({ id: cookingRecipe.id })
            .then(res => {
                getCookingRecipeList()
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
                cookingRecipeModifySubmit()
            } else {
                cookingRecipeAddSubmit()
            }
        } else {
            console.log('error submit!!')
        }
    })
}

function cookingRecipeAddSubmit() {
    const submitData = {
        ...formCookingRecipe.value,
        version: formCookingRecipe.value.version.map((id: number) => 
            versionList.value.find(v => v.id === id)
        ).filter(Boolean)
    }
    starveApi.cookingrecipeAdd(submitData)
        .then(() => {
            ElMessage({
                message: '添加成功',
                type: 'success'
            })
            getCookingRecipeList()
            editingId.value = null
            modalEdit.value = false
        })
        .catch(err => {
            console.error(err)
        })
}

function cookingRecipeModifySubmit() {
    const submitData = {
        ...formCookingRecipe.value,
        version: formCookingRecipe.value.version.map((id: number) => 
            versionList.value.find(v => v.id === id)
        ).filter(Boolean)
    }
    starveApi.cookingrecipeModify(submitData)
        .then(() => {
            ElMessage({
                message: '修改成功',
                type: 'success'
            })
            getCookingRecipeList()
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
    getCookingRecipeList()
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

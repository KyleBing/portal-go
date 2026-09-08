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
                <ElButton type="success" icon="Plus" @click="addNewMob()"> 添加</ElButton>
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
                        <ElTableColumn width="100" prop="kind" label="类型"/>
                        <ElTableColumn width="100" prop="size" label="大小"/>
                        <ElTableColumn width="100" prop="health" label="生命值"/>
                        <ElTableColumn width="100" prop="damage" label="伤害值"/>
                        <ElTableColumn width="120" prop="attack_period" label="攻击周期"/>
                        <ElTableColumn width="120" prop="attack_range" label="攻击范围"/>
                        <ElTableColumn width="120" prop="walking_speed" label="行走速度"/>
                        <ElTableColumn width="120" prop="running_speed" label="奔跑速度"/>
                        <ElTableColumn width="100" prop="sanityaura" label="理智光环"/>
                        <ElTableColumn prop="special_ability" label="特殊能力" show-overflow-tooltip/>
                        <ElTableColumn prop="detail" label="详细信息" show-overflow-tooltip/>
                        <ElTableColumn prop="loot" label="掉落物品" show-overflow-tooltip width="150"/>
                        <ElTableColumn prop="spawns_from" label="生成来源" show-overflow-tooltip/>
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
            width="60%"
            :before-close="closeModal">
            <ElForm
                :model="formMob"
                :rules="mobRules"
                size="small"
                ref="formRef" label-width="120px">
                <ElRow :gutter="20">
                    <ElCol :span="12">
                        <ElFormItem label="名称（中文）" prop="name">
                            <ElInput v-model="formMob.name" placeholder="请输入中文名称"/>
                        </ElFormItem>
                    </ElCol>
                    <ElCol :span="12">
                        <ElFormItem label="名称（英文）" prop="name_en">
                            <ElInput v-model="formMob.name_en" placeholder="请输入英文名称"/>
                        </ElFormItem>
                    </ElCol>
                </ElRow>
                <ElRow :gutter="20">
                    <ElCol :span="8">
                        <ElFormItem label="版本" prop="version">
                            <ElSelect 
                                v-model="formMob.version" 
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
                    <ElCol :span="8">
                        <ElFormItem label="类型" prop="kind">
                            <ElRadioGroup v-model="formMob.kind">
                                <ElRadio label="neutral">中立</ElRadio>
                                <ElRadio label="friendly">友好</ElRadio>
                                <ElRadio label="hostile">敌对</ElRadio>
                            </ElRadioGroup>
                        </ElFormItem>
                    </ElCol>
                    <ElCol :span="8">
                        <ElFormItem label="大小" prop="size">
                            <ElRadioGroup v-model="formMob.size">
                                <ElRadio label="small">小型</ElRadio>
                                <ElRadio label="middle">中型</ElRadio>
                                <ElRadio label="large">大型</ElRadio>
                            </ElRadioGroup>
                        </ElFormItem>
                    </ElCol>
                </ElRow>
                <ElRow :gutter="20">
                    <ElCol :span="12">
                        <ElFormItem label="生命值" prop="health">
                            <ElInput v-model="formMob.health" placeholder="请输入生命值"/>
                        </ElFormItem>
                    </ElCol>
                    <ElCol :span="12">
                        <ElFormItem label="伤害值" prop="damage">
                            <ElInput v-model="formMob.damage" placeholder="请输入伤害值"/>
                        </ElFormItem>
                    </ElCol>
                </ElRow>
                <ElRow :gutter="20">
                    <ElCol :span="12">
                        <ElFormItem label="攻击周期" prop="attack_period">
                            <ElInput v-model="formMob.attack_period" placeholder="请输入攻击周期"/>
                        </ElFormItem>
                    </ElCol>
                    <ElCol :span="12">
                        <ElFormItem label="攻击范围" prop="attack_range">
                            <ElInput v-model="formMob.attack_range" placeholder="请输入攻击范围"/>
                        </ElFormItem>
                    </ElCol>
                </ElRow>
                <ElRow :gutter="20">
                    <ElCol :span="12">
                        <ElFormItem label="行走速度" prop="walking_speed">
                            <ElInput v-model="formMob.walking_speed" placeholder="请输入行走速度"/>
                        </ElFormItem>
                    </ElCol>
                    <ElCol :span="12">
                        <ElFormItem label="奔跑速度" prop="running_speed">
                            <ElInput v-model="formMob.running_speed" placeholder="请输入奔跑速度"/>
                        </ElFormItem>
                    </ElCol>
                </ElRow>
                <ElFormItem label="理智光环" prop="sanityaura">
                    <ElInput v-model="formMob.sanityaura" placeholder="请输入理智光环"/>
                </ElFormItem>
                <ElFormItem label="特殊能力" prop="special_ability">
                    <ElInput type="textarea" :rows="2" v-model="formMob.special_ability" placeholder="请输入特殊能力"/>
                </ElFormItem>
                <ElFormItem label="详细信息" prop="detail">
                    <ElInput type="textarea" :rows="2" v-model="formMob.detail" placeholder="请输入详细信息"/>
                </ElFormItem>
                <ElFormItem label="掉落物品" prop="loot">
                    <ElInput type="textarea" :rows="2" v-model="formMob.loot" placeholder="请输入掉落物品"/>
                </ElFormItem>
                <ElFormItem label="生成来源" prop="spawns_from">
                    <ElInput type="textarea" :rows="2" v-model="formMob.spawns_from" placeholder="请输入生成来源"/>
                </ElFormItem>
                <ElRow :gutter="20">
                    <ElCol :span="12">
                        <ElFormItem label="调试代码" prop="debugspawn">
                            <ElInput v-model="formMob.debugspawn" placeholder="请输入调试代码"/>
                        </ElFormItem>
                    </ElCol>
                    <ElCol :span="12">
                        <ElFormItem label="图片路径" prop="pic">
                            <ElInput v-model="formMob.pic" placeholder="请输入图片路径"/>
                        </ElFormItem>
                    </ElCol>
                </ElRow>
                <ElFormItem label="缩略图路径" prop="thumb">
                    <ElInput v-model="formMob.thumb" placeholder="请输入缩略图路径"/>
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
import { Mob, MobKindEnum, MobSizeEnum, Version } from "@/model/starve";

const projectStore = useProjectStore()
const isLoading = ref(false)
const tableData = ref<Mob[]>([])
const allData = ref<Mob[]>([])
const filteredData = ref<Mob[]>([])
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

const formMob = ref<Mob>({
    id: undefined,
    name: '',
    name_en: '',
    version: [],
    kind: MobKindEnum.Neutral,
    size: MobSizeEnum.Middle,
    health: null,
    damage: null,
    attack_period: null,
    attack_range: null,
    walking_speed: null,
    running_speed: null,
    sanityaura: null,
    special_ability: null,
    detail: null,
    loot: null,
    spawns_from: null,
    debugspawn: null,
    pic: null,
    thumb: null,
    is_active: 1
})

const mobRules = {
    name: [
        { required: true, message: '请输入中文名称', trigger: 'blur' }
    ],
    name_en: [
        { required: true, message: '请输入英文名称', trigger: 'blur' }
    ],
    version: [
        { required: true, message: '请至少选择一个版本', trigger: 'change' }
    ],
    kind: [
        { required: true, message: '请选择类型', trigger: 'change' }
    ],
    size: [
        { required: true, message: '请选择大小', trigger: 'change' }
    ]
}

const modalTitle = computed(() => editingId.value ? '编辑生物' : '新增生物')

function addNewMob() {
    modalEdit.value = true
    editingId.value = null
    clearForm()
}

function clearForm() {
    formMob.value = {
        id: undefined,
        name: '',
        name_en: '',
        version: [],
        kind: MobKindEnum.Neutral,
        size: MobSizeEnum.Middle,
        health: null,
        damage: null,
        attack_period: null,
        attack_range: null,
        walking_speed: null,
        running_speed: null,
        sanityaura: null,
        special_ability: null,
        detail: null,
        loot: null,
        spawns_from: null,
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

function getMobList() {
    isLoading.value = true
    starveApi.mobList()
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

function goEdit(mob: Mob) {
    editingId.value = mob.id || null
    // 确保 version 是数组格式，如果是对象则转换为数组
    const versionIds = Array.isArray(mob.version) 
        ? mob.version.map((v: Version) => v.id).filter(Boolean)
        : []
    formMob.value = { 
        ...mob,
        version: versionIds
    }
    modalEdit.value = true
}

function goDelete(mob: Mob) {
    ElMessageBox.confirm(`删除生物 ${mob.name}`, '删除', {
        confirmButtonText: '确定',
        cancelButtonText: '取消',
        type: 'warning'
    }).then(() => {
        if (!mob.id) return
        starveApi.mobDelete({ id: mob.id })
            .then(res => {
                getMobList()
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
                mobModifySubmit()
            } else {
                mobAddSubmit()
            }
        } else {
            console.log('error submit!!')
        }
    })
}

function mobAddSubmit() {
    // 将版本 ID 数组转换为版本对象数组
    const submitData = {
        ...formMob.value,
        version: formMob.value.version.map((id: number) => 
            versionList.value.find(v => v.id === id)
        ).filter(Boolean)
    }
    starveApi.mobAdd(submitData)
        .then(() => {
            ElMessage({
                message: '添加成功',
                type: 'success'
            })
            getMobList()
            editingId.value = null
            modalEdit.value = false
        })
        .catch(err => {
            console.error(err)
        })
}

function mobModifySubmit() {
    // 将版本 ID 数组转换为版本对象数组
    const submitData = {
        ...formMob.value,
        version: formMob.value.version.map((id: number) => 
            versionList.value.find(v => v.id === id)
        ).filter(Boolean)
    }
    starveApi.mobModify(submitData)
        .then(() => {
            ElMessage({
                message: '修改成功',
                type: 'success'
            })
            getMobList()
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
    getMobList()
})

// 监听 keyword 变化，0.5秒后触发搜索
watch(keyword, () => {
    if (searchTimer) {
        clearTimeout(searchTimer)
    }
    pager.value.pageNo = 1
    searchTimer = setTimeout(() => {
        handleSearch()
    }, 500)
})

// 监听版本筛选变化，立即触发搜索
watch(selectedVersion, () => {
    pager.value.pageNo = 1
    handleSearch()
})
</script>

<style scoped lang="scss">
</style>

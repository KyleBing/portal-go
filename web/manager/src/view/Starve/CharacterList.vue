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
                <ElButton type="success" icon="Plus" @click="addNewCharacter()"> 添加</ElButton>
            </template>
        </Toolbar>

        <Content padding="0">
            <ElRow :gutter="10">
                <ElCol :span="24">
                    <ElTable
                        class="table-narrow"
                        size="small"
                        :height="projectStore.contentInsets.heightContent - 80"
                        stripe
                        :data="filteredData"
                        v-loading="isLoading"
                    >
                        <ElTableColumn width="60" prop="id" label="ID" fixed="left"/>
                        <ElTableColumn width="150" prop="name" label="名称（中文）"/>
                        <ElTableColumn width="150" prop="name_en" label="名称（英文）"/>
                        <ElTableColumn width="120" prop="nick_name" label="昵称"/>
                        <ElTableColumn width="200" prop="version" label="版本">
                            <template #default="{row}">
                                <div v-if="row.version && Array.isArray(row.version) && row.version.length > 0">
                                    <ElTag 
                                        v-for="v in row.version" 
                                        :key="v.id" 
                                        style="margin-right: 5px;"
                                    >
                                        {{ v.name }}
                                    </ElTag>
                                </div>
                                <span v-else>-</span>
                            </template>
                        </ElTableColumn>
                        <ElTableColumn width="100" prop="health" label="生命值"/>
                        <ElTableColumn width="100" prop="hunger" label="饥饿值"/>
                        <ElTableColumn width="100" prop="sanity" label="理智值"/>
                        <ElTableColumn width="120" prop="hunger_modifier" label="饥饿值修正"/>
                        <ElTableColumn width="120" prop="sanity_modifier" label="理智值修正"/>
                        <ElTableColumn width="120" prop="wetness_modifier" label="湿度修正"/>
                        <ElTableColumn width="120" prop="health_range" label="生命值范围"/>
                        <ElTableColumn width="120" prop="damage_range" label="伤害范围"/>
                        <ElTableColumn width="120" prop="hunger_range" label="饥饿值范围"/>
                        <ElTableColumn width="120" prop="sanity_range" label="理智值范围"/>
                        <ElTableColumn width="120" prop="speed_range" label="速度范围"/>
                        <ElTableColumn prop="motto" label="座右铭" show-overflow-tooltip/>
                        <ElTableColumn prop="perk" label="特殊能力" show-overflow-tooltip/>
                        <ElTableColumn width="120" prop="special_item" label="特殊物品"/>
                        <ElTableColumn width="120" prop="starting_item" label="起始物品"/>
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
        </Content>

        <!-- 编辑窗口-->
        <ElDialog
            :title="modalTitle"
            v-model="modalEdit"
            width="70%"
            :before-close="closeModal">
            <div style="max-height: 70vh;">
                <ElForm
                    :model="formCharacter"
                    :rules="characterRules"
                    size="small"
                    ref="formRef" label-width="120px">
                    <ElRow :gutter="20">
                        <ElCol :span="12">
                            <ElFormItem label="名称（中文）" prop="name">
                                <ElInput v-model="formCharacter.name" placeholder="请输入中文名称"/>
                            </ElFormItem>
                        </ElCol>
                        <ElCol :span="12">
                            <ElFormItem label="名称（英文）" prop="name_en">
                                <ElInput v-model="formCharacter.name_en" placeholder="请输入英文名称"/>
                            </ElFormItem>
                        </ElCol>
                    </ElRow>
                    <ElRow :gutter="20">
                        <ElCol :span="12">
                            <ElFormItem label="昵称" prop="nick_name">
                                <ElInput v-model="formCharacter.nick_name" placeholder="请输入昵称"/>
                            </ElFormItem>
                        </ElCol>
                        <ElCol :span="12">
                            <ElFormItem label="版本" prop="version">
                                <ElSelect 
                                    v-model="formCharacter.version" 
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
                    </ElRow>
                    <ElRow :gutter="20">
                        <ElCol :span="8">
                            <ElFormItem label="生命值" prop="health">
                                <ElInput v-model="formCharacter.health" placeholder="请输入生命值"/>
                            </ElFormItem>
                        </ElCol>
                        <ElCol :span="8">
                            <ElFormItem label="饥饿值" prop="hunger">
                                <ElInput v-model="formCharacter.hunger" placeholder="请输入饥饿值"/>
                            </ElFormItem>
                        </ElCol>
                        <ElCol :span="8">
                            <ElFormItem label="理智值" prop="sanity">
                                <ElInput v-model="formCharacter.sanity" placeholder="请输入理智值"/>
                            </ElFormItem>
                        </ElCol>
                    </ElRow>
                    <ElRow :gutter="20">
                        <ElCol :span="8">
                            <ElFormItem label="饥饿值修正" prop="hunger_modifier">
                                <ElInput v-model="formCharacter.hunger_modifier" placeholder="请输入饥饿值修正"/>
                            </ElFormItem>
                        </ElCol>
                        <ElCol :span="8">
                            <ElFormItem label="理智值修正" prop="sanity_modifier">
                                <ElInput v-model="formCharacter.sanity_modifier" placeholder="请输入理智值修正"/>
                            </ElFormItem>
                        </ElCol>
                        <ElCol :span="8">
                            <ElFormItem label="湿度修正" prop="wetness_modifier">
                                <ElInput v-model="formCharacter.wetness_modifier" placeholder="请输入湿度修正"/>
                            </ElFormItem>
                        </ElCol>
                    </ElRow>
                    <ElRow :gutter="20">
                        <ElCol :span="8">
                            <ElFormItem label="生命值范围" prop="health_range">
                                <ElInput v-model="formCharacter.health_range" placeholder="请输入生命值范围"/>
                            </ElFormItem>
                        </ElCol>
                        <ElCol :span="8">
                            <ElFormItem label="伤害范围" prop="damage_range">
                                <ElInput v-model="formCharacter.damage_range" placeholder="请输入伤害范围"/>
                            </ElFormItem>
                        </ElCol>
                        <ElCol :span="8">
                            <ElFormItem label="饥饿值范围" prop="hunger_range">
                                <ElInput v-model="formCharacter.hunger_range" placeholder="请输入饥饿值范围"/>
                            </ElFormItem>
                        </ElCol>
                    </ElRow>
                    <ElRow :gutter="20">
                        <ElCol :span="8">
                            <ElFormItem label="理智值范围" prop="sanity_range">
                                <ElInput v-model="formCharacter.sanity_range" placeholder="请输入理智值范围"/>
                            </ElFormItem>
                        </ElCol>
                        <ElCol :span="8">
                            <ElFormItem label="速度范围" prop="speed_range">
                                <ElInput v-model="formCharacter.speed_range" placeholder="请输入速度范围"/>
                            </ElFormItem>
                        </ElCol>
                    </ElRow>
                    <ElFormItem label="座右铭" prop="motto">
                        <ElInput v-model="formCharacter.motto" placeholder="请输入座右铭"/>
                    </ElFormItem>
                    <ElFormItem label="特殊能力" prop="perk">
                        <ElInput type="textarea" :rows="2" v-model="formCharacter.perk" placeholder="请输入特殊能力"/>
                    </ElFormItem>
                    <ElRow :gutter="20">
                        <ElCol :span="12">
                            <ElFormItem label="特殊物品" prop="special_item">
                                <ElInput v-model="formCharacter.special_item" placeholder="请输入特殊物品"/>
                            </ElFormItem>
                        </ElCol>
                        <ElCol :span="12">
                            <ElFormItem label="起始物品" prop="starting_item">
                                <ElInput v-model="formCharacter.starting_item" placeholder="请输入起始物品"/>
                            </ElFormItem>
                        </ElCol>
                    </ElRow>
                    <ElRow :gutter="20">
                        <ElCol :span="12">
                            <ElFormItem label="调试代码" prop="debugspawn">
                                <ElInput v-model="formCharacter.debugspawn" placeholder="请输入调试代码"/>
                            </ElFormItem>
                        </ElCol>
                        <ElCol :span="12">
                            <ElFormItem label="图片路径" prop="pic">
                                <ElInput v-model="formCharacter.pic" placeholder="请输入图片路径"/>
                            </ElFormItem>
                        </ElCol>
                    </ElRow>
                    <ElFormItem label="缩略图路径" prop="thumb">
                        <ElInput v-model="formCharacter.thumb" placeholder="请输入缩略图路径"/>
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
import { Character, Version } from "@/model/starve";

const projectStore = useProjectStore()
const isLoading = ref(false)
const tableData = ref<Character[]>([])
const allData = ref<Character[]>([])
const filteredData = ref<Character[]>([])
const keyword = ref('')
const selectedVersion = ref<number | null>(null)
const versionList = ref<Version[]>([])
let searchTimer: number | null = null
const editingId = ref<number | null>(null)
const modalEdit = ref(false)
const formRef = ref<FormInstance>()

const formCharacter = ref<Character & { version: number[] }>({
    id: undefined,
    name: null,
    name_en: null,
    nick_name: null,
    motto: null,
    perk: null,
    health: '',
    hunger: '',
    sanity: '',
    hunger_modifier: null,
    sanity_modifier: null,
    wetness_modifier: null,
    health_range: null,
    damage_range: null,
    hunger_range: null,
    sanity_range: null,
    speed_range: null,
    debugspawn: null,
    pic: null,
    thumb: null,
    special_item: null,
    starting_item: null,
    version: [],
    is_active: 1
})

const characterRules = {
    health: [
        { required: true, message: '请输入生命值', trigger: 'blur' }
    ],
    hunger: [
        { required: true, message: '请输入饥饿值', trigger: 'blur' }
    ],
    sanity: [
        { required: true, message: '请输入理智值', trigger: 'blur' }
    ],
    version: [
        { required: true, message: '请选择版本', trigger: 'change' }
    ]
}

const modalTitle = computed(() => editingId.value ? '编辑角色' : '新增角色')

function addNewCharacter() {
    modalEdit.value = true
    editingId.value = null
    clearForm()
}

function clearForm() {
    formCharacter.value = {
        id: undefined,
        name: null,
        name_en: null,
        nick_name: null,
        motto: null,
        perk: null,
        health: '',
        hunger: '',
        sanity: '',
        hunger_modifier: null,
        sanity_modifier: null,
        wetness_modifier: null,
        health_range: null,
        damage_range: null,
        hunger_range: null,
        sanity_range: null,
        speed_range: null,
        debugspawn: null,
        pic: null,
        thumb: null,
        special_item: null,
        starting_item: null,
        version: [],
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

function getCharacterList() {
    isLoading.value = true
    starveApi.characterList()
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
    
    filteredData.value = result
}

function goEdit(character: Character) {
    editingId.value = character.id || null
    const versionIds = Array.isArray(character.version) 
        ? character.version.map((v: Version) => v.id).filter((id): id is number => id !== undefined && id !== null)
        : []
    formCharacter.value = { 
        ...character,
        version: versionIds
    } as Character & { version: number[] }
    modalEdit.value = true
}

function goDelete(character: Character) {
    ElMessageBox.confirm(`删除角色 ${character.name || character.name_en}`, '删除', {
        confirmButtonText: '确定',
        cancelButtonText: '取消',
        type: 'warning'
    }).then(() => {
        if (!character.id) return
        starveApi.characterDelete({ id: character.id })
            .then(res => {
                getCharacterList()
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
                characterModifySubmit()
            } else {
                characterAddSubmit()
            }
        } else {
            console.log('error submit!!')
        }
    })
}

function characterAddSubmit() {
    const submitData = {
        ...formCharacter.value,
        version: formCharacter.value.version.map((id: number) => 
            versionList.value.find(v => v.id === id)
        ).filter((v): v is Version => v !== undefined)
    }
    starveApi.characterAdd(submitData)
        .then(() => {
            ElMessage({
                message: '添加成功',
                type: 'success'
            })
            getCharacterList()
            editingId.value = null
            modalEdit.value = false
        })
        .catch(err => {
            console.error(err)
        })
}

function characterModifySubmit() {
    const submitData = {
        ...formCharacter.value,
        version: formCharacter.value.version.map((id: number) => 
            versionList.value.find(v => v.id === id)
        ).filter((v): v is Version => v !== undefined)
    }
    starveApi.characterModify(submitData)
        .then(() => {
            ElMessage({
                message: '修改成功',
                type: 'success'
            })
            getCharacterList()
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
    getCharacterList()
})

watch(keyword, () => {
    if (searchTimer) {
        clearTimeout(searchTimer)
    }
    searchTimer = setTimeout(() => {
        handleSearch()
    }, 500)
})

watch(selectedVersion, () => {
    handleSearch()
})
</script>

<style scoped lang="scss">
</style>

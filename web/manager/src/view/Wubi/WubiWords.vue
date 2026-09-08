<template>
    <Container>
        <Toolbar :is-show-search-bar="true">
            <template #left>
                <ElForm size="" inline>
                    <ElFormItem>
                        <ElButton type="success" @click="addNewWord" icon="Plus"> 添加</ElButton>
                        <ElButton type="success" @click="addNewWordBatch" icon="Upload"> 批量添加</ElButton>
                        <ElButton v-if="isAdmin" type="primary" @click="showModalModifyBatchCategory" icon="MagicStick"> 修改类别</ElButton>
                        <ElButton v-if="isAdmin" type="primary" @click="showModalModifyBatchApproved" icon="MagicStick"> 修改状态</ElButton>
                        <ElButton type="danger" @click="deleteBatch" icon="Delete"> 删除</ElButton>
                    </ElFormItem>
                    <ElFormItem label="审核状态" class="ml-2">
                        <ElSelect v-model="formSearch.approved" style="width: 100px">
                            <ElOption value="" label="全部"/>
                            <ElOption :value="0" label="待审核"/>
                            <ElOption :value="1" label="已通过"/>
                            <ElOption :value="2" label="已拒绝"/>
                        </ElSelect>
                    </ElFormItem>
                    <ElFormItem label="关键字" class="ml-2">
                        <ElInput clearable placeholder="检索词条、编码、注释" v-model="formSearch.keyword"/>
                    </ElFormItem>
                    <ElFormItem label="添加时间区间">
                        <ElDatePicker type="datetimerange" v-model="formSearch.dateRange"/>
                    </ElFormItem>
                    <ElFormItem>
                        <ElButton type="primary" @click="search" icon="Search"> 查询</ElButton>
                    </ElFormItem>
                </ElForm>
            </template>
            <template #center>
            </template>
            <template #right>
            </template>
        </Toolbar>

        <!-- Table -->
        <Content padding="0">
            <ElRow :gutter="10">
                <ElCol :span="4">
                    <side-wubi-category-list
                        :style="`height: ${projectStore.contentInsets.heightContent - projectStore.contentInsets.heightToolbar - projectStore.contentInsets.heightPagination}px`"
                        :categories="wubiCategoryArray"
                        :activeCategoryId="formSearch.category_id"
                        @click="changeCurrentActiveCategoryId"
                    />
                </ElCol>
                <ElCol :span="20">
                    <ElTable
                        v-if="tableData.length > 0"
                        @selection-change="tableSelectionChange"
                        :height="projectStore.contentInsets.heightContent - projectStore.contentInsets.heightToolbar - projectStore.contentInsets.heightPagination"
                        :data="tableData" v-loading="isLoading"
                        class="table-narrow" size="small" stripe
                    >
                        <ElTableColumn type="selection" width="50"/>
                        <ElTableColumn width="60" prop="id" label="ID"/>
                        <ElTableColumn prop="word" min-width="150px" label="词条"/>
                        <ElTableColumn width="60" align="left" prop="code" label="编码">
                            <template #default="{ row }">
                                <span class="font-jetbrains">{{ row.code }}</span>
                            </template>
                        </ElTableColumn>
                        <ElTableColumn align="center" prop="priority" label="权重"/>
                        <ElTableColumn align="center" width="100" prop="comment" label="备注"/>
                        <ElTableColumn align="center" width="200px" label="操作">
                            <template #default="{ row }" >
                                <div v-if="isAdmin || selfUid === row.uid_init">
                                    <ElButton @click="goEdit(row)" type="primary" plain size="small"> 编辑</ElButton>
                                    <ElButton @click="goDelete([row.id])" type="danger" plain size="small"> 删除</ElButton>
                                </div>
                            </template>
                        </ElTableColumn>
                        <ElTableColumn align="center" width="50" prop="up" label="支持"/>
                        <ElTableColumn align="center" width="50" prop="down" label="反对"/>
                        <ElTableColumn align="center" prop="category_id" label="分组">
                            <template #default="{ row }">
                                {{ wubiCategoryMap.get(row.category_id)?.name }}
                            </template>
                        </ElTableColumn>
                        <ElTableColumn align="center" width="180" prop="date_create" label="创建时间">
                            <template #default="{row}">
                                <span v-if="row.date_create">{{Moment(row.date_create).format('YYYY-MM-DD HH:mm:ss')}}</span>
                            </template>
                        </ElTableColumn>
                        <ElTableColumn align="center" width="180" prop="date_modify" label="编辑时间">
                            <template #default="{row}">
                                <span v-if="row.date_modify">{{Moment(row.date_modify).format('YYYY-MM-DD HH:mm:ss')}}</span>
                            </template>
                        </ElTableColumn>
                        <ElTableColumn align="center" width="50" prop="uid_create" label="状态">
                            <template #default="{ row }">
                                <i v-if="row.approved === 1" class="text-approved ElIconCheck"></i>
                                <i v-else-if="row.approved === 2" class="text-reject ElIconClose"></i>
                                <i v-else-if="row.approved === 0" class="text-pending ElIconMinus"></i>
                            </template>
                        </ElTableColumn>
                        <ElTableColumn align="left" width="150" prop="user_init" label="添加者">
                            <template #default="{ row }">
                                {{ row.nickname_init }} <span class="">({{ row.group_id_init === 1 ? '管' : '普' }})</span>
                            </template>
                        </ElTableColumn>
                        <ElTableColumn align="left" width="150" prop="user_modify" label="修改者">
                            <template #default="{ row }">
                                {{ row.nickname_modify }} <span class="">({{ row.group_id_modify === 1 ? '管' : '普' }})</span>
                            </template>
                        </ElTableColumn>
                    </ElTable>

                </ElCol>
            </ElRow>
        </Content>
        <!--  PAGINATION  -->
        <FooterPagination
            :pager-option="pager"
            @size-change="sizeChange"
            @pager-change="pagerChange"
        />

        <!--  DIALOG 添加词条  -->
        <ElDialog
            :title="modalTitle"
            v-model="modalEdit"
            width="40%"
            :before-close="closeModalWordNew">
            <ElForm
                :model="formWord"
                :rules="wordRules"
                size="small"
                ref="formNewWord" label-width="100px">
                <ElFormItem label="词条" prop="word">
                    <ElInput placeholder="请输入词条" autocomplete="off" v-model.lazy="formWord.word"/>
                </ElFormItem>
                <ElFormItem label="编码" prop="code">
                    <ElInput placeholder="请输入编码" autocomplete="off" v-model.lazy="formWord.code"/>
                </ElFormItem>
                <ElFormItem v-if="existWords.length > 0 && !editingWordId">
                    <div class="exist-word-list">
                        <div class="exist-word-list-item" v-for="item in existWords" :key="item.id">
                            <div>[ {{ wubiCategoryMap.get(item.category_id)?.name }} ]</div>
                            <div>{{ item.priority }}</div>
                            <div>{{ item.code }}</div>
                            <div>{{ item.word }}</div>
                        </div>
                    </div>
                </ElFormItem>
                <ElFormItem label="权重" prop="priority">
                    <ElInput autocomplete="off" type="number" v-model="formWord.priority"/>
                </ElFormItem>
                <ElFormItem label="支持" prop="up">
                    <ElInput autocomplete="off" type="number" v-model="formWord.up"/>
                </ElFormItem>
                <ElFormItem label="不支持" prop="down">
                    <ElInput autocomplete="off" type="number" v-model="formWord.down"/>
                </ElFormItem>
                <ElFormItem label="注释" prop="comment">
                    <ElInput placeholder="注释词条"  autocomplete="off" v-model="formWord.comment"/>
                </ElFormItem>
                <ElFormItem label="组别" prop="category_id">
                    <ElRadioGroup v-model="formWord.category_id">
                        <ElRadio v-for="item in wubiCategoryArray"
                                  :key="item.id"
                                  :label="item.id">{{ item.name }}
                        </ElRadio>
                    </ElRadioGroup>
                </ElFormItem>
            </ElForm>
            <div slot="footer" class="dialog-footer">
                <ElButton size="small" @click="clearForm" type="warning"> 清空</ElButton>
                <ElButton size="small" @click="closeModalWordNew"> 取 消</ElButton>
                <ElButton size="small" type="primary" @click="submit">{{ editingWordId ? ' 修改' : ' 添加' }}</ElButton>
            </div>
        </ElDialog>

        <!--  批量添加词条  -->
        <ElDialog
            :model="formBatch"
            :rules="{}"
            title="批量添加词条"
            v-model="modalAddWordBatch"
            width="40%"
            :before-close="closeModalWordBatch">
            <ElForm
                size="small"
                ref="formNewWordBatch" label-width="100px">
                <ElFormItem label="已选词条 ID">
                    <ElInput clearable type="textarea" :rows="10" :disabled="!isAdmin" autocomplete="off" v-model.lazy="formBatch.dict_string"/>
                </ElFormItem>

                <ElFormItem label="组别" prop="category_id">
                    <ElRadioGroup v-model="formBatch.category_id">
                        <ElRadio v-for="item in wubiCategoryArray"
                                  :key="item.id"
                                  :label="item.id">{{ item.name }}
                        </ElRadio>
                    </ElRadioGroup>
                </ElFormItem>

            </ElForm>
            <div slot="footer" class="dialog-footer">
                <ElButton size="small" @click="formBatch.dict_string = ''" type="warning"> 清空</ElButton>
                <ElButton size="small" @click="closeModalWordBatch"> 取 消</ElButton>
                <ElButton size="small" type="primary" @click="generateWords"> 测试码表内容</ElButton>
                <ElButton size="small" type="primary" @click="submitBatchWords"> 上传</ElButton>
            </div>
        </ElDialog>

        <!--  批量修改词条类别  -->
        <ElDialog
            :model="formModifyWordBatchCategory"
            :rules="{}"
            title="批量修改词条类别"
            v-model="modalModifyWordBatchCategory"
            width="40%">
            <ElForm
                size="small"
                ref="formNewWordBatch" label-width="100px">
                <ElFormItem label="已选词条">
                    <ElInput type="textarea" disabled :value="tableData.filter(item => selectedIds.includes(item.id)).map(item => item.word).join(', ')" />
                </ElFormItem>

                <ElFormItem label="组别" prop="category_id">
                    <ElRadioGroup v-model="formModifyWordBatchCategory.category_id">
                        <ElRadio v-for="item in wubiCategoryArray"
                                  :key="item.id"
                                  :label="item.id">{{ item.name }}
                        </ElRadio>
                    </ElRadioGroup>
                </ElFormItem>

            </ElForm>
            <div slot="footer" class="dialog-footer">
                <ElButton size="small" @click="modalModifyWordBatchCategory = false"> 取 消</ElButton>
                <ElButton size="small" type="primary" @click="submitModifyWordCategoryBatch"> 确定修改</ElButton>
            </div>
        </ElDialog>

        <!--  批量修改词条审核状态  -->
        <ElDialog
            :model="formModifyWordBatchApproved"
            :rules="{}"
            title="批量修改词条审核状态"
            v-model="modalModifyWordBatchApproved"
            width="40%">
            <ElForm
                size="small"
                ref="formNewWordBatch" label-width="100px">
                <ElFormItem label="已选词条">
                    <ElInput type="textarea" disabled :value="tableData.filter(item => selectedIds.includes(item.id)).map(item => item.word).join(', ')" />
                </ElFormItem>

                <ElFormItem label="审核状态">
                    <ElRadio v-model="formModifyWordBatchApproved.approved" :label="0">未审核</ElRadio>
                    <ElRadio v-model="formModifyWordBatchApproved.approved" :label="1">通过</ElRadio>
                    <ElRadio v-model="formModifyWordBatchApproved.approved" :label="2">拒绝</ElRadio>
                </ElFormItem>

            </ElForm>
            <div slot="footer" class="dialog-footer">
                <ElButton size="small" @click="modalModifyWordBatchApproved = false"> 取 消</ElButton>
                <ElButton size="small" type="primary" @click="submitModifyWordApprovedBatch"> 确定修改</ElButton>
            </div>
        </ElDialog>
    </Container>
</template>

<script setup lang="ts">
import { ref, computed, watch, onMounted } from 'vue'
import { ElMessage, ElMessageBox, ElLoading } from 'element-plus'
import type { FormInstance } from 'element-plus'
import { DictOther } from "./DictOther"
import wubiApi from "@/api/wubiApi"
import SideWubiCategoryList from "@/view/Wubi/SideWubiCategoryList.vue"
import { Word } from "@/view/Wubi/Word"
import { WordGroup } from "@/view/Wubi/WordGroup"
import {useProjectStore} from "@/pinia";
import Container from "@/layout/Container.vue";
import Content from "@/layout/Content.vue";
import Toolbar from "@/layout/Toolbar.vue";
import {downloadBase64File, getAuthorization} from "@/utility";
import FooterPagination from "@/layout/FooterPagination.vue";
import Moment from "moment";

// Types
interface WordListRequest {
    category_id: number;
    dateRange: string[];
    keyword: string;
    pageNo: number;
    pageSize: number;
    approved: number;
}

interface WordRequest {
    word: string;
    code: string;
    priority: number;
    up: number;
    down: number;
    comment: string;
    category_id: number;
    uid?: string;
}

interface WordItem {
    id: number;
    word: string;
    code: string;
    priority: number;
    comment: string;
    category_id: number;
    date_create: string;
    date_modify: string;
    approved: number;
    uid_init: number;
    nickname_init: string;
    group_id_init: number;
    uid_modify: number;
    nickname_modify: string;
    group_id_modify: number;
}

interface CategoryItem {
    id: number;
    name: string;
}

interface ApiResponse<T> {
    data: T;
    message: string;
}

interface Pager {
    total: number,
    pageNo: number,
    pageSize: number
}

interface WordListResponse {
    pager: Pager;
    list: WordItem[];
}

interface DictWord {
    word: string;
    code: string;
    priority: number;
    comment: string;
}

const projectStore = useProjectStore()

function changeCurrentActiveCategoryId(id: number){
    formSearch.value.category_id = id
    pager.value.pageNo = 1
}

// State
const isLoading = ref(false)
const editingWordId = ref<number | null>(null)
const wubiCategoryMap = ref(new Map<number, CategoryItem>())
const wubiCategoryArray = ref<CategoryItem[]>([])
const existWords = ref<WordItem[]>([])
const tableData = ref<WordItem[]>([])
const formSearch = ref({
    category_id: 0,
    approved: '', // 0: pending, 1: approved, 2: rejected
    keyword: '',
    dateRange: [] as string[]
})

// Pager
const pager = ref<Pager>({
    pageSize: 30,
    pageNo: 1,
    total: 0
})

function sizeChange(size: number){
    pager.value.pageSize = size
    getWordList()
}

function pagerChange(pageNo: number){
    pager.value.pageNo = pageNo
    getWordList()
}


// Modal states
const formNewWord = ref<FormInstance>()
const modalEdit = ref(false)
const modalAddWordBatch = ref(false)
const modalModifyWordBatchCategory = ref(false)
const modalModifyWordBatchApproved = ref(false)

// Form data
const formWord = ref<WordRequest>({
    word: '',
    code: '',
    priority: 0,
    up: 0,
    down: 0,
    comment: '',
    category_id: 0,
    uid: ''
})

const formBatch = ref({
    dict_string: '',
    category_id: 1
})

const formModifyWordBatchCategory = ref({
    category_id: 1
})

const formModifyWordBatchApproved = ref({
    approved: 1 // 0: pending, 1:fine, 2:reject
})

onMounted(() => {
    getWubiCategoryList()
})

// Batch processing
const currentChunkIndex = ref(0)
const countAdded = ref(0)
const countExisted = ref(0)
const wordsChunks = ref<DictWord[][]>([])
const batchLoading = ref(null)
const selectedIds = ref<number[]>([])
const dict = ref<DictOther | null>(null)

// Computed
const modalTitle = computed(() => editingWordId.value ? '编辑词条' : '新增词条')
const isAdmin = computed(() => getAuthorization().email === 'kylebing@163.com')
const selfUid = computed(() => getAuthorization().uid)

// Form validation rules
const wordRules = {
    word: [
        {required: true, message: '请填写词条', trigger: 'blur'},
        {
            validator: (rule: any, value: string, callback: Function) => {
                if (/.*\t.*/.test(value)) {
                    callback(new Error("词条内容不能包含 tab"))
                } else {
                    callback()
                }
            }
        }
    ],
    code: {required: true, message: '请填写编码', trigger: 'blur'},
    category_id: {required: true, message: '请选择组别', trigger: 'blur'},
}

// Methods
function getWubiCategoryList() {
    wubiApi.category.list()
        .then(res => {
            wubiCategoryArray.value = res.data
            res.data.forEach((category: any) => {
                wubiCategoryMap.value.set(category.id, category)
            })
            getWordList()
        })
}

function getWordList() {
    isLoading.value = true
    const requestData: WordListRequest = {
        category_id: formSearch.value.category_id,
        dateRange: formSearch.value.dateRange.map(date => date.toString()),
        keyword: formSearch.value.keyword,
        pageNo: pager.value.pageNo,
        pageSize: pager.value.pageSize,
        approved: formSearch.value.approved
    }
    wubiApi.word.list(requestData)
        .then((res: ApiResponse<WordListResponse>) => {
            if (res && res.data) {
                pager.value = res.data.pager
                tableData.value = res.data.list
            }
        })
        .catch(err => {
            console.error(err)
        })
        .finally(() => {
            isLoading.value = false
        })
}

function search() {
    pager.value = {
        pageNo: 1,
        pageSize: pager.value.pageSize,
        total: pager.value.total
    }
    getWordList()
}

function addNewWord() {
    clearForm()
    existWords.value = []
    modalEdit.value = true
    editingWordId.value = null
    formWord.value.category_id = formSearch.value.category_id
    formWord.value.word = formSearch.value.keyword
}

function clearForm() {
    formWord.value = {
        word: '',
        code: '',
        priority: 0,
        up: 0,
        down: 0,
        comment: '',
        category_id: 0,
        uid: ''
    }
}

function closeModalWordNew(done: Function) {
    if (done) {
        done()
    } else {
        modalEdit.value = false
    }
}

function closeModalWordBatch(done: Function) {
    if (done) {
        done()
    } else {
        modalAddWordBatch.value = false
    }
}

function submit() {
    formNewWord.value?.validate((valid: boolean) => {
        if (valid) {
            if (editingWordId.value) {
                wordModifySubmit()
            } else {
                wordNewSubmit()
            }
        }
    })
}

function wordModifySubmit() {
    wubiApi.word.modify({
        ...formWord.value,
        id: editingWordId.value!
    })
        .then(() => {
            ElMessage.success('修改成功')
            modalEdit.value = false
            getWordList()
        })
}

function wordNewSubmit() {
    wubiApi.word.add(formWord.value)
        .then(() => {
            ElMessage.success('添加成功')
            modalEdit.value = false
            getWordList()
        })
}

function goEdit(word: any) {
    editingWordId.value = word.id
    formWord.value = {
        word: word.word,
        code: word.code,
        priority: word.priority,
        up: word.up,
        down: word.down,
        comment: word.comment,
        category_id: word.category_id,
        uid: ''
    }
    modalEdit.value = true
}

function goDelete(ids: number[]) {
    ElMessageBox.confirm('确定要删除这些词条吗？', '提示', {
        confirmButtonText: '确定',
        cancelButtonText: '取消',
        type: 'warning'
    }).then(() => {
        wubiApi.word.delete({ ids })
            .then(() => {
                ElMessage.success('删除成功')
                getWordList()
            })
    }).catch(() => {})
}

function deleteBatch() {
    if (selectedIds.value.length === 0) {
        ElMessage.warning('请选择要删除的词条')
        return
    }
    goDelete(selectedIds.value)
}

function showModalModifyBatchCategory() {
    if (selectedIds.value.length === 0) {
        ElMessage.warning('请选择要修改的词条')
        return
    }
    modalModifyWordBatchCategory.value = true
}

function showModalModifyBatchApproved() {
    if (selectedIds.value.length === 0) {
        ElMessage.warning('请选择要修改的词条')
        return
    }
    modalModifyWordBatchApproved.value = true
}

function tableSelectionChange(selections: any[]) {
    selectedIds.value = selections.map(item => item.id)
}

function generateWords() {
    if (!formBatch.value.dict_string) {
        ElMessage.warning('请输入词条内容')
        return
    }
    dict.value = new DictOther(formBatch.value.dict_string, 'temp', '', '\t')
    if (dict.value) {
        ElMessage.success(`识别到 ${dict.value.wordsOrigin.length} 个词条内容`)
    }
}

function addNewWordBatch() {
    modalAddWordBatch.value = true
    formBatch.value.dict_string = ''
}

function submitBatchWords() {
    if (!formBatch.value.dict_string) {
        ElMessage.warning('请输入词条内容')
        return
    }
    currentChunkIndex.value = 0
    countAdded.value = 0
    countExisted.value = 0
    wordsChunks.value = []

    generateWords()

    if (dict.value) {
        dict.value.wordsOrigin.forEach((word: DictWord, index: number) => {
            if (index % 100 === 0) {
                wordsChunks.value.push([])
            }
            wordsChunks.value[wordsChunks.value.length - 1].push(word)
        })
    }

    const loading = ElLoading.service({
        lock: true,
        text: "词条上传中，请稍候...",
        background: 'rgba(0, 0, 0, 0.7)'
    })

    addNewWords(wordsChunks.value)
        .finally(() => {
            loading.close()
        })
}

function addNewWords(wordsChunk: any[]) {
    return wubiApi.word.addBatch({
        category_id: formBatch.value.category_id,
        words: wordsChunk[currentChunkIndex.value]
    })
        .then(res => {
            countExisted.value += res.data.existCount
            countAdded.value += res.data.addedCount
            currentChunkIndex.value++

            if (currentChunkIndex.value < wordsChunks.value.length) {
                return addNewWords(wordsChunks.value)
            } else {
                addNewWordsFinished()
            }
        })
        .catch(err => {
            console.error(err)
        })
}

function addNewWordsFinished() {
    let message = `添加 ${countAdded.value} 条`
    if (countExisted.value > 0) {
        message = message + `，已存在词条 ${countExisted.value} 条`
    }
    ElMessage.success(message)
}

function submitModifyWordCategoryBatch() {
    if (selectedIds.value.length === 0) {
        ElMessage.warning('请选择要修改的词条')
        return
    }
    wubiApi.word.modifyBatch({
        ids: selectedIds.value,
        category_id: formModifyWordBatchCategory.value.category_id
    })
        .then(() => {
            ElMessage.success('修改成功')
            modalModifyWordBatchCategory.value = false
            getWordList()
        })
}

function submitModifyWordApprovedBatch() {
    if (selectedIds.value.length === 0) {
        ElMessage.warning('请选择要修改的词条')
        return
    }
    wubiApi.word.modifyBatch({
        ids: selectedIds.value,
        approved: formModifyWordBatchApproved.value.approved
    })
        .then(() => {
            ElMessage.success('修改成功')
            modalModifyWordBatchApproved.value = false
            getWordList()
        })
}

function exportExtraWords() {
    wubiApi.word.exportExtraWords()
        .then(res => {
            let wordGroups = []
            let lastCategoryName = ''
            res.data.forEach((item: any) => {
                if (lastCategoryName !== item.category_name) {
                    wordGroups.push(new WordGroup(
                        item.category_id,
                        item.category_name,
                        [new Word(item.id, item.code, item.word, item.priority, item.comment)]
                    ))
                } else {
                    wordGroups[wordGroups.length - 1].dict.push(new Word(item.id, item.code, item.word, item.priority, item.comment))
                }
                lastCategoryName = item.category_name
            })
            downloadBase64File('wubi86_jidian_extra.yaml', toYamlString(wordGroups))
        })
        .catch(err => {
            console.error(err)
        })
}

// Watchers
watch(formWord, (newValue) => {
    if (newValue.word || newValue.code) {
        wubiApi.word.checkExist({
            word: newValue.word,
            code: newValue.code
        })
            .then(res => {
                existWords.value = res.data
            })
            .catch(err => {
                console.error(err)
            })
    }
}, { deep: true })
watch(formSearch, (newValue) => {
    getWordList()
}, { deep: true })
</script>

<style scoped>
/* Add your styles here */
</style>

<template>
    <Container>
        <Toolbar>
            <template #left>
            </template>
            <template #center>
            </template>
            <template #right>
                <ElButton type="success" @click="addNewDiaryCategory" icon="Plus"> 添加</ElButton>
            </template>
        </Toolbar>
        <Content padding="0">
            <ElRow :gutter="10">
                <ElCol :span="24" >
                    <ElTable size="small" class="table-narrow" stripe :data="categoryList" v-loading="isLoading">
                        <ElTableColumn align="center" width="80" prop="id" sortable label="ID"/>
                        <ElTableColumn align="left" width="100" prop="name" label="类别"/>
                        <ElTableColumn width="100" align="right" prop="sort_id" sortable label="排列序号"/>
                        <ElTableColumn width="100" align="right" prop="count" label="词条数量"/>
                        <ElTableColumn align="center" width="200" prop="date_init" label="添加时间">
                            <template #default="scope">
                                <TableListDate :dates="[scope.row.date_init]" :names="['创建']"/>
                            </template>
                        </ElTableColumn>
                        <ElTableColumn align="left" label="操作" width="">
                            <template #default="scope">
                                <ElButton @click="goEdit(scope.row)" type="primary" icon="Edit" plain size="small"> 编辑</ElButton>
                                <ElButton @click="goDelete(scope.row)" type="danger" icon="delete" plain size="small"> 删除</ElButton>
                            </template>
                        </ElTableColumn>
                    </ElTable>
                </ElCol>
            </ElRow>
        </Content>

        <ElDialog
            :title="modalTitle"
            v-model="modalEdit"
            width="600px"
            :before-close="closeModal">
            <ElForm
                label-position="right"
                label-width="120px"
                size="small"
                :model="formCategory"
                :rules="categoryRules"
                ref="categoryModify">
                <ElRow>
                    <ElCol>
                        <ElFormItem label="类别名" prop="name">
                            <ElInput autocomplete="off" :disabled="!isAdmin" v-model="formCategory.name"></ElInput>
                        </ElFormItem>
                    </ElCol>
                    <ElCol>
                        <ElFormItem label="排序序号" prop="sort_id">
                            <ElInput type="number" autocomplete="off" :disabled="!isAdmin" v-model="formCategory.sort_id"></ElInput>
                        </ElFormItem>
                    </ElCol>
                </ElRow>

            </ElForm>
            <div slot="footer" class="dialog-footer">
                <ElButton size="small" @click="clearForm" type="warning"> 清空</ElButton>
                <ElButton size="small" @click="closeModal"> 取 消</ElButton>
                <ElButton size="small" type="primary" @click="submit">{{ editingCategoryId ? ' 修改' : ' 添加' }}</ElButton>
            </div>
        </ElDialog>
    </Container>
</template>

<script setup lang="ts">
import { ref, computed, onMounted } from 'vue';
import { ElMessageBox, ElNotification } from 'element-plus';
import wubiApi from "@/api/wubiApi";
import { getAuthorization, dateFormatter } from "@/utility";
import Container from "@/layout/Container.vue";
import Toolbar from "@/layout/Toolbar.vue";
import Content from "@/layout/Content.vue";
import TableListDate from "@/components/TableListDate.vue";

interface Category {
    id: number;
    name: string;
    sort_id: number;
    count: number;
    date_init: string;
}

const isLoading = ref(false);
const categoryList = ref<Category[]>([]);
const editingCategoryId = ref<number | null>(null);
const modalEdit = ref(false);
const categoryModify = ref();

const formCategory = ref({
    name: '',
    sort_id: 0
});

const categoryRules = {
    name: { required: true, message: '请填写名字', trigger: 'blur' },
    sort_id: { required: true, message: '请填写排序序号', trigger: 'blur' },
};

const isAdmin = ref(false);

const modalTitle = computed(() => editingCategoryId.value ? '编辑类别' : '新增类别');

onMounted(() => {
    getCategoryList();
    isAdmin.value = getAuthorization().email === 'kylebing@163.com';
});

function addNewDiaryCategory() {
    modalEdit.value = true;
    editingCategoryId.value = null;
    clearForm();
}

function clearForm() {
    formCategory.value = {
        name: '',
        sort_id: 0
    };
}

function closeModal() {
    editingCategoryId.value = null;
    modalEdit.value = false;
}

function getCategoryList() {
    isLoading.value = true;
    const params = {
        pageNo: 1,
        pageSize: 50
    };
    wubiApi.category.list(params)
        .then(res => {
            isLoading.value = false;
            categoryList.value = res.data
        })
        .catch(() => {
            isLoading.value = false;
        });
}

function goEdit(category: Category) {
    editingCategoryId.value = category.id;
    formCategory.value = { ...category };
    modalEdit.value = true;
}

function goDelete(category: Category) {
    ElMessageBox.confirm(`删除类别 ${category.name}`, '删除', {
        confirmButtonText: '确定',
        cancelButtonText: '取消',
        type: 'warning'
    }).then(() => {
        const requestData = {
            id: category.id
        };
        wubiApi.category.delete(requestData)
            .then(res => {
                getCategoryList();
                ElNotification({
                    title: res.message,
                    position: 'top-right',
                    type: 'success'
                });
            });
    });
}

function submit() {
    categoryModify.value?.validate((valid: boolean) => {
        if (valid) {
            if (editingCategoryId.value) {
                categoryModifySubmit();
            } else {
                categoryNewSubmit();
            }
        } else {
            console.log('error submit!!');
            return false;
        }
    });
}

function categoryNewSubmit() {
    wubiApi.category.add(formCategory.value)
        .then(res => {
            ElNotification({
                title: res.message,
                position: 'top-right',
                type: 'success'
            });
            getCategoryList();
            editingCategoryId.value = null;
            modalEdit.value = false;
        });
}

function categoryModifySubmit() {
    const requestData = {
        ...formCategory.value,
        id: editingCategoryId.value
    };
    wubiApi.category.modify(requestData)
        .then(res => {
            ElNotification({
                title: res.message,
                position: 'top-right',
                type: 'success'
            });
            getCategoryList();
            editingCategoryId.value = null;
            modalEdit.value = false;
        });
}
</script>

<style scoped lang="scss">
.table-description {
    white-space: nowrap;
    overflow: hidden;
    text-overflow: ellipsis;
    cursor: pointer;
}
</style>

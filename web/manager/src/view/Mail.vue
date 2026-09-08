<template>
    <Container>
        <Toolbar>
            <template #left>
            </template>
            <template #center>
            </template>
            <template #right>
                <ElButton @click="clearForm" type="warning" icon="RefreshLeft">清空</ElButton>
                <ElButton type="primary" @click="submit" icon="Promotion">发送</ElButton>
            </template>
        </Toolbar>

        <Content padding="0">
            <ElRow :gutter="10">
                <ElCol :span="24">
                    <ElForm
                        size="small"
                        :model="mail"
                        :rules="rules"
                        ref="codeModify" label-width="100px">
                        <ElFormItem label="标题" prop="title">
                            <ElInput v-model="mail.title"/>
                        </ElFormItem>
                        <ElFormItem label="内容" prop="content">
                            <ElInput :rows="10" type="textarea" v-model="mail.content"/>
                        </ElFormItem>
                        <ElFormItem label="收信人" prop="receiver">
                            <ElCheckboxGroup v-model="mail.receiver">
                                <ElCheckbox v-for="item in userList" :key="item.id"
                                             :label="item.email">
                                    {{item.nickname}}( {{item.email}} )
                                </ElCheckbox>
                            </ElCheckboxGroup>
                        </ElFormItem>
                    </ElForm>
                </ElCol>
            </ElRow>
        </Content>
    </Container>
</template>

<script setup lang="ts">
import { ref, onMounted } from 'vue'
import { ElNotification } from 'element-plus'
import type { FormInstance } from 'element-plus'
import userApi from "@/api/userApi"
import { dateFormatter } from "@/utility"
import mailApi from "@/api/mailApi"
import Container from "@/layout/Container.vue";
import Toolbar from "@/layout/Toolbar.vue";
import Content from "@/layout/Content.vue";

interface User {
    id: number
    email: string
    nickname: string
    register_time: string
    last_visit_time: string
}

interface MailForm {
    title: string
    content: string
    receiver: string[]
}

const codeModify = ref<FormInstance>()
const userList = ref<User[]>([])
const mail = ref<MailForm>({
    title: '',
    content: '',
    receiver: []
})

const rules = {
    title: {required: true, message: '请填写邮件标题', trigger: 'blur'},
    content: {required: true, message: '请填写邮件内容', trigger: 'blur'},
}

function clearForm() {
    mail.value = {
        title: '',
        content: '',
        receiver: []
    }
}

function getUserList() {
    const requestData = {
        pageNo: 1,
        pageSize: 500
    }
    userApi
        .list(requestData)
        .then(function(res) {
            userList.value = res.data.list.map(function(item) {
                item.register_time = dateFormatter(new Date(item.register_time))
                item.last_visit_time = dateFormatter(new Date(item.last_visit_time))
                return item
            })
        })
        .catch(function(err) {
            console.error(err)
        })
}

function submit() {
    if (!codeModify.value) return
    codeModify.value.validate(valid => {
        if (valid) {
            codeNewSubmit()
        } else {
            console.log('error submit!!')
            return false
        }
    })
}

function codeNewSubmit() {
    mailApi.sendEmail(mail.value)
        .then(function(res) {
            ElNotification({
                title: res.message,
                position: 'top-right',
                type: 'success',
                onClose: function() {}
            })
        })
}

onMounted(function() {
    getUserList()
})
</script>

<style scoped lang="scss">


</style>

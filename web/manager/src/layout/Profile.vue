<template>
    <ElDropdown ref="refDropDown" trigger="click" @command="menuClicked">
        <div class="ElDropdown-link">
            <ElAvatar
                :size="32"
                class="avatar"
                :src="`${getAuthorization()?.avatar}-thumbnail_200px`"
            />
        </div>
        <template #dropdown>
            <ElDropdown-menu>

                <ElDropdownItem command="changePassword">
                    <ElIcon><Key /></ElIcon>  修改密码
                </ElDropdownItem>
<!--                <ElDropdownItem command="changeProfile">-->
<!--                    <ElIcon><Postcard /></ElIcon>  编辑资料-->
<!--                </ElDropdownItem>-->
<!--                <ElDropdownItem divided command="setting">-->
<!--                    <ElIcon><Setting /></ElIcon> 设置-->
<!--                </ElDropdownItem>-->
                <ElDropdownItem divided disabled>
                    <ElIcon><UserFilled /></ElIcon>
                    {{ getAuthorization().nickname }}
                </ElDropdownItem>
                <ElDropdownItem disabled>
                    <ElIcon><PriceTag /></ElIcon>
                    {{ getAuthorization().role_name }}
                </ElDropdownItem>
                <ElDropdownItem command="logout" divided>
                    <ElIcon><SwitchButton /></ElIcon> 退出
                </ElDropdownItem>
            </ElDropdown-menu>
        </template>
    </ElDropdown>
</template>

<script setup lang="ts">
import {useRouter} from "vue-router";
import {ref} from "vue";
import {getAuthorization} from "@/utility";

const router = useRouter()
const refDropDown = ref()
function menuClicked(command: string){
    switch (command){
        case 'changePassword':
            router.push('/system/change-password')
            break;
        case 'changeProfile':
            router.push('/system/change-profile')
            break;
        case 'logout':
            router.push('/logout')
            break;
        default:break;
    }
}

</script>

<style scoped lang="scss">
@use "../assets/scss/variables" as *;
@use "../assets/scss/utility" as *;

.avatar{
    @extend .unselectable;
    cursor: pointer;
    &:hover{
        outline: 2px solid $cyan;
    }
}
</style>

<template>
    <ElDropdown ref="refDropDown" trigger="click" @command="menuClicked">
        <div class="dropdown-link">
            <div class="profile-brief">
                <span class="profile-brief__name">{{ auth?.nickname || '用户' }}</span>
                <span class="profile-brief__email">{{ auth?.email }}</span>
            </div>
            <ElAvatar
                :size="40"
                class="avatar"
                :src="avatarSrc"
            >
                <ElIcon><UserFilled /></ElIcon>
            </ElAvatar>
        </div>
        <template #dropdown>
            <ElDropdown-menu>
                <ElDropdownItem command="changePassword">
                    <ElIcon><Key /></ElIcon>  修改密码
                </ElDropdownItem>
                <ElDropdownItem divided disabled>
                    <ElIcon><UserFilled /></ElIcon>
                    {{ auth?.nickname }}
                </ElDropdownItem>
                <ElDropdownItem command="logout" divided>
                    <ElIcon><SwitchButton /></ElIcon> 退出
                </ElDropdownItem>
            </ElDropdown-menu>
        </template>
    </ElDropdown>
</template>

<script setup lang="ts">
import { computed, ref } from 'vue'
import { useRouter } from 'vue-router'
import { getAuthorization } from '@/utility'

const router = useRouter()
const refDropDown = ref()

const auth = computed(() => getAuthorization())

// 有缩略图后缀则直接用，否则拼 200px 缩略图
const avatarSrc = computed(() => {
    const url = auth.value?.avatar?.trim()
    if (!url) return ''
    if (url.includes('-thumbnail_')) return url
    return `${url}-thumbnail_200px`
})

function menuClicked(command: string) {
    switch (command) {
        case 'changePassword':
            router.push('/system/change-password')
            break
        case 'changeProfile':
            router.push('/system/change-profile')
            break
        case 'logout':
            router.push('/logout')
            break
        default:
            break
    }
}
</script>

<style scoped lang="scss">
@use '../assets/scss/plugin' as *;

.dropdown-link {
    display: flex;
    align-items: center;
    gap: 12px;
    cursor: pointer;
}

.profile-brief {
    display: flex;
    flex-direction: column;
    align-items: flex-end;
    line-height: 1.25;
    max-width: 200px;

    &__name {
        font-size: 14px;
        font-weight: 600;
        color: $text-main;
        overflow: hidden;
        text-overflow: ellipsis;
        white-space: nowrap;
        max-width: 100%;
    }

    &__email {
        font-size: 11px;
        color: $text-main;
        overflow: hidden;
        text-overflow: ellipsis;
        white-space: nowrap;
        max-width: 100%;
    }
}

.avatar {
    @extend .unselectable;
    flex-shrink: 0;
    cursor: pointer;

    &:hover {
        outline: 2px solid $cyan;
    }
}
</style>

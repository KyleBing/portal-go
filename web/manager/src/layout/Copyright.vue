<template>
    <div class="copyright" :style="`height: ${height}px`">
        <div class="info">
            <dl>
                <dt>用户名：</dt>
                <dd>{{ authorization.nickname }}
                    <span class="logout-text" @click="logout">退出</span>
                </dd>
            </dl>
            <dl><dt>版本：</dt><dd>v{{packageInfo.version}}</dd></dl>
        </div>
    </div>

</template>

<script setup lang="ts">
import packageInfo from '../../package.json'
import {getAuthorization} from "@/utility";
import {useRouter} from "vue-router";

defineProps( {
    height: {
        type: Number,
        default: 100
    }
})

const router = useRouter()

const authorization = getAuthorization()

function logout(){
    router.push('/logout')
}

</script>

<style lang="scss" scoped>
@use "../assets/scss/plugin" as *;

.copyright{
    overflow: hidden;
    display: flex;
    flex-flow: column nowrap;
    justify-content: flex-end;
    .info{
        flex-shrink: 0;
        border-top: 1px solid $theme-border-main;
        display: flex;
        flex-flow: column nowrap;
        justify-content: flex-end;
        padding: 10px 30px;
        dl{
            font-size: 0.7rem;
            line-height: 1.5;
            color: $theme-text-secondary;
            display: flex;
            justify-content: space-between;
            dd:hover {
                color: $theme-primary;
            }
        }
    }
}

.logout-text{
    cursor: pointer;
    color: $theme-text-muted;
    font-size: 0.7rem;
    &:hover{
        color: $red;
        text-decoration: underline;
    }
}

</style>

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
    height: { // 高度
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
@use "sass:color";
@use "../assets/scss/variables" as *;
.copyright{
    overflow: hidden;
    display: flex;
    flex-flow: column nowrap;
    justify-content: flex-end;
    .logo{
        flex-shrink: 1;
        img{
            display: block;
            max-width: 100%;
        }
    }
    .info{
        flex-shrink: 0;
        border-top: 1px solid color.adjust($blue, $alpha: -0.8);
        display: flex;
        flex-flow: column nowrap;
        justify-content: flex-end;
        padding: 10px 30px;
        dl{
            font-size: 0.7rem;
            line-height: 1.5;
            color: $text-subtitle;
            display: flex;
            justify-content: space-between;
            dd{
            }
        }
    }
}

a{
    &:hover{
        color: $blue;
        text-decoration: underline;
    }
}

.logout-text{
    cursor: pointer;
    color: $text-subtitle;
    &:hover{
        text-decoration: underline;
    }
}

</style>

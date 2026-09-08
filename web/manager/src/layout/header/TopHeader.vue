<template>
    <div class="header">
        <div class="header-left">
            <LogoAndTitle/>
        </div>
        <div class="header-center" v-if="!isRegisterMode">
            <TopNavbar :ticketTotal="systemStatisticInfo.ticket_total"/>
        </div>
        <div class="header-right">
            <Clock/>
            <Profile v-if="!isRegisterMode"/>
        </div>
    </div>
</template>

<script setup lang="ts">
import Profile from "../Profile.vue";
import {onMounted, onUnmounted, ref} from "vue";
import TopNavbar from "@/layout/header/TopNavbar.vue";
import Clock from "@/layout/header/Clock.vue";
import LogoAndTitle from "@/layout/header/LogoAndTitle.vue";

interface SystemStatisticInfo {
    device_total: number;
    alarm_total: number;
    duration: number;
    ticket_total: number;
}

const props = defineProps<{
    isRegisterMode?: boolean;
}>();

let timeoutHandle: number | null = null;

onMounted(() => {
    // Add any initialization code here
});

onUnmounted(() => {
    if (timeoutHandle) {
        clearTimeout(timeoutHandle);
        timeoutHandle = null;
    }
});

const systemStatisticInfo = ref<SystemStatisticInfo>({
    device_total: 0,
    alarm_total: 0,
    duration: 0,
    ticket_total: 0
});
</script>

<style lang="scss" scoped>
@use "../../assets/scss/variables" as *;
@use "sass:color";

.header {
    justify-content: space-between;
    padding: 0 20px;
    background: linear-gradient(to right, color.scale($blue, $lightness: -20%), color.scale($blue, $saturation: -20%));
    background: linear-gradient(to right, color.scale($color-main, $lightness: -20%), color.scale($color-main, $saturation: -20%));
    background-color: $blue;
    display: flex;
    font-size: 1.3rem;
    flex-flow: row nowrap;
    align-items: center;
    line-height: 60px;
    color: color.adjust(white, $alpha: -0.1);
}

.header-left {
    flex-flow: row nowrap;
    display: flex;
    justify-content: flex-start;
    align-items: center;
}

.header-center {
    flex-flow: row nowrap;
    display: flex;
    justify-content: flex-start;
    align-items: center;
}

.header-right {
    padding-right: 20px;
    flex-flow: row nowrap;
    display: flex;
    justify-content: flex-start;
    align-items: center;
}
</style>

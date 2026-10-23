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
@use 'sass:color';
@use '../../assets/scss/plugin' as *;

.header {
    justify-content: space-between;
    padding: 0 20px;
    background: linear-gradient(to right, $theme-bg-header-start, $theme-bg-header-end);
    display: flex;
    font-size: 1.3rem;
    flex-flow: row nowrap;
    align-items: center;
    line-height: 60px;
    color: color.adjust(white, $alpha: -0.1);
    border-bottom: 1px solid $theme-border-light;
}

.header-left,
.header-center,
.header-right {
    display: flex;
    flex-flow: row nowrap;
    align-items: center;
}

.header-left {
    justify-content: flex-start;
}

.header-center {
    justify-content: flex-start;
}

.header-right {
    padding-right: 20px;
    gap: 5px;
}
</style>

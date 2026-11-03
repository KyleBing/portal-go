<template>
    <div class="header">
        <div class="header-left">
            <LogoAndTitle/>
            <TopNavbar class="ml-5" v-if="!isRegisterMode" :ticketTotal="systemStatisticInfo.ticket_total"/>
        </div>
        <div class="header-right">
            <Profile v-if="!isRegisterMode"/>
        </div>
    </div>
</template>

<script setup lang="ts">
import Profile from "../Profile.vue";
import {onMounted, onUnmounted, ref} from "vue";
import TopNavbar from "@/layout/header/TopNavbar.vue";
import LogoAndTitle from "@/layout/header/LogoAndTitle.vue";

interface SystemStatisticInfo {
    device_total: number;
    alarm_total: number;
    duration: number;
    ticket_total: number;
}

defineProps<{
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
    justify-content: flex-start;
    gap: 12px;
    height: 100%;
    padding: 0 20px;
    background: linear-gradient(to bottom, $theme-bg-header-start, $theme-bg-header-end);
    display: flex;
    font-size: 1.3rem;
    flex-flow: row nowrap;
    align-items: center;
    line-height: 60px;
    color: $text-main;
    border-bottom: 1px solid $theme-border-card;
}

.header-left,
.header-right {
    display: flex;
    flex-flow: row nowrap;
    align-items: center;
}

.header-left {
    flex: 1;
    min-width: 0;
    justify-content: flex-start;
    gap: 8px;
}

.header-right {
    flex-shrink: 0;
    margin-left: auto;
    gap: 5px;
}
</style>

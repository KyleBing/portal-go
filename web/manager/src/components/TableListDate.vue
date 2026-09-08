
<template>
    <div class="time-stage-item">
        <div class="time-stage-label" v-if="dates.length > 1">
            <el-text size="small" type="">{{names[0] || '时间'}}</el-text>
        </div>
        <div class="time-stage-value">
            <span v-if="dates[0]">{{dateFormatter(new Date(dates[0]))}}</span>
            <span v-else>-</span>
        </div>
    </div>
    <div class="time-stage-item" v-if="dates.length > 1">
        <div class="time-stage-label">
            <el-text size="small" type="warning">{{names[1] || '更新'}}</el-text>
        </div>
        <div class="time-stage-value">
            <span v-if="dates[1]">{{dateFormatter(new Date(dates[1]))}}</span>
            <span v-else>-</span>
        </div>
    </div>
</template>

<script setup lang="ts">

withDefaults(defineProps<{
    names?: Array<string>,
    dates?: Array<string>
}>(), {
    names: [],
    dates: []
})

import {dateFormatter} from "@/utility";
</script>


<style scoped lang="scss">
.time-stage-item{
    display: flex;
    font-size: 12px;
    line-height: 1.4;
    .time-stage-label{
        flex-shrink: 0;
    }
    .time-stage-value{
        margin-left: 5px;
    }
}
</style>

<template>
    <div class="clock">
        <div class="date">{{clock.date}}</div>
        <div class="time">{{clock.time}}</div>
    </div>
</template>

<script setup lang="ts">
import {onMounted, onUnmounted, ref} from "vue";
import {dateFormatter} from "@/utility.ts";

const clock = ref({
    date: '',
    time: ''
})

let clockIntervalHandle: number | null = null

onMounted(()=>{
    // clock start
    // init
    clock.value = {
        date: dateFormatter(new Date(), 'yyyy/MM/dd'),
        time: dateFormatter(new Date(), 'HH:mm:ss')
    }
    clockIntervalHandle = setInterval(() => {
        clock.value = {
            date: dateFormatter(new Date(), 'yyyy/MM/dd'),
            time: dateFormatter(new Date(), 'HH:mm:ss')
        }
    }, 1000)
})

onUnmounted(()=>{
    if (clockIntervalHandle){
        clearInterval(clockIntervalHandle!) // 清除定时器
    }
})
</script>

<style scoped lang="scss">
@use "../../assets/scss/variables" as *;
@use "../../assets/scss/utility" as *;
@use "../../assets/scss/font" as *;

.clock{
    display: flex;
    flex-flow: column nowrap;
    align-items: flex-end;
    justify-content: center;
    white-space: nowrap;
    margin-right: 20px;
    line-height: 1.3;
    .date{
        font-size: 14px;
    }
    .time{
        font-size: 14px;
    }
}

@media screen and (max-width: 1500px) {
    .clock{
        //display: none;
    }
}
</style>

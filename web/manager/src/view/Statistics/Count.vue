<template>
    <div :style="`color: ${color}`" class="count" ref="countUp">0</div>
</template>

<script setup lang="ts">
import { ref, onMounted } from 'vue';
import { CountUp } from "countup.js";

interface Props {
    endVal?: number;
    color?: string;
}

const props = withDefaults(defineProps<Props>(), {
    endVal: 100,
    color: 'black'
});

const countUp = ref<HTMLElement | null>(null);

onMounted(() => {
    setTimeout(() => {
        if (countUp.value) {
            const countUpInstance = new CountUp(countUp.value, props.endVal, {});
            if (!countUpInstance.error) {
                countUpInstance.start();
            } else {
                console.error(countUpInstance.error);
            }
        }
    }, 100);
});
</script>

<style scoped lang="scss">
.count{
    text-align: center;
    line-height: 1;
    font-size: 60px;
    color: white;
    font-family: 'Digit', sans-serif;
}

</style>

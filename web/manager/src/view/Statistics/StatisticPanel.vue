<template>
    <div class="statistic-panel-container">
        <div class="statistic-panel">
            <h3 v-if="title">{{title}}</h3>
            <slot/>
        </div>
    </div>
</template>

<script setup lang="ts">
interface Props {
    title?: string;
}

withDefaults(defineProps<Props>(), {
    title: ''
});
</script>

<style scoped lang="scss">
@use "../../assets/scss/variables" as *;
@use "../../assets/scss/utility" as *;
@use "../../assets/scss/font" as *;
@use "sass:color";

.statistic-panel-container {
    padding: 10px;
}

.statistic-panel {
    flex-shrink: 0;
    background-color: $bg-light;
    @include border-radius($radius-pc);
    padding: 20px 30px;
}

h3 {
    color: $text-title;
    font-size: $fz-big;
    line-height: 1;
    padding-bottom: 15px;
}
</style>


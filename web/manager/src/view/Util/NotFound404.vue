<template>
<div class="page-404" :style="`height: ${projectStore.windowInsets.height}px`">
    <p  class="template">
        404 <br>
        Page not found...
    </p>
</div>

</template>
<script setup lang="ts">
    import {useProjectStore} from "@/pinia";

    const projectStore = useProjectStore()
</script>

<style scoped lang="scss">
@use "../../assets/scss/variables" as *;
@use "../../assets/scss/utility" as *;
@use "../../assets/scss/font" as *;

.page-404{
    display: flex;
    justify-content: center;
    align-items: center;
    height: 100%;
    width: 100%;
    backdrop-filter: blur(3px);
}
.template{
    font-weight: bold;
    font-size: $fz-big;
    font-family: "JetBrains Mono", sans-serif;
    color: $color-main;
}
</style>

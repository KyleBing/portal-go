<template>
    <ElContainer>
        <ElHeader>
            <TopHeader/>
        </ElHeader>

        <ElContainer>
            <Aside class="animate__animated"/>
            <ElContainer>
                <ElMain>
                    <router-view/>
                </ElMain>
            </ElContainer>
        </ElContainer>
    </ElContainer>
</template>

<script setup lang="ts">
import TopHeader from "./header/TopHeader.vue";
import Aside from "./Aside.vue";
</script>

<style scoped lang="scss">
@use "../assets/scss/variables" as *;
@use "../assets/scss/utility" as *;
@use "../assets/scss/font" as *;

.el-header {
    padding: 0;
}
</style>

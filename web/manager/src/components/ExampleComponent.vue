<template>
    <div class="example-component">
        <button class="primary-button">Click me</button>
        <div class="flex-container">
            <div class="item">Item 1</div>
            <div class="item">Item 2</div>
        </div>
    </div>
</template>

<script setup lang="ts">
// Component logic here
</script>

<style lang="scss" scoped>
@use "@/assets/scss/variables" as *;
@use "@/assets/scss/utility" as *;
@use "@/assets/scss/font" as *;

.example-component {
    @include flex(column, center, center);
    padding: 2rem;

    .primary-button {
        @include button(#007bff, white);
        margin-bottom: 1rem;
    }

    .flex-container {
        @include flex(row, space-between, center);
        width: 100%;
        max-width: 800px;

        .item {
            padding: 1rem;
            background: #f0f0f0;
        }

        @include respond-to('medium') {
            @include flex(column, center, center);
            gap: 1rem;
        }
    }
}
</style> 
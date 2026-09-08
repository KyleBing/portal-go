<template>
    <div class="wubi-category-list">
        <div @click="$emit('click', '')"
             :class="['wubi-category-list-item', {active: activeCategoryId === ''  }]" >
            <div class="title">全部</div>
        </div>
        <div @click="$emit('click', item.id)"
             :class="['wubi-category-list-item', {active: item.id ===  activeCategoryId }]"
             v-for="item in categories" :key="item.id"
        >
            <div class="title">{{ item.name }}</div>
            <div class="count">{{ item.count }}</div>
        </div>
    </div>
</template>

<script setup lang="ts">
interface Category {
    id: number;
    name: string;
    count: number;
}

defineProps<{
    categories: Category[];
    activeCategoryId: string | number;
}>();

defineEmits<{
    (e: 'click', id: string | number): void;
}>();
</script>

<style scoped lang="scss">
@use "../../assets/scss/variables" as *;
@use "../../assets/scss/utility" as *;
@use "../../assets/scss/font" as *;

.wubi-category-list {
    border-right: 1px solid $border-color;
    overflow-y: auto;
    padding: 15px;
    background-color: white;
    @include border-radius($radius-mobile);
}

.wubi-category-list-item {
    text-align: center;
    padding: 3px 10px;
    font-size: $fz-main;
    color: $text-main;
    display: flex;
    cursor: pointer;
    justify-content: space-between;
    @include border-radius($radius-mobile);
    @extend .unselectable;

    .title {
        // Empty block for future styles
    }

    .count {
        font-size: $fz-main;
        color: $text-subtitle;
    }

    &.active {
        background-color: $bg-active;
    }

    &:hover {
        background-color: $bg-hover;
    }
}
</style>

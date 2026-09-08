<template>
    <div class="file-selector">
        <input type="file" @change="fileHasChanged"></input>
        <div class="file-info" v-if="currentFile">
            <div>{{currentFile.name}}</div>
            <div>{{currentFile.size}}</div>
        </div>
    </div>

</template>

<script setup lang="ts">
import { ref } from 'vue';

interface FileInfo {
    name: string;
    size: number;
}

const currentFile = ref<FileInfo | null>(null);

const emit = defineEmits<{
    (e: 'change', file: FileInfo): void;
}>();

function fileHasChanged(event: Event) {
    const target = event.target as HTMLInputElement;
    if (target.files && target.files[0]) {
        const file = target.files[0];
        currentFile.value = {
            name: file.name,
            size: file.size
        };
        emit('change', currentFile.value);
    }
}
</script>

<style lang="scss" scoped>
@use "sass:color";

.file-info {
    padding: 10px;
    border-radius: 5px;
    background-color: color.adjust(white, $alpha: -0.5);
    display: flex;
    justify-content: flex-start;
}
</style>

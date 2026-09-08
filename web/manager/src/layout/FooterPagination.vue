<template>
    <div class="page-footer">
        <ElPagination
            :total="pagerOption.total"
            :current-page="pagerOption.pageNo"
            :page-size="pagerOption.pageSize"
            :page-sizes="[15, 20, 50, 100, 200, 500]"
            layout="total, sizes, prev, pager, next, jumper"
            @size-change="sizeChange"
            @current-change="pagerChange"
        />
    </div>
</template>

<script setup lang="ts">
const emit = defineEmits(['pagerChange', 'sizeChange'])

withDefaults(defineProps<{
    pagerOption: {
        total: number,
        pageNo: number,
        pageSize: number
    }
}>(), {
    pagerOption: {
        total: 0,
        pageNo: 1,
        pageSize: 20
    }
})

function sizeChange(size: number){
    emit('sizeChange', size)
}
function pagerChange(pageNo: number){
    emit('pagerChange', pageNo)
}


</script>

<style scoped lang="scss">
@use "@/assets/scss/variables" as *;
.page-footer{
    font-size: $fz-small;
    padding: 0 10px;
    align-items: center;
    height: 50px;
    display: flex;
    justify-content: flex-end;
}
</style>

<template>
    <div ref="BarDom" class="charts" :style="`height: 250px; width: ${width}`"></div>
</template>

<script setup lang="ts">
import * as echarts from 'echarts'
import { ref, watch, onMounted, onBeforeUnmount, computed } from 'vue'

const props = withDefaults(defineProps<{
    data?: Array<{ value: number, name: string }>,
    title?: string,
    subTitle?: string
}>(), {
    data: () => [],
    title: '',
    subTitle: ''
})

const BarDom = ref<HTMLElement | null>(null)
const chart = ref<echarts.ECharts | null>(null)
const option = ref<any>(null)
const width = ref('500px')

const xAxisData = computed(() => props.data)

watch(() => props.data, (newValue) => {
    if (newValue) {
        resetData(newValue)
    }
})

onMounted(() => {
    if (window.innerWidth < 400) {
        width.value = '100%'
    }
    initChart()
    resetData(props.data)
})

onBeforeUnmount(() => {
    window.onresize = null
})

function resetData(newValue: Array<{ value: number, name: string }>) {
    if (option.value && chart.value) {
        let xAxisData = []
        let seriesData = []
        newValue.forEach(item => {
            seriesData.push(item.value)
            xAxisData.push(item.name)
        })
        option.value.xAxis[0].data = xAxisData
        option.value.series[0].data = seriesData
        chart.value.setOption(option.value)
    }
}

function initChart() {
    option.value = {
        color: [
            '#FFA41C',
            '#2F3037',
            '#EE6666',
            '#FFDC60',
            '#5C7BD9',
            '#7ED3F4',
            '#c7c7c7',
            '#9FE080'
        ],
        grid: {
            bottom: 50,
            right: 30
        },
        title: {
            text: '',
            left: 'center'
        },
        tooltip: {
            trigger: 'axis',
            axisPointer: {
                type: 'shadow'
            }
        },
        xAxis: [{
            type: 'category',
            data: [],
            axisLabel: {
                fontSize: 11,
                rotate: -45
            }
        }],
        yAxis: [{
            type: 'value'
        }],
        series: [{
            name: '到访人员类型',
            type: 'bar',
            data: [],
            label: {
                show: true,
                position: 'top',
                fontSize: 12,
            },
        }]
    }

    option.value.title.text = props.title
    if (BarDom.value) {
        chart.value = echarts.init(BarDom.value)
        chart.value.setOption(option.value)
        option.value.series[0].name = props.title
    }
}

function resize() {
    if (chart.value) {
        chart.value.resize()
    }
}

defineExpose({
    resize
})
</script>

<style scoped>

</style>

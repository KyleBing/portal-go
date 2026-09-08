<template>
    <div ref="BarDom" class="charts" :style="`height: 350px; width: ${width}`"></div>
</template>

<script setup lang="ts">
import { ref, watch, onMounted, onBeforeUnmount, computed } from 'vue'
import * as echarts from 'echarts/core'
import { PieChart } from 'echarts/charts'
import {
    TooltipComponent,
    TitleComponent,
} from 'echarts/components'
import { CanvasRenderer } from 'echarts/renderers'
import type { EChartsType } from 'echarts/core'

echarts.use([PieChart, TooltipComponent, TitleComponent, CanvasRenderer])

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
const chart = ref<EChartsType | null>(null)
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
    chart.value?.dispose()
    chart.value = null
})

function resetData(newValue: Array<{ value: number, name: string }>) {
    if (option.value && chart.value) {
        option.value.series[0].data = newValue
        chart.value.setOption(option.value)
    }
}

function initChart() {
    option.value = {
        color: [
            '#FFA41C',
            '#2F3037',
            '#9FE080',
            '#5C7BD9',
            '#7ED3F4',
            '#EE6666',
            '#c7c7c7',
            '#FFDC60'
        ],
        grid: {
            top: 20,
            bottom: 20,
            left: 20,
            right: 20
        },
        title: {
            text: '',
            left: 'center',
        },
        tooltip: {
            trigger: 'item'
        },
        series: [
            {
                name: '饼状图',
                type: 'pie',
                radius: '60%',
                data: [],
                label: {
                    show: true,
                    position: 'outside',
                    fontSize: 12,
                    formatter: '{b} {d}%'
                },
                emphasis: {
                    itemStyle: {
                        shadowBlur: 10,
                        shadowOffsetX: 0,
                        shadowColor: 'rgba(0, 0, 0, 0.2)'
                    }
                }
            }
        ]
    }

    if (window.innerWidth < 400) {
        option.value.series[0].radius = '40%'
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

<style lang="scss" scoped>
</style>

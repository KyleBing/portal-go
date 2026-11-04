export interface CpuConfig {
    fire: { rate: number; core: number }
    ice: { rate: number; core: number }
}

export interface GpuConfig {
    brand: string
    core: number
    info: string
}

export interface Chip {
    id: string
    series: string
    name: string
    model: string
    tech: string
    techCompany: string
    dieSize: string
    isa: string
    cpu: CpuConfig[]
    gpu: GpuConfig[]
    ai: { core: string; rate: string }
    release: string
    devices: string[]
    os: { init: string; latest: string }
    transistorCount?: string
}

function asText(value: unknown): string {
    if (value === null || value === undefined) return ''
    return String(value)
}

function asNumber(value: unknown): number {
    if (typeof value === 'number' && Number.isFinite(value)) return value
    const n = parseFloat(String(value ?? '').replace(/[^\d.]/g, ''))
    return Number.isFinite(n) ? n : 0
}

function asCpu(raw: any): CpuConfig {
    const fire = raw?.fire || {}
    const ice = raw?.ice || {}
    return {
        fire: { rate: asNumber(fire.rate), core: asNumber(fire.core) },
        ice: { rate: asNumber(ice.rate), core: asNumber(ice.core) }
    }
}

function asGpu(raw: any): GpuConfig {
    return {
        brand: asText(raw?.brand),
        core: asNumber(raw?.core),
        info: asText(raw?.info)
    }
}

export function blankChip(series = 'A'): Chip {
    return {
        id: '',
        series,
        name: '',
        model: '',
        tech: '',
        techCompany: '',
        dieSize: '',
        isa: '',
        cpu: [{ fire: { rate: 0, core: 0 }, ice: { rate: 0, core: 0 } }],
        gpu: [{ brand: 'Apple', core: 0, info: '' }],
        ai: { core: '', rate: '' },
        release: '',
        devices: [],
        os: { init: '', latest: '' },
        transistorCount: ''
    }
}

// 目录表扁平字段 + 嵌套 source 合成编辑表单。
export function rowToChip(row: any): Chip {
    const src = row?.source && typeof row.source === 'object' ? row.source : {}
    const cpuList = Array.isArray(src.cpu) ? src.cpu.map(asCpu) : []
    const gpuList = Array.isArray(src.gpu) ? src.gpu.map(asGpu) : []
    const fallbackCpu = [{
        fire: { rate: asNumber(row?.fireRate), core: asNumber(row?.fireCore) },
        ice: { rate: asNumber(row?.iceRate), core: asNumber(row?.iceCore) }
    }]
    const fallbackGpu = [{
        brand: asText(src.gpu?.[0]?.brand || 'Apple'),
        core: asNumber(row?.gpuCore),
        info: asText(row?.gpuInfo)
    }]
    return {
        id: asText(row?.id),
        series: asText(row?.series || 'A'),
        name: asText(row?.name || src.name),
        model: asText(row?.model || src.model),
        tech: asText(row?.tech ?? src.tech),
        techCompany: asText(src.techCompany),
        dieSize: asText(row?.die ?? src.dieSize),
        isa: asText(row?.isa || src.isa),
        cpu: cpuList.length ? cpuList : fallbackCpu,
        gpu: gpuList.length ? gpuList : fallbackGpu,
        ai: {
            core: asText(row?.aiCore ?? src.ai?.core),
            rate: asText(row?.aiRate ?? src.ai?.rate)
        },
        release: asText(row?.release || src.release),
        devices: Array.isArray(row?.devices) ? row.devices.map(asText).filter(Boolean)
            : Array.isArray(src.devices) ? src.devices.map(asText).filter(Boolean) : [],
        os: {
            init: asText(src.os?.init),
            latest: asText(src.os?.latest)
        },
        transistorCount: asText(row?.transistor || src.transistorCount)
    }
}

// 保存时写回扁平展示字段，嵌套 CPU/GPU 放进 source。
export function chipToPayload(chip: Chip) {
    const cpu = chip.cpu[0] || { fire: { rate: 0, core: 0 }, ice: { rate: 0, core: 0 } }
    const gpuCores = chip.gpu.reduce((sum, item) => sum + (item.core || 0), 0)
    const gpu0 = chip.gpu[0] || { brand: '', core: 0, info: '' }
    const devices = (chip.devices || []).map(item => item.trim()).filter(Boolean)
    return {
        id: chip.id,
        series: chip.series,
        name: chip.name.trim(),
        model: chip.model.trim(),
        tech: asNumber(chip.tech),
        die: asNumber(chip.dieSize),
        isa: chip.isa.trim(),
        fireCore: cpu.fire.core,
        iceCore: cpu.ice.core,
        fireRate: cpu.fire.rate,
        iceRate: cpu.ice.rate,
        gpuCore: gpuCores,
        gpuInfo: gpu0.info || (gpuCores ? `${gpuCores} 核` : ''),
        aiCore: asNumber(chip.ai.core) || chip.ai.core,
        aiRate: chip.ai.rate,
        transistor: asNumber(chip.transistorCount),
        release: chip.release.trim(),
        year: asNumber(String(chip.release).slice(0, 4)),
        devices,
        source: {
            name: chip.name.trim(),
            model: chip.model.trim(),
            tech: chip.tech,
            techCompany: chip.techCompany.trim(),
            dieSize: chip.dieSize,
            isa: chip.isa.trim(),
            cpu: chip.cpu,
            gpu: chip.gpu,
            ai: { core: chip.ai.core, rate: chip.ai.rate },
            release: chip.release.trim(),
            devices,
            os: { init: chip.os.init, latest: chip.os.latest },
            transistorCount: chip.transistorCount
        }
    }
}

export function fireCores(chip: Chip) {
    return chip.cpu.reduce((sum, item) => sum + (item.fire?.core || 0), 0)
}

export function iceCores(chip: Chip) {
    return chip.cpu.reduce((sum, item) => sum + (item.ice?.core || 0), 0)
}

export function maxRate(chip: Chip) {
    const rates = chip.cpu.flatMap(item => [item.fire?.rate || 0, item.ice?.rate || 0])
    return rates.length ? Math.max(...rates) : 0
}

export function gpuCores(chip: Chip) {
    return chip.gpu.reduce((sum, item) => sum + (item.core || 0), 0)
}

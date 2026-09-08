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


/** 心形上浮 canvas 背景（对齐 portal-wubi-go 登录背景画法，粒子改回心形） */

const MAX_DPR = 2
const FRAME_INTERVAL = 1000 / 60

function randomInt(min: number, max: number): number {
    return Math.round(Math.random() * (max - min) + min)
}

function randomFloat(min: number, max: number): number {
    return Math.random() * (max - min) + min
}

function clamp(value: number, min: number, max: number): number {
    return value < min ? min : value > max ? max : value
}

type HeartParticle = {
    /** 预生成的颜色（不含透明度），透明度交给 globalAlpha */
    color: string
    size: number
    x: number
    y: number
    /** 上浮速度（px/s），与帧率无关 */
    speed: number
    opacityMax: number
    age: number
    fadeIn: number
    fadeDelay: number
}

export class AnimateHeartCanvas {
    isPlaying = true
    private destroyed = false
    private config = {
        count: 180,
        sizeMin: 28,
        sizeMax: 72,
        colorSaturate: 70,
        colorLight: 62,
        hMin: 0,
        hMax: 360,
        minOpacity: 28,
        maxOpacity: 85,
        speedMin: 6,
        speedMax: 22,
    }
    private bgColor: string | undefined
    private particles: HeartParticle[] = []
    private canvas: HTMLCanvasElement | null = null
    private ctx: CanvasRenderingContext2D | null = null
    private viewWidth = 0
    private viewHeight = 0
    private dpr = 1
    private rafId = 0
    private lastTime = 0
    private resizeTimer: ReturnType<typeof setTimeout> | null = null
    private prefersReducedMotion = false
    private readonly canvasId = 'heartLayer'

    /**
     * @param hMin 色相最小值
     * @param hMax 色相最大值
     * @param count 心数量
     * @param sizeMin 心形最小值
     * @param sizeMax 心形最大值
     * @param minOpacity 心最大透明度的下限（百分比 0~100）
     * @param maxOpacity 心最大透明度的上限（百分比 0~100）
     * @param bgColor 背景色（可选；由 CSS 绘制）
     */
    constructor(
        hMin = 330,
        hMax = 350,
        count = 180,
        sizeMin = 28,
        sizeMax = 72,
        minOpacity = 28,
        maxOpacity = 85,
        bgColor?: string,
    ) {
        this.config.hMin = Number.isNaN(hMin) ? 330 : hMin
        this.config.hMax = Number.isNaN(hMax) ? 350 : hMax
        this.config.count = count
        this.config.sizeMin = Number.isNaN(sizeMin) ? 28 : sizeMin
        this.config.sizeMax = Number.isNaN(sizeMax) ? 72 : sizeMax
        const opacityMin = clamp(Number.isNaN(minOpacity) ? 28 : minOpacity, 0, 100)
        const opacityMax = clamp(Number.isNaN(maxOpacity) ? 85 : maxOpacity, 0, 100)
        this.config.minOpacity = Math.min(opacityMin, opacityMax)
        this.config.maxOpacity = Math.max(opacityMin, opacityMax)
        this.bgColor = bgColor
        this.init()
    }

    destroy() {
        this.destroyed = true
        this.isPlaying = false
        if (this.rafId) cancelAnimationFrame(this.rafId)
        this.rafId = 0
        if (this.resizeTimer) clearTimeout(this.resizeTimer)
        this.resizeTimer = null
        window.removeEventListener('resize', this.onResize)
        document.removeEventListener('visibilitychange', this.onVisibilityChange)
        this.canvas?.remove()
        this.canvas = null
        this.ctx = null
        this.particles = []
    }

    private init() {
        this.prefersReducedMotion = typeof window.matchMedia === 'function'
            && window.matchMedia('(prefers-reduced-motion: reduce)').matches

        const layer = document.createElement('canvas')
        layer.id = this.canvasId
        this.canvas = layer
        this.ctx = layer.getContext('2d')

        layer.style.position = 'fixed'
        layer.style.top = '0'
        layer.style.left = '0'
        layer.style.zIndex = '0'
        layer.style.pointerEvents = 'none'
        layer.style.userSelect = 'none'
        if (this.bgColor) {
            layer.style.background = `radial-gradient(90% 70% at 50% 32%, rgba(255,255,255,0.06) 0%, rgba(255,255,255,0) 62%), ${this.bgColor}`
        }

        this.layout()
        document.documentElement.append(layer)

        this.particles = []
        for (let i = 0; i < this.config.count; i++) {
            this.particles.push(this.buildParticle(true))
        }

        window.addEventListener('resize', this.onResize)
        document.addEventListener('visibilitychange', this.onVisibilityChange)

        if (this.prefersReducedMotion) {
            this.config.speedMin *= 0.35
            this.config.speedMax *= 0.35
            for (const particle of this.particles) {
                particle.speed *= 0.35
                particle.age = particle.fadeDelay + particle.fadeIn
            }
        }

        this.rafId = requestAnimationFrame(this.loop)
    }

    private layout() {
        const layer = this.canvas
        if (!layer) return

        const prevWidth = this.viewWidth
        const prevHeight = this.viewHeight
        this.viewWidth = window.innerWidth
        this.viewHeight = window.innerHeight
        this.dpr = Math.min(window.devicePixelRatio || 1, MAX_DPR)

        layer.width = Math.round(this.viewWidth * this.dpr)
        layer.height = Math.round(this.viewHeight * this.dpr)
        layer.style.width = `${this.viewWidth}px`
        layer.style.height = `${this.viewHeight}px`

        if (prevWidth > 0 && prevHeight > 0) {
            const scaleX = this.viewWidth / prevWidth
            const scaleY = this.viewHeight / prevHeight
            for (const particle of this.particles) {
                particle.x *= scaleX
                particle.y *= scaleY
            }
        }

        this.ctx?.setTransform(this.dpr, 0, 0, this.dpr, 0, 0)

        if (!this.isPlaying && !this.destroyed) this.render(0)
    }

    private onResize = () => {
        if (this.resizeTimer) clearTimeout(this.resizeTimer)
        this.resizeTimer = setTimeout(() => {
            this.resizeTimer = null
            this.layout()
        }, 150)
    }

    private onVisibilityChange = () => {
        if (document.hidden) {
            this.pause()
        } else {
            this.resume()
        }
    }

    private pause() {
        this.isPlaying = false
        if (this.rafId) cancelAnimationFrame(this.rafId)
        this.rafId = 0
        this.lastTime = 0
    }

    private resume() {
        if (this.destroyed || this.isPlaying) return
        this.isPlaying = true
        this.lastTime = 0
        this.rafId = requestAnimationFrame(this.loop)
    }

    private loop = (timestamp: number) => {
        if (this.destroyed) return
        this.rafId = requestAnimationFrame(this.loop)

        if (this.lastTime && timestamp - this.lastTime < FRAME_INTERVAL - 1) return
        const dt = this.lastTime
            ? Math.min(timestamp - this.lastTime, 120) / 1000
            : FRAME_INTERVAL / 1000
        this.lastTime = timestamp
        this.render(dt)
    }

    private randomSize(): number {
        const t = Math.pow(Math.random(), 1.8)
        return Math.round(this.config.sizeMin + (this.config.sizeMax - this.config.sizeMin) * t)
    }

    private buildParticle(initial: boolean): HeartParticle {
        const size = this.randomSize()
        const colorH = randomInt(this.config.hMin, this.config.hMax)
        const speedScale = Math.max(this.viewHeight / 900, 0.6)
        return {
            size,
            color: `hsl(${colorH} ${this.config.colorSaturate}% ${this.config.colorLight}%)`,
            x: randomFloat(0, this.viewWidth),
            y: initial
                ? randomFloat(-size, this.viewHeight)
                : this.viewHeight + size + randomFloat(0, this.viewHeight * 0.25),
            speed: randomFloat(this.config.speedMin, this.config.speedMax) * speedScale,
            opacityMax: randomInt(this.config.minOpacity, this.config.maxOpacity) / 100,
            age: 0,
            fadeIn: randomFloat(0.9, 2.2),
            fadeDelay: randomFloat(0, initial ? 1.2 : 0.5),
        }
    }

    private respawn(particle: HeartParticle) {
        const size = this.randomSize()
        const colorH = randomInt(this.config.hMin, this.config.hMax)
        const speedScale = Math.max(this.viewHeight / 900, 0.6)
        particle.size = size
        particle.color = `hsl(${colorH} ${this.config.colorSaturate}% ${this.config.colorLight}%)`
        particle.x = randomFloat(0, this.viewWidth)
        particle.y = this.viewHeight + size + randomFloat(0, this.viewHeight * 0.25)
        particle.speed = randomFloat(this.config.speedMin, this.config.speedMax) * speedScale
        particle.opacityMax = randomInt(this.config.minOpacity, this.config.maxOpacity) / 100
        particle.age = 0
        particle.fadeIn = randomFloat(0.9, 2.2)
        particle.fadeDelay = randomFloat(0, 0.5)
    }

    /** 与 animate-heart-canvas 相同的贝塞尔心形，(x,y) 为心尖附近参考点 */
    private drawHeart(ctx: CanvasRenderingContext2D, x: number, y: number, size: number) {
        const width = size
        const height = size
        const topCurveHeight = height * 0.3
        ctx.beginPath()
        ctx.moveTo(x, y + topCurveHeight)
        ctx.bezierCurveTo(x, y, x - width / 2, y, x - width / 2, y + topCurveHeight)
        ctx.bezierCurveTo(
            x - width / 2, y + (height + topCurveHeight) / 2,
            x, y + (height + topCurveHeight) / 1.4,
            x, y + height,
        )
        ctx.bezierCurveTo(
            x, y + (height + topCurveHeight) / 1.8,
            x + width / 2, y + (height + topCurveHeight) / 2,
            x + width / 2, y + topCurveHeight,
        )
        ctx.bezierCurveTo(x + width / 2, y, x, y, x, y + topCurveHeight)
        ctx.closePath()
        ctx.fill()
    }

    private render(dt: number) {
        const ctx = this.ctx
        if (!ctx) return

        const height = this.viewHeight
        ctx.clearRect(0, 0, this.viewWidth, height)

        const fadeTop = height * 0.1

        for (const particle of this.particles) {
            particle.age += dt
            particle.y -= particle.speed * dt

            if (particle.y < -particle.size) {
                this.respawn(particle)
                continue
            }
            if (particle.y > height + particle.size) continue

            let alpha = clamp((particle.age - particle.fadeDelay) / particle.fadeIn, 0, 1) * particle.opacityMax
            if (particle.y < fadeTop) alpha *= clamp(particle.y / fadeTop, 0, 1)
            if (alpha <= 0.003) continue

            ctx.globalAlpha = alpha
            ctx.fillStyle = particle.color
            this.drawHeart(ctx, particle.x, particle.y, particle.size)
        }

        ctx.globalAlpha = 1
    }
}

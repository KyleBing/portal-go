import type { PayPlatform } from '@/api/registrationApplicationApi'
import type { RegistrationPayQrItem } from '@/api/registrationPayQrApi'

import qrAlipay from '@/assets/qr_payment/qr_alipay.png'
import qrWepay from '@/assets/qr_payment/qr_wepay.png'

export function isRegistrationPayAmount(value: number): boolean {
    return Number.isInteger(value) && value >= 0 && value <= 100000
}

export function formatRegistrationPayAmount(amountYuan: number | null | undefined): string {
    if (amountYuan === null || amountYuan === undefined) return '—'
    if (amountYuan === 0) return '自定义'
    return `¥${amountYuan}`
}

// 配置里可能是完整 URL，或相对 /portal 的路径
export function resolvePayQrUrl(raw: string | undefined | null): string {
    const s = String(raw || '').trim()
    if (!s) return ''
    if (/^https?:\/\//i.test(s) || s.startsWith('data:') || s.startsWith('blob:')) return s
    if (s.startsWith('/portal/')) return s
    if (s.startsWith('/')) return `/portal${s}`
    return `/portal/${s.replace(/^\/+/, '')}`
}

export function payQrImage(platform: PayPlatform, configuredUrl?: string | null): string {
    const resolved = resolvePayQrUrl(configuredUrl)
    if (resolved) return resolved
    return platform === 'wechat' ? qrWepay : qrAlipay
}

export function parsePayQrList(raw: unknown): RegistrationPayQrItem[] {
    if (!Array.isArray(raw)) return []
    return raw
        .map((item) => {
            if (!item || typeof item !== 'object') return null
            const row = item as Record<string, unknown>
            const amount = Number(row.amount_yuan)
            if (!isRegistrationPayAmount(amount)) return null
            return {
                id: Number(row.id) || 0,
                amount_yuan: amount,
                alipay_url: String(row.alipay_url || '').trim(),
                alipay_note: String(row.alipay_note || '').trim(),
                wechat_url: String(row.wechat_url || '').trim(),
                wechat_note: String(row.wechat_note || '').trim(),
            }
        })
        .filter((item): item is RegistrationPayQrItem => !!item)
}

export function pickPayQrItem(list: RegistrationPayQrItem[], amountYuan: number): RegistrationPayQrItem | undefined {
    return list.find((item) => item.amount_yuan === amountYuan) || list[0]
}

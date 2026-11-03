import type { PayPlatform } from '@/api/registrationApplicationApi'

import qrAlipay from '@/assets/qr_payment/qr_alipay.png'
import qrWepay from '@/assets/qr_payment/qr_wepay.png'
import qrAlipay6 from '@/assets/qr_payment/qr_alipay_6.png'
import qrWepay6 from '@/assets/qr_payment/qr_wepay_6.png'
import qrAlipay8 from '@/assets/qr_payment/qr_alipay_8.png'
import qrWepay8 from '@/assets/qr_payment/qr_wepay_8.png'
import qrAlipay18 from '@/assets/qr_payment/qr_alipay_18.png'
import qrWepay18 from '@/assets/qr_payment/qr_wepay_18.png'
import qrAlipay28 from '@/assets/qr_payment/qr_alipay_28.png'
import qrWepay28 from '@/assets/qr_payment/qr_wepay_28.png'
import qrAlipay38 from '@/assets/qr_payment/qr_alipay_38.png'
import qrWepay38 from '@/assets/qr_payment/qr_wepay_38.png'
import qrAlipay48 from '@/assets/qr_payment/qr_alipay_48.png'
import qrWepay48 from '@/assets/qr_payment/qr_wepay_48.png'

/** 后台可切换的注册收款档位（0 = 无固定金额收款码） */
export const REGISTRATION_PAY_AMOUNT_OPTIONS: { value: number; label: string }[] = [
    { value: 6, label: '¥6' },
    { value: 8, label: '¥8' },
    { value: 18, label: '¥18' },
    { value: 28, label: '¥28' },
    { value: 38, label: '¥38' },
    { value: 48, label: '¥48' },
    { value: 0, label: '自定义金额（无固定码）' },
]

const ALIPAY_BY_AMOUNT: Record<number, string> = {
    0: qrAlipay,
    6: qrAlipay6,
    8: qrAlipay8,
    18: qrAlipay18,
    28: qrAlipay28,
    38: qrAlipay38,
    48: qrAlipay48,
}

const WEPAY_BY_AMOUNT: Record<number, string> = {
    0: qrWepay,
    6: qrWepay6,
    8: qrWepay8,
    18: qrWepay18,
    28: qrWepay28,
    38: qrWepay38,
    48: qrWepay48,
}

export function isRegistrationPayAmount(value: number): boolean {
    return value === 0 || [6, 8, 18, 28, 38, 48].includes(value)
}

export function payQrImage(platform: PayPlatform, amountYuan: number): string {
    const map = platform === 'wechat' ? WEPAY_BY_AMOUNT : ALIPAY_BY_AMOUNT
    return map[amountYuan] ?? map[28]
}

export function formatRegistrationPayAmount(amountYuan: number | null | undefined): string {
    if (amountYuan === null || amountYuan === undefined) return '—'
    if (amountYuan === 0) return '自定义'
    return `¥${amountYuan}`
}

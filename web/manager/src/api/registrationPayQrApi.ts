import { request } from './request'
import axios from 'axios'
import { getAuthorization, updateAuthorizationToken } from '@/utility'

export type PayQrPlatform = 'alipay' | 'wechat'

export interface RegistrationPayQrItem {
    id: number
    amount_yuan: number
    alipay_url: string
    alipay_note: string
    wechat_url: string
    wechat_note: string
    date_create?: string
    date_modify?: string
}

export default {
    list() {
        return request('get', null, null, false, '/registration-pay-qr')
    },
    create(body: Omit<RegistrationPayQrItem, 'id'>) {
        return request('post', null, body, false, '/registration-pay-qr')
    },
    update(body: RegistrationPayQrItem) {
        return request('put', null, body, false, '/registration-pay-qr')
    },
    remove(id: number) {
        return request('delete', null, { id }, false, '/registration-pay-qr')
    },
    // 上传某一档的支付宝或微信收款码
    upload(id: number, platform: PayQrPlatform, file: File) {
        const auth = getAuthorization()
        const formData = new FormData()
        formData.append('file', file)
        return axios({
            url: `/portal/registration-pay-qr/${id}/upload/${platform}`,
            method: 'post',
            data: formData,
            headers: {
                Authorization: auth?.token ? `Bearer ${auth.token}` : '',
            },
            withCredentials: true,
        }).then((res) => {
            const renewed = res.headers?.['x-access-token']
            if (typeof renewed === 'string' && renewed) {
                updateAuthorizationToken(renewed)
            }
            if (res.data?.success) return res.data
            return Promise.reject(res.data)
        })
    },
}

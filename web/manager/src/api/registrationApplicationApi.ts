import { request } from '@/api/request'

export type ApplicationStatus = 'all' | 'pending' | 'approved' | 'rejected'
export type PayPlatform = 'alipay' | 'wechat'

export interface RegistrationApplicationRow {
    id: number
    email: string
    paid_account: string
    pay_platform: PayPlatform
    /** 提交时系统配置的应付金额；0 为自定义金额；历史记录可能为空 */
    pay_amount_yuan?: number | null
    status: 'pending' | 'approved' | 'rejected'
    invitation_code: string | null
    /** 1 when invitation_code was issued but the code row was deleted */
    invitation_missing?: number
    admin_note: string | null
    date_create: string
    date_review: string | null
    reviewer_uid: number | null
}

const registrationApplicationApi = {
    submit(data: { email: string; paidAccount: string; payPlatform: PayPlatform }) {
        return request('post', null, data, false, 'registration-application/submit')
    },
    checkEmail(email: string) {
        return request('get', { email }, null, false, 'registration-application/check-email')
    },
    manage(params: { status?: ApplicationStatus; pageNo?: number; pageSize?: number } = {}) {
        return request('get', params, null, false, 'registration-application/manage')
    },
    detail(params: { id: number }) {
        return request('get', params, null, false, 'registration-application/detail')
    },
    approve(data: { id: number; adminNote?: string }) {
        return request('post', null, data, false, 'registration-application/approve')
    },
    reject(data: { id: number; adminNote?: string }) {
        return request('post', null, data, false, 'registration-application/reject')
    },
    delete(params: { id: number }) {
        return request('delete', params, null, false, 'registration-application/delete')
    },
}

export default registrationApplicationApi

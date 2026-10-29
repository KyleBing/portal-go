import { request } from './request'
import axios from 'axios'
import { getAuthorization, updateAuthorizationToken } from '@/utility'

export interface AliasRule {
    match: string[]
    name?: string
    meal_place?: string
    when_product?: string[]
    slot_names?: Record<string, string>
    meal_order?: string
    guess?: boolean
    confirm?: boolean
    // 持久化参考流水
    ref_when?: string
    ref_amount?: number | null
    ref_who?: string
    ref_product?: string
}

export interface AliasSlot {
    name: string
    start: number
    end: number
}

export interface AliasConfig {
    slots: AliasSlot[]
    rules: AliasRule[]
}

export default {
    getAliases() {
        return request('get', null, null, false, '/bill-import/aliases')
    },
    putAliases(data: AliasConfig) {
        return request('put', null, data, false, '/bill-import/aliases')
    },
    seedAliases(batchId: number) {
        return request('post', null, { batchId }, false, '/bill-import/aliases/seed')
    },
    preview(batchId: number) {
        return request('get', { batchId }, null, false, '/bill-import/preview')
    },
    apply(data: {
        batchId: number
        includeGuess?: boolean
        includeConfirm?: boolean
        includeZero?: boolean
    }) {
        return request('post', null, data, true, '/bill-import/apply')
    },
    batches(params: { pageNo?: number; pageSize?: number }) {
        return request('get', params, null, false, '/bill-import/batches')
    },
    orders(batchId: number) {
        return request('get', { batchId }, null, false, '/bill-import/orders')
    },
    upload(formData: FormData) {
        const auth = getAuthorization()
        return axios({
            url: '/portal/bill-import/upload',
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

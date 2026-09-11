import { request } from './request'
import axios from 'axios'
import { getAuthorization, updateAuthorizationToken } from '@/utility'

export default {
    list(params: { pageNo?: number; pageSize?: number; keywords?: string; dateFilter?: string }) {
        return request('get', params, null, false, '/file-manager/list')
    },
    upload(formData: FormData) {
        const auth = getAuthorization()
        return axios({
            url: '/portal/file-manager/upload',
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
    modify(requestData: { fileId: number; description: string }) {
        return request('post', null, requestData, false, '/file-manager/modify')
    },
    delete(requestData: { fileId: number }) {
        return request('delete', null, requestData, false, '/file-manager/delete')
    },
}

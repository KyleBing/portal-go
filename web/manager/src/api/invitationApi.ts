import { request } from './request'

export type InvitationStatus = 'all' | 'unused' | 'used'

export default {
    manage(params: { status?: InvitationStatus; pageNo?: number; pageSize?: number } = {}) {
        return request('get', params, null, false, 'invitation/manage')
    },
    list() {
        return request('get', null, null, false, 'invitation/list')
    },
    generate() {
        return request('post', null, null, false, 'invitation/generate')
    },
    markShared(requestData: { id: string }) {
        return request('post', null, requestData, false, 'invitation/mark-shared')
    },
    delete(params: { id: string }) {
        return request('delete', params, null, false, 'invitation/delete')
    },
}

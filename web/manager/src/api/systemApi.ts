import { request } from './request'

export default {
    getPublic() { return request('get', null, null, false, '/system-config') },
    getAdmin() { return request('get', null, null, false, '/system-config/admin') },
    save(requestData: Record<string, unknown>) {
        return request('put', null, requestData, false, '/system-config')
    },
}

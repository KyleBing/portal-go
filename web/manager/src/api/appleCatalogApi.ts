import { request } from "./request";

export default {
    // full=1 把原始嵌套参数一并带回编辑框。
    list() { return request('get', { full: 1 }, null, false, '/apple-catalog') },
    saveDevice(body: object) { return request('post', {}, body, false, '/apple-catalog/device') },
    deleteDevice(id: string) { return request('delete', {}, { id }, false, '/apple-catalog/device') },
    saveChip(body: object) { return request('post', {}, body, false, '/apple-catalog/chip') },
    deleteChip(id: string) { return request('delete', {}, { id }, false, '/apple-catalog/chip') },
}

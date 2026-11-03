import { request } from "./request";

export interface ThumbsUpPayload {
    name: string
    origin_name?: string
    count: number | string
    description: string
    link_address: string
}

export default {
    add(requestData: ThumbsUpPayload) {return request('post', {}, requestData, false, '/thumbs-up/add')},
    delete(requestData: { name: string }) {return request('delete', {}, requestData, false, '/thumbs-up/delete')},
    modify(requestData: ThumbsUpPayload) {return request('put', {}, requestData, false, '/thumbs-up/modify')},
    list(params: { pageNo?: number; pageSize?: number }) {return request('get', params, null, false, '/thumbs-up/list')},
    all() {return request('get', {}, null, false, '/thumbs-up/all')},
}

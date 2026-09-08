import { request } from "./request";

interface ThumbsUpRequest {
    id?: number;
    target_id?: number;
    target_type?: string;
    uid?: string;
}

interface ThumbsUpListRequest {
    target_id?: number;
    target_type?: string;
    pageNo?: number;
    pageSize?: number;
}

export default {
    add(requestData: ThumbsUpRequest) {return request('post', {}, requestData, false, '/thumbs-up/add')},
    delete(requestData: { id: number }) {return request('delete', {}, requestData, false, '/thumbs-up/delete')},
    modify(requestData: ThumbsUpRequest) {return request('put', {}, requestData, false, '/thumbs-up/modify')},
    list(params: ThumbsUpListRequest) {return request('get', params, null, false, '/thumbs-up/list')},
    all(params: ThumbsUpListRequest) {return request('get', params, null, false, '/thumbs-up/list')},
}

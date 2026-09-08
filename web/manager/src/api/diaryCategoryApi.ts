import { request } from "./request";

interface DiaryCategoryRequest {
    id?: number;
    name?: string;
    description?: string;
    visit_count?: number;
}

interface DiaryCategoryListRequest {
    keyword?: string;
    pageNo?: number;
    pageSize?: number;
}

export default {
    add(requestData: DiaryCategoryRequest) {return request('post', {}, requestData, false, '/diary-category/add')},
    delete(requestData: { id: number }) {return request('delete', {}, requestData, false, '/diary-category/delete')},
    modify(requestData: DiaryCategoryRequest) {return request('put', {}, requestData, false, '/diary-category/modify')},
    detail(params: { id: number }) {return request('get', params, null, false, '/diary-category/detail')},
    list(params: DiaryCategoryListRequest) {return request('get', params, null, false, '/diary-category/list')},
    clearCount(requestData: { id: number }) {return request('post', {}, requestData, false, '/diary-category/clear-visit-count')},
}

import { request } from "./request";

interface CodeRequest {
    id?: number;
    name?: string;
    url?: string;
    description?: string;
    visit_count?: number;
}

interface CodeListRequest {
    keyword?: string;
    pageNo?: number;
    pageSize?: number;
}

export default {
    add(requestData: CodeRequest) {return request('post', {}, requestData, false, '/qr-manager/add')},
    delete(requestData: { id: number }) {return request('delete', {}, requestData, false, '/qr-manager/delete')},
    modify(requestData: CodeRequest) {return request('put', {}, requestData, false, '/qr-manager/modify')},
    detail(params: { id: number }) {return request('get', params, null, false, '/qr-manager/detail')},
    list(params: CodeListRequest) {return request('get', params, null, false, '/qr-manager/list')},
    clearCount(requestData: { id: number }) {return request('post', {}, requestData, false, '/qr-manager/clear-visit-count')},
}

import { request } from "./request";

export interface ImageQiniuRequest {
    id?: number;
    name?: string;
    url?: string;
    description?: string;
}

export interface ImageQiniuListRequest {
    keyword?: string;
    pageNo?: number;
    pageSize?: number;
}

export interface UploadTokenResponse {
    token: string;
    key: string;
}

export default {
    getUploadToken(params?: {}): Promise<UploadTokenResponse> {return request('get', params, null, false, '/image-qiniu/')},
    list(params: ImageQiniuListRequest) {return request('get', params, null, false, '/image-qiniu/list')},
    add(requestData: ImageQiniuRequest) {return request('post', null, requestData, false, '/image-qiniu/add')},
    delete(requestData: { id: string; bucket: string }) {return request('delete', null, requestData, false, '/image-qiniu/delete')},
    batchDelete(requestData: { ids: string[]; bucket: string }) {return request('delete', null, requestData, false, '/image-qiniu/batch-delete')},
    update(requestData: { id: string; description: string }) {return request('put', null, requestData, false, '/image-qiniu/update')},
}

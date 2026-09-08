import { request } from "./request";

interface DiaryRequest {
    id?: number;
    title?: string;
    content?: string;
    category_id?: number;
    tags?: string[];
    is_public?: boolean;
}

interface DiaryListRequest {
    category_id?: number;
    keyword?: string;
    pageNo?: number;
    pageSize?: number;
    dateRange?: string[];
}

export default {
    // login(requestData) {return request('post', null, requestData, false, null, '/user/login')}
    add(requestData: DiaryRequest) {return request('post', {}, requestData, false, '/diary/add')},
    delete(requestData: { id: number }) {return request('delete', {}, requestData, false, '/diary/delete')},
    modify(requestData: DiaryRequest) {return request('put', {}, requestData, false, '/diary/modify')},
    detail(params: { id: number }) {return request('get', params, null, false, '/diary/detail')},
    list(requestData: DiaryListRequest) { return request('post', null, requestData, false, '/diary/list') },
    /**
     * 获取最新的公开日记
     */
    getLatestPublicDiaryWithKeyword(params: { keyword: string }) { return request('get', params, null, false, '/diary/get-latest-public-diary-with-keyword') },

}

import { request } from "./request";

interface CategoryRequest {
    id?: number;
    name?: string;
    description?: string;
}

interface WordRequest {
    id?: number;
    word?: string;
    code?: string;
    priority?: number;
    up?: number;
    down?: number;
    comment?: string;
    category_id?: number;
    uid?: string;
    approved?: number;
}

interface WordListRequest {
    category_id?: number;
    dateRange?: string[];
    keyword?: string;
    pageNo?: number;
    pageSize?: number;
    approved?: number;
}

interface WordBatchRequest {
    ids?: number[];
    category_id?: number;
    approved?: number;
    words?: string[];
}

export default {
    category: {
        add(requestData: CategoryRequest) {return request('post', {}, requestData, false, '/wubi/category/add')},
        delete(requestData: { id: number }) {return request('delete', {}, requestData, false, '/wubi/category/delete')},
        modify(requestData: CategoryRequest) {return request('put', {}, requestData, false, '/wubi/category/modify')},
        detail(params: { id: number }) {return request('get', params, null, false, '/wubi/category/detail')},
        list(params?: {}) {return request('get', params, null, false, '/wubi/category/list')},
    },
    word: {
        addBatch(requestData: WordBatchRequest) {return request('post', {}, requestData, false, '/wubi/word/add-batch')},
        add(requestData: WordRequest) {return request('post', {}, requestData, false, '/wubi/word/add')},
        checkExist(requestData: { word: string, code: string }) {return request('post', {}, requestData, false, '/wubi/word/check-exist')},
        delete(requestData: { ids: number[] }) {return request('delete', {}, requestData, false, '/wubi/word/delete')},
        modify(requestData: WordRequest) {return request('put', {}, requestData, false, '/wubi/word/modify')},
        detail(params: { id: number }) {return request('get', params, null, false, '/wubi/word/detail')},
        list(requestData: WordListRequest) {return request('post', null, requestData, false, '/wubi/word/list')},
        exportExtraWords() {return request('post', null, null, false, '/wubi/word/export-extra')},
        modifyBatch(requestData: WordBatchRequest) {return request('put', null, requestData, false, '/wubi/word/modify-batch')},
    }
}

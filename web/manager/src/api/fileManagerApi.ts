import { request } from "./request";

interface FileListRequest {
    path?: string;
    keyword?: string;
    pageNo?: number;
    pageSize?: number;
}

interface FileUploadRequest {
    file: File;
    path?: string;
}

interface FileDeleteRequest {
    path: string;
    filename: string;
}

export default {
    list(params: FileListRequest) {return request('get', params, null, false, '/file-manager/list')},
    upload(requestData: FileUploadRequest) {return request('post', null, requestData, false, '/file-manager/upload')},
    delete(requestData: FileDeleteRequest) {return request('delete', null, requestData, false, '/file-manager/delete')},
}

import { request } from "./request";

interface UserRequest {
    id?: number;
    name?: string;
    email?: string;
    password?: string;
    role?: string;
    status?: number;
}


interface UserListRequest {
    keyword?: string;
    pageNo?: number;
    pageSize?: number;
    role?: string;
    status?: number;
}

interface LoginRequest {
    email: string;
    password: string;
}

interface RegisterRequest {
    nickname: string;
    email: string;
    password: string;
    invitationCode?: string;
}

interface ChangePasswordRequest {
    oldPassword: string;
    newPassword: string;
}

interface UserResponse {
    id: number;
    name: string;
    email: string;
    role: string;
    status: number;
    date_create: string;
    date_modify: string;
}

export default {
    login(requestData: LoginRequest): Promise<UserResponse> {return request('post', null, requestData, false, '/user/login')},
    register(requestData: RegisterRequest): Promise<UserResponse> {return request('post', {}, requestData, false, '/user/register')},

    add(requestData: UserRequest): Promise<UserResponse> {return request('post', {}, requestData, false, '/user/add')},
    delete(requestData: { id: number }): Promise<{ message: string }> {return request('delete', {}, requestData, false, '/user/delete')},
    modify(requestData: UserRequest): Promise<UserResponse> {return request('put', {}, requestData, false, '/user/modify')},
    detail(params: { id: number }): Promise<UserResponse> {return request('get', params, null, false, '/user/detail')},
    list(requestData: UserListRequest): Promise<{ list: UserResponse[], pager: { total: number, pageNo: number, pageSize: number } }> {return request('post', null, requestData, false, '/user/list')},
    changePassword(requestData: ChangePasswordRequest): Promise<{ message: string }> {return request('put', null, requestData, false, '/user/change-password')},
    // 找回密码 / 重发验证
    forgot(requestData: { email: string }): Promise<{ message: string }> {return request('post', null, requestData, false, '/user/forgot')},
    resendVerify(requestData: { email: string }): Promise<{ message: string }> {return request('post', null, requestData, false, '/user/resend-verify')},
    // 管理员：强制验证 / 发送重置邮件
    forceVerify(requestData: { uid: number | string }): Promise<{ message: string }> {return request('post', null, requestData, false, '/user/force-verify')},
    sendResetPassword(requestData: { uid: number | string }): Promise<{ message: string }> {return request('post', null, requestData, false, '/user/send-reset-password')},
}

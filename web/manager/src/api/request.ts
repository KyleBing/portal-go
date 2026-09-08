import axios from "axios";
import {ElLoading, ElMessage} from "element-plus";
import {getAuthorization} from "@/utility.ts";


const LOADING_OPTION = {
    lock: true,
    text: "载入中，请稍候...",
    background: "rgba(0, 0, 0, 0.7)"
}

const BASE_URL = '/portal' // absolute API prefix; paths already start with '/'

function joinURL(base: string, path: string) {
    const b = base.replace(/\/+$/, '')
    const p = path.startsWith('/') ? path : `/${path}`
    return `${b}${p}`
}

function request(
    method: 'get' | 'post' | 'delete' | 'patch' | 'put',
    params: object | null,
    requestData: object | null,
    showLoading = false, url: string) {
    let layerLoading = null
    if (showLoading) layerLoading = ElLoading.service(LOADING_OPTION)

    let headers = {}
    /*
    * 所有 requestData 都会自动添加  authorization 信息
    * 给 requestData 添加 authorization 内部的数据： username email uid 等等
    * */
    const path = url.replace(/^\/+/, '')
    if (path !== 'user/login' && path !== 'user/register'){ // 注册和登录时不添加 Token 数据
        Object.assign(headers, {
            'Diary-Token':  getAuthorization() && getAuthorization().token,
            'Diary-Uid':  getAuthorization() && getAuthorization().uid
        })
    }

    return new Promise((resolve, reject) => {
        axios({
            url: joinURL(BASE_URL, url),
            method,
            data: requestData,
            params,
            headers,
            withCredentials: true
        })
            .then(res => {
                if (showLoading) layerLoading.close()
                if (res.status === 200) {
                    if (res.data.success){
                        resolve(res.data)
                    } else {
                        ElMessage.error({
                            message: res.data.message || 'Error'
                        })
                        reject(res.data)
                    }
                } else {
                    console.log('request err: ', res.data) // 如果演示模式，不用显示网络请求错误
                }
            })
            .catch(err => {
                reject(err)
                if (showLoading) layerLoading.close()
                ElMessage.error({
                    message: err.message
                })
                console.log(err, err.message)
            })
    })
}


export {request}

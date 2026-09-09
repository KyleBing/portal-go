import dayjs from "dayjs"

interface AuthorizationData {
    nickname: string,
    uid: string,
    email: string,
    phone: string,
    avatar: string,
    token: string,
    group_id: string,
    city: string,
    geolocation: string
}

const AUTHORIZATION_NAME = 'Authorization' // 存储用户信息的 localStorage name，跟 Diary 通用

function downloadBase64File(fileName: string, data: string) { // 下载 base64 图片
    let aLink = document.createElement('a')
    let blob = new Blob([data]); //new Blob([content])
    let evt = document.createEvent("HTMLEvents")
    evt.initEvent("click", true, true); //initEvent 不加后两个参数在FF下会报错  事件类型，是否冒泡，是否阻止浏览器的默认行为
    aLink.download = fileName
    aLink.href = URL.createObjectURL(blob)
    aLink.click()
}

function downloadFile(fileName: string, url: string) { // 下载文件
    let aLink = document.createElement('a')
    let evt = document.createEvent("HTMLEvents")
    evt.initEvent("click", true, true); //initEvent 不加后两个参数在FF下会报错  事件类型，是否冒泡，是否阻止浏览器的默认行为
    aLink.download = fileName
    aLink.href = url
    aLink.click()
}

// 设置 authorization
function setAuthorization(
    nickname: string,
    uid: string,
    email: string,
    phone: string,
    avatar: string,
    token: string,
    group_id: string,
    city: string,
    geolocation: string
): void {
    const authData: AuthorizationData = {
        nickname,
        uid,
        email,
        phone,
        avatar,
        token,
        group_id,
        city,
        geolocation
    }
    localStorage.setItem(AUTHORIZATION_NAME, JSON.stringify(authData))
}

// 获取 authorization
function getAuthorization(): AuthorizationData | null {
    const authStr = localStorage.getItem(AUTHORIZATION_NAME)
    return authStr ? JSON.parse(authStr) : null
}

// 删除 authorization
function deleteAuthorization(): void {
    localStorage.removeItem(AUTHORIZATION_NAME)
}

function autoScale(selector, option) {
    const el = document.getElementById(selector);
    const {width, height} = option;
    function init() {
        const scaleX = innerWidth / width;
        const scaleY = innerHeight / height;
        const scale  = Math.min(scaleX, scaleY);
        const left = (innerWidth - width * scale) / 2;
        const top = (innerHeight - height * scale) / 2;
        el.style.transform = `translate(${left}px, ${top}px) scale(${scale})`;
        el.style.position = "static";
        el.style.transformOrigin = "left top";
    }
    init();
}

function resizeHolePage (routePath: string , containerDom: HTMLElement) {
    console.log('--- Util: Render function is working')

    let fullHeight = document.documentElement.clientHeight;
    let fullWidth = document.documentElement.clientWidth;
    let scaleWidth = fullWidth / 1920;
    let scaleHeight = fullHeight / 1080;
    let scale = scaleWidth;
    let appBox = document.getElementById("app");
    // 大屏时，画面内容全部展示在画面内
    if (routePath.indexOf('/bigscreen') > -1 || routePath.indexOf('/multipleBigScreen') > -1 || routePath.indexOf('/energyBigScreen') > -1) {
        // scale = scaleWidth > scaleHeight ? scaleHeight : scaleWidth;
        // containerDom.style.transform = `translate(-50%, 0%) scale(${scale})`;
        autoScale('app', {
            width: 1920,
            height: 1080
        });

        // 大屏目前不用在这处理
    } else {
        appBox.style.transformOrigin = "top";
        containerDom.style.position = "absolute";
        containerDom.style.transform = `translate(-50%,0%) scale(${scale})`
    }

    containerDom.setAttribute('data-scale', scale.toString())  // 将 scale 存储到 属性中，下面会用

    // store.commit('SET_SCREEN_ZOOM_SCALE', scale) // 将这个缩放比例存到 store 中

    // let new_element = document.getElementById("document_css")
    // if (routePath.indexOf('/bigscreen') > -1){
    //     new_element.innerText = `body{background-color: #036aae}!important;}`
    // } else{
    //     new_element.innerText = `body{background-color: white}!important;}`
    // }
    // document.body.appendChild(new_element)
}

function dateFormatter(date: Date, format: string = 'yyyy-MM-dd HH:mm:ss'): string {
    const dayjsFormat = format
        .replace(/yyyy/g, 'YYYY')
        .replace(/dd/g, 'DD')
    return dayjs(date).format(dayjsFormat)
}

export {
    getAuthorization,
    setAuthorization,
    deleteAuthorization,
    downloadFile,
    downloadBase64File,
    resizeHolePage,
    dateFormatter,

    type AuthorizationData
}

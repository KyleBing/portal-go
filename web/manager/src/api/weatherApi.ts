import { request } from './request'

function withH(hours: string) {
    const h = String(hours || '24h').trim()
    return h.endsWith('h') ? h : `${h}h`
}

function withD(days: string) {
    const d = String(days || '7d').trim()
    return d.endsWith('d') ? d : `${d}d`
}

export default {
    status() {
        return request('get', null, null, false, '/weather/status')
    },
    now(location: string) {
        return request('get', { location, lang: 'zh' }, null, true, '/v7/weather/now')
    },
    hourly(location: string, hours = '24h') {
        return request('get', { location, lang: 'zh' }, null, true, `/v7/weather/${withH(hours)}`)
    },
    daily(location: string, days: string) {
        return request('get', { location, lang: 'zh' }, null, true, `/v7/weather/${withD(days)}`)
    },
    geo(location: string, number = 10) {
        return request('get', { location, number, lang: 'zh', range: 'cn' }, null, true, '/geo/v2/city/lookup')
    },
}

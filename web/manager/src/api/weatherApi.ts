import { request } from './request'

export default {
    status() {
        return request('get', null, null, false, '/weather/status')
    },
    now(location: string) {
        return request('get', { location }, null, true, '/weather/now')
    },
    hourly(location: string, hours = '24h') {
        return request('get', { location, hours }, null, true, '/weather/hourly')
    },
    daily(location: string, days: string) {
        return request('get', { location, days }, null, true, '/weather/daily')
    },
    geo(location: string, number = 10) {
        return request('get', { location, number }, null, true, '/weather/geo')
    },
}

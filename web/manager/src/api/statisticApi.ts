import { request } from './request'

export default {
    main() { return request('get', {}, {}, false, 'statistic/')},
    category() { return request('get', {}, {}, false, 'statistic/category')},
    year() { return request('get', {}, {}, false, 'statistic/year')},
    /** @deprecated removed; use managerUsers */
    users() { return request('get', {}, {}, false, 'statistic/manager-users')},
    managerUsers() { return request('get', {}, {}, false, 'statistic/manager-users')},
    userWordsCountData() { return request('get', {}, {}, false, 'statistic/user-data-words')},
    userDiaryCountData() { return request('get', {}, {}, false, 'statistic/user-data-diary')},
}

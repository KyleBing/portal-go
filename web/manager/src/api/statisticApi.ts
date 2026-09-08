import { request } from './request'

interface StatisticMainResponse {
    total: number;
    categories: number;
    users: number;
    words: number;
    diaries: number;
}

interface StatisticCategoryResponse {
    categories: {
        id: number;
        name: string;
        count: number;
    }[];
}

interface StatisticYearResponse {
    years: {
        year: number;
        count: number;
    }[];
}

interface StatisticUsersResponse {
    users: {
        id: number;
        name: string;
        count: number;
    }[];
}

interface UserDataResponse {
    data: {
        date: string;
        count: number;
    }[];
}

export default {
    main(): Promise<StatisticMainResponse> { return request('get', {}, {}, false, 'statistic/')},
    category(): Promise<StatisticCategoryResponse> { return request('get', {}, {}, false, 'statistic/category')},
    year(): Promise<StatisticYearResponse> { return request('get', {}, {}, false, 'statistic/year')},
    users(): Promise<StatisticUsersResponse> { return request('get', {}, {}, false, 'statistic/users')},
    userWordsCountData(): Promise<UserDataResponse> { return request('get', {}, {}, false, 'statistic/user-data-words')},
    userDiaryCountData(): Promise<UserDataResponse> { return request('get', {}, {}, false, 'statistic/user-data-diary')},
}

import { request } from "./request";

interface BillBriefResponse {
    total: number;
    income: number;
    expense: number;
    categories: {
        id: number;
        name: string;
        amount: number;
    }[];
}

export default {
    billBrief(): Promise<BillBriefResponse> {return request('get', {}, null, false, '/bill/')},
}

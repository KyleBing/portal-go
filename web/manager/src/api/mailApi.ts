import { request } from "./request";

interface MailRequest {
    to: string;
    subject: string;
    content: string;
    html?: boolean;
}

export default {
    sendEmail(requestData: MailRequest) {return request('post', null, requestData, false, '/mail/send')},
}

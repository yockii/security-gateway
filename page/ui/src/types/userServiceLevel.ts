import {Service} from "./service";

export type UserServiceLevel = {
    id?: string;
    userId?: string;
    serviceId?: string;
    secLevel?: number;
    createTime?: string;

    service?: Service;

    // 分页
    page?: number;
    pageSize?: number;
};

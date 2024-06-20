import {Upstream} from "./upstream";

export type Route = {
    id?: string;
    serviceId?: string;
    uri?: string;
    loadBalance?: number;
    createTime?: string;

    // 分页
    page?: number;
};

export type RouteWithTarget = {
    id?: string;
    serviceId?: string;
    uri?: string;
    createTime?: string;
    target?: Upstream;
};

export type RouteWithTargets = {
    id?: string;
    serviceId?: string;
    uri?: string;
    createTime?: string;
    targets?: Upstream[];
};

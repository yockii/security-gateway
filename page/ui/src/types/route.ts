import {Upstream} from "./upstream";

export type Route = {
    id?: string;
    serviceId?: string;
    uri?: string;
    createTime?: string;
};

export type RouteWithTarget = {
    id?: string;
    serviceId?: string;
    uri?: string;
    createTime?: string;
    target?: Upstream;
};

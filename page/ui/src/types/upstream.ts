export type Upstream = {
    id?: string;
    name?: string;
    targetUrl?: string;

    healthCheckUrl?: string;
    status?: number;
    lastCheckTime?: number;

    createTime?: string;

    // 分页查询参数
    page?: number;
    pageSize?: number;
};

export type TargetWithUpstream = {
    id?: string;
    routeId?: string;
    upstreamId?: string;
    weight?: number;
    upstream?: Upstream;
};

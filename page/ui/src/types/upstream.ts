export type Upstream = {
    id?: string;
    name?: string;
    targetUrl?: string;
    createTime?: string;

    // 分页查询参数
    page?: number;
    pageSize?: number;
};

export type Upstream = {
    id?: string;
    name?: string;
    targetUrl?: string;
    createTime?: string;

    // 权重
    weight?: number;

    // 分页查询参数
    page?: number;
    pageSize?: number;
};

export type User = {
    id?: string;
    username?: string;
    uniKey?: string;
    uniKeysJson?: string;
    secLevel?: number;
    createTime?: string;

    // 分页查询
    page?: number;
    pageSize?: number;
};

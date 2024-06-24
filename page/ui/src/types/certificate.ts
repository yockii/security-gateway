export type Certificate = {
    id?: string;
    certName?: string;
    serveDomain?: string;
    certDesc?: string;
    certPem?: string;
    keyPem?: string;
    createTime?: string;

    // 分页
    page?: number;
    pageSize?: number;
};

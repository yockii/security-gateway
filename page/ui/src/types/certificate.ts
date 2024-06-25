export type Certificate = {
    id?: string;
    certName?: string;
    serveDomain?: string;
    certDesc?: string;
    certPem?: string;
    keyPem?: string;
    signCertPem?: string;
    signKeyPem?: string;
    encCertPem?: string;
    encKeyPem?: string;
    createTime?: string;

    // 分页
    page?: number;
    pageSize?: number;
};

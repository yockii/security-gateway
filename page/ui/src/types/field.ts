export type ServiceField = {
    id?: string;
    serviceId?: string;
    fieldName?: string;
    comment?: string;
    level1?: string;
    level2?: string;
    level3?: string;
    level4?: string;
    createTime?: string;

    // 分页参数
    page?: number;
    pageSize?: number;
};

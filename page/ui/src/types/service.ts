export type Port = {
    port: number;
    inUse: boolean;
};

export type Service = {
    id?: string;
    name?: string;
    domain?: string;
    port?: number;
    certificateId?: string;
    createTime?: string;

    // 分页参数
    page?: number;
};

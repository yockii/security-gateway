export type Port = {
    port: number;
    inUse: boolean;
};

export type Service = {
    id?: string;
    name?: string;
    domain?: string;
    port?: number;
    createTime?: string;
};

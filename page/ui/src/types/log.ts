export type ProxyTraceLog = {
    id: number;
    customIp?: string;
    domain?: string;
    maskingLevel?: number;
    path?: string;
    port?: number;
    targetUrl?: string;
    username?: string;
    startTime?: number;
    endTime?: number;
    count?: number;

    inEdit?: boolean;
};

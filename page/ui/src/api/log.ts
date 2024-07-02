import {Response} from "@/types/common";
import {get} from "./api";
import {ProxyTraceLog} from "@/types/log";

export async function countProxyTraceLog(
    params: ProxyTraceLog
): Promise<Response<number>> {
    return get("/api/v1/log/count", params);
}

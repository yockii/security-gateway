import {PaginationResponse, Response} from "@/types/common";
import {get, post} from "./api";
import {TargetWithUpstream, Upstream} from "@/types/upstream";

export async function getUpstream(id: string): Promise<Response<Upstream>> {
    return get(`/api/v1/upstream/${id}`);
}

export async function getUpstreamList(
    params: Upstream
): Promise<Response<PaginationResponse<Upstream>>> {
    return get("/api/v1/upstream/list", params);
}

export async function addUpstream(data: Upstream): Promise<Response<Upstream>> {
    return post("/api/v1/upstream/add", data);
}

export async function updateUpstream(
    data: Upstream
): Promise<Response<Upstream>> {
    return post("/api/v1/upstream/update", data);
}

export async function deleteUpstream(id: string): Promise<Response<boolean>> {
    return post(`/api/v1/upstream/delete/${id}`);
}

export async function getUpstreamListByRoute(
    routeId: string,
    page?: number,
    name?: string
): Promise<Response<PaginationResponse<TargetWithUpstream>>> {
    return get(`/api/v1/upstream/listByRoute`, {routeId, page, name});
}

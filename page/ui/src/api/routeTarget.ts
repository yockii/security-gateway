import {PaginationResponse, Response} from "@/types/common";
import {get, post} from "./api";
import {RouteTarget} from "@/types/routeTarget";

export async function getRouteTarget(
    id: string
): Promise<Response<RouteTarget>> {
    return get(`/api/v1/routeTarget/${id}`);
}

export async function getRouteTargetList(
    params: RouteTarget
): Promise<Response<PaginationResponse<RouteTarget>>> {
    return get("/api/v1/routeTarget/list", params);
}

export async function saveRouteTarget(
    data: RouteTarget
): Promise<Response<RouteTarget>> {
    return post("/api/v1/routeTarget/save", data);
}

export async function deleteRouteTarget(
    id: string
): Promise<Response<boolean>> {
    return post(`/api/v1/routeTarget/delete/${id}`);
}

export async function deleteByRouteAndUpstreamID(
    routeId: string,
    upstreamId: string
): Promise<Response<boolean>> {
    return post("/api/v1/routeTarget/deleteByRouteAndUpstream", {
        routeId,
        upstreamId,
    });
}

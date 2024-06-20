import {PaginationResponse, Response} from "@/types/common";
import {get, post} from "./api";
import {Route, RouteWithTargets} from "@/types/route";

export async function getRoute(id: string): Promise<Response<Route>> {
    return get(`/api/v1/route/${id}`);
}

export async function getRouteList(
    params: Route
): Promise<Response<PaginationResponse<Route>>> {
    return get("/api/v1/route/list", params);
}

export async function addRoute(data: Route): Promise<Response<Route>> {
    return post("/api/v1/route/add", data);
}

export async function updateRoute(data: Route): Promise<Response<Route>> {
    return post("/api/v1/route/update", data);
}

export async function deleteRoute(id: string): Promise<Response<boolean>> {
    return post(`/api/v1/route/delete/${id}`);
}

export async function getRouteWithTargetsList(
    params: Route
): Promise<Response<PaginationResponse<RouteWithTargets>>> {
    return get("/api/v1/route/listWithTargets", params);
}

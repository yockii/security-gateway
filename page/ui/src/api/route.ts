import {PaginationResponse, Response} from "@/types/common";
import {get, post} from "./api";
import {Route, RouteWithTarget} from "@/types/route";

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

export async function getRouteWithTargetList(
    params: Route
): Promise<Response<PaginationResponse<RouteWithTarget>>> {
    return get("/api/v1/route/listWithTarget", params);
}

import {PaginationResponse, Response} from "@/types/common";
import {get, post} from "./api";
import {RouteField} from "@/types/field";

export async function getField(id: string): Promise<Response<RouteField>> {
    return get(`/api/v1/routeField/instance/${id}`);
}

export async function getFieldList(
    params: RouteField
): Promise<Response<PaginationResponse<RouteField>>> {
    return get("/api/v1/routeField/list", params);
}

export async function addField(
    data: RouteField
): Promise<Response<RouteField>> {
    return post("/api/v1/routeField/add", data);
}

export async function updateField(
    data: RouteField
): Promise<Response<RouteField>> {
    return post("/api/v1/routeField/update", data);
}

export async function deleteField(id: string): Promise<Response<boolean>> {
    return post(`/api/v1/routeField/delete/${id}`);
}

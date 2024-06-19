import {PaginationResponse, Response} from "@/types/common";
import {get, post} from "./api";
import {ServiceField} from "@/types/field";

export async function getField(id: string): Promise<Response<ServiceField>> {
    return get(`/api/v1/serviceField/instance/${id}`);
}

export async function getFieldList(
    params: ServiceField
): Promise<Response<PaginationResponse<ServiceField>>> {
    return get("/api/v1/serviceField/list", params);
}

export async function addField(
    data: ServiceField
): Promise<Response<ServiceField>> {
    return post("/api/v1/serviceField/add", data);
}

export async function updateField(
    data: ServiceField
): Promise<Response<ServiceField>> {
    return post("/api/v1/serviceField/update", data);
}

export async function deleteField(id: string): Promise<Response<boolean>> {
    return post(`/api/v1/serviceField/delete/${id}`);
}

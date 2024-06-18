import {PaginationResponse, Response} from "@/types/common";
import {get, post} from "./api";
import {Field} from "@/types/field";

export async function getField(id: string): Promise<Response<Field>> {
    return get(`/api/v1/secretField/${id}`);
}

export async function getFieldList(
    params: Field
): Promise<Response<PaginationResponse<Field>>> {
    return get("/api/v1/secretField/list", params);
}

export async function addField(data: Field): Promise<Response<Field>> {
    return post("/api/v1/secretField/add", data);
}

export async function updateField(data: Field): Promise<Response<Field>> {
    return post("/api/v1/secretField/update", data);
}

export async function deleteField(id: string): Promise<Response<boolean>> {
    return post(`/api/v1/secretField/delete/${id}`);
}

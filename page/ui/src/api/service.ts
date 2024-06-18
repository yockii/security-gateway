import {PaginationResponse, Response} from "@/types/common";
import {Port, Service} from "@/types/service";
import {get, post} from "./api";

export async function getAllPorts(): Promise<Response<Port[]>> {
    return get("/api/v1/service/ports");
}

export async function getService(id: string): Promise<Response<Service>> {
    return get(`/api/v1/service/instance/${id}`);
}

export async function getServiceList(
    params: Service
): Promise<Response<PaginationResponse<Service>>> {
    return get("/api/v1/service/list", params);
}

export async function addService(data: Service): Promise<Response<Service>> {
    return post("/api/v1/service/add", data);
}

export async function updateService(data: Service): Promise<Response<Service>> {
    return post("/api/v1/service/update", data);
}

export async function deleteService(id: string): Promise<Response<boolean>> {
    return post(`/api/service/delete/${id}`);
}

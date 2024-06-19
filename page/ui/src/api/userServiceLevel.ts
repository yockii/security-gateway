import {PaginationResponse, Response} from "@/types/common";
import {get, post} from "./api";
import {UserServiceLevel} from "@/types/userServiceLevel";

export async function getUserServiceLevel(
    id: string
): Promise<Response<UserServiceLevel>> {
    return get(`/api/v1/userServiceLevel/${id}`);
}

export async function getUserServiceLevelList(
    params: UserServiceLevel
): Promise<Response<PaginationResponse<UserServiceLevel>>> {
    return get("/api/v1/userServiceLevel/list", params);
}

export async function addUserServiceLevel(
    data: UserServiceLevel
): Promise<Response<UserServiceLevel>> {
    return post("/api/v1/userServiceLevel/add", data);
}

export async function updateUserServiceLevel(
    data: UserServiceLevel
): Promise<Response<UserServiceLevel>> {
    return post("/api/v1/userServiceLevel/update", data);
}

export async function deleteUserServiceLevel(
    id: string
): Promise<Response<boolean>> {
    return post(`/api/v1/userServiceLevel/delete/${id}`);
}

export async function getUserServiceLevelListWithService(
    params: UserServiceLevel
): Promise<Response<PaginationResponse<UserServiceLevel>>> {
    return get("/api/v1/userServiceLevel/ListWithService", params);
}

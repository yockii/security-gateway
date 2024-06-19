import {PaginationResponse, Response} from "@/types/common";
import {get, post} from "./api";
import {User} from "@/types/user";

export async function getUser(id: string): Promise<Response<User>> {
    return get(`/api/v1/user/${id}`);
}

export async function getUserList(
    params: User
): Promise<Response<PaginationResponse<User>>> {
    return get("/api/v1/user/list", params);
}

export async function addUser(data: User): Promise<Response<User>> {
    return post("/api/v1/user/add", data);
}

export async function updateUser(data: User): Promise<Response<User>> {
    return post("/api/v1/user/update", data);
}

export async function deleteUser(id: string): Promise<Response<boolean>> {
    return post(`/api/v1/user/delete/${id}`);
}

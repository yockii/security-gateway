import {Response} from "@/types/common";
import {get, post} from "./api";
import {UserInfoRoute} from "@/types/userInfoRoute";

export async function getUserInfoRouteByServiceId(
    serviceId: string
): Promise<Response<UserInfoRoute>> {
    return get(`/api/v1/userInfoRoute/instanceByService/${serviceId}`);
}

export async function addUserInfoRoute(
    data: UserInfoRoute
): Promise<Response<UserInfoRoute>> {
    return post("/api/v1/userInfoRoute/add", data);
}

export async function updateUserInfoRoute(
    data: UserInfoRoute
): Promise<Response<UserInfoRoute>> {
    return post("/api/v1/userInfoRoute/update", data);
}

export async function deleteUserInfoRoute(
    id: string
): Promise<Response<boolean>> {
    return post(`/api/v1/userInfoRoute/delete/${id}`);
}

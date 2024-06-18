import http from "@/utils/http";

import {Response} from "@/types/common";

export async function get<T>(url: string, params?: any): Promise<Response<T>> {
    const response = await http.get(url, {params});
    return response.data;
}

export async function post<T>(url: string, data?: any): Promise<Response<T>> {
    const response = await http.post(url, data);
    return response.data;
}

export async function put<T>(url: string, data?: any): Promise<Response<T>> {
    const response = await http.put(url, data);
    return response.data;
}

export async function del<T>(url: string, data?: any): Promise<Response<T>> {
    const response = await http.delete(url, {data});
    return response.data;
}

import {PaginationResponse, Response} from "@/types/common";
import {get, post} from "./api";
import {Certificate} from "@/types/certificate";

export async function getCertificate(
    id: string
): Promise<Response<Certificate>> {
    return get(`/api/v1/certificate/${id}`);
}

export async function getCertificateList(
    params: Certificate
): Promise<Response<PaginationResponse<Certificate>>> {
    return get("/api/v1/certificate/list", params);
}

export async function addCertificate(
    data: Certificate
): Promise<Response<Certificate>> {
    return post("/api/v1/certificate/add", data);
}

export async function updateCertificate(
    data: Certificate
): Promise<Response<Certificate>> {
    return post("/api/v1/certificate/update", data);
}

export async function deleteCertificate(
    id: string
): Promise<Response<boolean>> {
    return post(`/api/v1/certificate/delete/${id}`);
}

export async function listByDomain(
    domain: string
): Promise<Response<Certificate[]>> {
    return get(`/api/v1/certificate/listByDomain`, {domain});
}

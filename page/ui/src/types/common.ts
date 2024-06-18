export interface Response<T> {
    code: number;
    data: T;
    msg: string;
}

export interface PaginationResponse<T> {
    items: T[];
    total: number;
}

import axios, {
  AxiosError,
  AxiosInstance,
  AxiosRequestConfig,
  AxiosResponse,
} from "axios";

export interface ErrorResponse {
  error: {
    code: string;
    message?: string;
    fields?: Record<string, string>;
    sku?: string;
    available?: number;
  };
}

export class ApiError extends Error {
  readonly status: number;
  readonly code: string;
  readonly fields?: Record<string, string>;
  readonly sku?: string;
  readonly available?: number;

  constructor(status: number, detail: ErrorResponse["error"]) {
    super(detail.message ?? detail.code);
    this.name = "ApiError";
    this.status = status;
    this.code = detail.code;
    this.fields = detail.fields;
    this.sku = detail.sku;
    this.available = detail.available;
  }
}

function toApiError(error: AxiosError): ApiError {
  if (error.response) {
    const body = error.response.data as ErrorResponse | undefined;
    const detail = body?.error ?? { code: "UNKNOWN", message: error.message };
    return new ApiError(error.response.status, detail);
  }
  if (error.code === "ECONNABORTED") {
    return new ApiError(0, { code: "TIMEOUT", message: "Request timed out" });
  }
  return new ApiError(0, {
    code: "NETWORK",
    message: error.message || "Network error",
  });
}

const instance: AxiosInstance = axios.create({
  baseURL: "/api", // resolved by the Vite dev proxy (localhost:8080) or nginx in Docker
  timeout: 15_000,
  headers: {
    "Content-Type": "application/json",
  },
});

instance.interceptors.response.use(
  (response: AxiosResponse) => response.data,
  (error: AxiosError) => Promise.reject(toApiError(error)),
);

type ApiClient = {
  get<T = unknown>(url: string, config?: AxiosRequestConfig): Promise<T>;
  delete<T = unknown>(url: string, config?: AxiosRequestConfig): Promise<T>;
  post<T = unknown>(
    url: string,
    data?: unknown,
    config?: AxiosRequestConfig,
  ): Promise<T>;
  put<T = unknown>(
    url: string,
    data?: unknown,
    config?: AxiosRequestConfig,
  ): Promise<T>;
  patch<T = unknown>(
    url: string,
    data?: unknown,
    config?: AxiosRequestConfig,
  ): Promise<T>;
};

export const http: ApiClient = instance as unknown as ApiClient;

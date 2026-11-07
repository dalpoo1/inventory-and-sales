// Sales-order endpoints.
import { http } from "./client";
import type { CreateOrderInput, Order, Paginated } from "./types";

export interface OrderListParams {
  page?: number;
  limit?: number;
}

export const ordersApi = {
  listOrders(params: OrderListParams = {}) {
    return http.get<Paginated<Order>>("/orders", { params });
  },

  createOrder(input: CreateOrderInput, idempotencyKey?: string) {
    const config = idempotencyKey
      ? { headers: { "Idempotency-Key": idempotencyKey } }
      : undefined;
    return http.post<Order>("/orders", input, config);
  },

  getOrder(id: number) {
    return http.get<Order>(`/orders/${id}`);
  },

  cancelOrder(id: number) {
    return http.post<Order>(`/orders/${id}/cancel`);
  },
};

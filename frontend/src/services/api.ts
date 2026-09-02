import axios from "axios";

const api = axios.create({
  baseURL: "/api/v1",
});

api.interceptors.request.use((config) => {
  const token = localStorage.getItem("token");
  if (token) config.headers.Authorization = `Bearer ${token}`;
  return config;
});

export default api;

// Auth
export const authApi = {
  nonce: (address: string) => api.get("/auth/nonce", { params: { address } }),
  walletVerify: (data: { address: string; message: string; signature: string }) =>
    api.post("/auth/wallet/verify", data),
  walletLink: (data: { address: string; message: string; signature: string }) =>
    api.post("/auth/wallet/link", data),
  register: (data: { username: string; email: string; wallet_address: string }) =>
    api.post("/auth/register", data),
  login: (data: { wallet_address: string }) => api.post("/auth/login", data),
};

// Strategies
export const strategyApi = {
  list: () => api.get("/strategies"),
  get: (id: string) => api.get(`/strategies/${id}`),
  create: (data: { name: string; description?: string; natural_language?: string; config?: unknown }) =>
    api.post("/strategies", data),
  update: (id: string, data: Record<string, unknown>) => api.put(`/strategies/${id}`, data),
  remove: (id: string) => api.delete(`/strategies/${id}`),
  parse: (id: string, text: string) => api.post(`/strategies/${id}/parse`, { text }),
  deploy: (id: string, data: { mode?: string; initial_capital?: number }) =>
    api.post(`/strategies/${id}/deploy`, data),
  pause: (id: string) => api.post(`/strategies/${id}/pause`),
  stop: (id: string) => api.post(`/strategies/${id}/stop`),
};

// Backtests
export const backtestApi = {
  create: (data: { strategy_id: string; start_date?: string; end_date?: string; initial_capital?: number }) =>
    api.post("/backtests", data),
  list: (strategyId?: string) =>
    api.get("/backtests", { params: strategyId ? { strategy_id: strategyId } : {} }),
  get: (id: string) => api.get(`/backtests/${id}`),
};

// Portfolio
export const portfolioApi = {
  summary: () => api.get("/portfolio/summary"),
  positions: () => api.get("/portfolio/positions"),
  closePosition: (id: string) => api.post(`/portfolio/positions/${id}/close`),
};

// Events — network-aware (testnet | mainnet)
export const eventApi = {
  list: (params?: Record<string, string>) => {
    const network = localStorage.getItem("network") || "testnet";
    return api.get("/events", { params: { network, ...params } });
  },
  get: (id: string) => api.get(`/events/${id}`),
  prices: (id: string, params?: Record<string, string>) => api.get(`/events/${id}/prices`, { params }),
};

// Alerts
export const alertApi = {
  list: () => api.get("/alerts"),
  create: (data: { alert_type: string; strategy_id?: string; condition: unknown }) =>
    api.post("/alerts", data),
  remove: (id: string) => api.delete(`/alerts/${id}`),
};

// Settings
export const settingsApi = {
  get: () => api.get("/settings"),
  update: (data: { ai_provider?: string; ai_model?: string; api_key?: string; network?: string }) =>
    api.put("/settings", data),
};

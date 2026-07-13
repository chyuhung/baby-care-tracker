import axios from 'axios'
import { useAuthStore } from '@/stores/auth'

const api = axios.create({
  baseURL: '/api',
  timeout: 10000,
  headers: { 'Content-Type': 'application/json' },
})

// 请求拦截器：附加 JWT
api.interceptors.request.use((config) => {
  const auth = useAuthStore()
  if (auth.token) {
    config.headers.Authorization = `Bearer ${auth.token}`
  }
  return config
})

// 响应拦截器
api.interceptors.response.use(
  (res) => res,
  (err) => {
    if (err.response?.status === 401) {
      const auth = useAuthStore()
      auth.logout()
      window.location.href = '/login'
    }
    return Promise.reject(err)
  }
)

// 认证
export const authAPI = {
  register: (username: string, password: string) =>
    api.post('/auth/register', { username, password }),
  login: (username: string, password: string) =>
    api.post('/auth/login', { username, password }),
  getMe: () => api.get('/me'),
}

// 宝宝
export const babyAPI = {
  list: () => api.get('/babies'),
  create: (data: any) => api.post('/babies', data),
  update: (id: number, data: any) => api.put(`/babies/${id}`, data),
  delete: (id: number) => api.delete(`/babies/${id}`),
  stats: (id: number) => api.get(`/babies/${id}/stats`),
  trend: (id: number) => api.get(`/babies/${id}/trend`),
  latestFeeding: (id: number) => api.get(`/babies/${id}/latest-feeding`),
}

// 记录
export const recordAPI = {
  list: (babyId: number, type?: string) =>
    api.get(`/babies/${babyId}/records`, { params: type ? { type } : {} }),
  createFeeding: (babyId: number, data: any) =>
    api.post(`/babies/${babyId}/feeding`, data),
  createDiaper: (babyId: number, data: any) =>
    api.post(`/babies/${babyId}/diaper`, data),
  update: (id: number, type: string, data: any) =>
    api.put(`/records/${id}?type=${type}`, data),
  delete: (id: number, type: string) =>
    api.delete(`/records/${id}?type=${type}`),
}

export default api

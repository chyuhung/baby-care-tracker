import axios from 'axios'
import { useAuthStore } from '@/stores/auth'

export interface Baby {
  id: number
  user_id: number
  name: string
  birth_date: string
  gender: string
  avatar_color: string
  created_at: string
}

export interface FeedingRecord {
  id: number
  baby_id: number
  user_id: number
  type: string
  duration_minutes: number
  amount_ml: number
  side: string
  brand: string
  note: string
  occurred_at: string
  created_at: string
}

export interface DiaperRecord {
  id: number
  baby_id: number
  user_id: number
  type: string
  note: string
  occurred_at: string
  created_at: string
}

export interface SleepRecord {
  id: number
  baby_id: number
  user_id: number
  started_at: string
  ended_at: string | null
  note: string
  created_at: string
}

export interface TemperatureRecord {
  id: number
  baby_id: number
  user_id: number
  temperature: number
  location: string
  note: string
  occurred_at: string
  created_at: string
}

export interface Record {
  id: number
  baby_id: number
  user_id: number
  record_type: string
  data: FeedingRecord | DiaperRecord | SleepRecord | TemperatureRecord
  occurred_at: string
  created_at: string
}

export interface BabyStats {
  feeding_count: number
  diaper_count: number
  last_feeding: string
  last_diaper: string
  total_ml_today: number
  sleep_count: number
  sleep_duration: number
  last_sleep_end: string
  temperature_count: number
  latest_temperature: number
  last_temperature: string
}

export interface DailyStats {
  date: string
  feeding_count: number
  diaper_count: number
  total_ml: number
  sleep_duration_minutes: number
  temperature_avg: number
  temperature_high: number
}

export interface CreateBabyData {
  name: string
  birth_date: string
  gender: string
  avatar_color: string
}

export interface CreateFeedingData {
  type: string
  duration_minutes: number
  amount_ml: number
  side: string
  brand: string
  note: string
  occurred_at: string
}

export interface CreateDiaperData {
  type: string
  note: string
  occurred_at: string
}

export interface UpdateRecordData {
  occurred_at?: string
  type?: string
  duration_minutes?: number
  amount_ml?: number
  side?: string
  brand?: string
  note?: string
  started_at?: string
  ended_at?: string
  temperature?: number
  location?: string
}

export interface CreateSleepData {
  started_at: string
  note?: string
}

export interface CreateTemperatureData {
  temperature: number
  location?: string
  note?: string
  occurred_at: string
}

const api = axios.create({
  baseURL: '/api',
  timeout: 10000,
  headers: { 'Content-Type': 'application/json' },
})

api.interceptors.request.use((config) => {
  const auth = useAuthStore()
  if (auth.token) {
    config.headers.Authorization = `Bearer ${auth.token}`
  }
  config.headers['X-Timezone-Offset'] = String(-new Date().getTimezoneOffset())
  return config
})

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

export const authAPI = {
  register: (username: string, password: string) =>
    api.post('/auth/register', { username, password }),
  login: (username: string, password: string) =>
    api.post('/auth/login', { username, password }),
  getMe: () => api.get('/me'),
}

export const babyAPI = {
  list: () => api.get<Baby[]>('/babies'),
  create: (data: CreateBabyData) => api.post<Baby>('/babies', data),
  update: (id: number, data: Partial<CreateBabyData>) => api.put<Baby>(`/babies/${id}`, data),
  delete: (id: number) => api.delete(`/babies/${id}`),
  stats: (id: number) => api.get<BabyStats>(`/babies/${id}/stats`),
  trend: (id: number, days?: number) => {
    const params = days ? { days } : {}
    return api.get<DailyStats[]>(`/babies/${id}/trend`, { params })
  },
  latestFeeding: (id: number) => api.get(`/babies/${id}/latest-feeding`),
}

export const recordAPI = {
  list: (babyId: number, type?: string, days?: number) => {
    const params: Record<string, string | number> = {}
    if (type) params.type = type
    if (days) params.days = days
    return api.get<Record[]>(`/babies/${babyId}/records`, { params })
  },
  count: (babyId: number) =>
    api.get<{ feeding_count: number; diaper_count: number; sleep_count: number; temperature_count: number; total: number }>(`/babies/${babyId}/records/count`),
  createFeeding: (babyId: number, data: CreateFeedingData) =>
    api.post<Record>(`/babies/${babyId}/feeding`, data),
  createDiaper: (babyId: number, data: CreateDiaperData) =>
    api.post<Record>(`/babies/${babyId}/diaper`, data),
  createSleepStart: (babyId: number, data: CreateSleepData) =>
    api.post<Record>(`/babies/${babyId}/sleep/start`, data),
  stopSleep: (babyId: number, sleepId: number, data: { ended_at: string; note?: string }) =>
    api.put<Record>(`/babies/${babyId}/sleep/${sleepId}/stop`, data),
  getCurrentSleep: (babyId: number) =>
    api.get<SleepRecord | Record<string, never>>(`/babies/${babyId}/sleep/current`),
  createTemperature: (babyId: number, data: CreateTemperatureData) =>
    api.post<Record>(`/babies/${babyId}/temperature`, data),
  update: (id: number, type: string, data: UpdateRecordData) =>
    api.put(`/records/${id}?type=${type}`, data),
  delete: (id: number, type: string) =>
    api.delete(`/records/${id}?type=${type}`),
}

export const familyAPI = {
  getMyFamily: () => api.get('/family'),
  join: (inviteCode: string) => api.post('/family/join', { invite_code: inviteCode }),
  leave: () => api.post('/family/leave'),
  regenerateCode: () => api.post('/family/regenerate-code'),
}

export default api

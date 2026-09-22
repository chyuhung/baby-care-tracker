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

export interface OutdoorRecord {
  id: number
  baby_id: number
  user_id: number
  started_at: string
  ended_at: string | null
  note: string
  created_at: string
}

export interface SupplementRecord {
  id: number
  baby_id: number
  user_id: number
  name: string
  dosage_value: number
  dosage_unit: string
  note: string
  occurred_at: string
  created_at: string
}

export interface Record {
  id: number
  baby_id: number
  user_id: number
  record_type: string
  data: FeedingRecord | DiaperRecord | SleepRecord | TemperatureRecord | OutdoorRecord | SupplementRecord
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
  outdoor_count: number
  outdoor_duration: number
  last_outdoor_end: string
  supplement_count: number
  last_supplement: string
}

export interface DailyStats {
  date: string
  feeding_count: number
  diaper_count: number
  total_ml: number
  sleep_duration_minutes: number
  temperature_avg: number
  temperature_high: number
  outdoor_duration_minutes: number
  supplement_count: number
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
  name?: string
  dosage_value?: number
  dosage_unit?: string
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

export interface CreateOutdoorData {
  started_at: string
  note?: string
}

export interface CreateSupplementData {
  name: string
  dosage_value?: number
  dosage_unit?: string
  note?: string
  occurred_at: string
}

export interface GrowthRecord {
  id: number
  baby_id: number
  user_id: number
  measured_at: string
  weight_kg: number
  height_cm: number
  head_cm: number
  note: string
  created_at: string
}

export interface GrowthStats {
  empty?: boolean
  age_months?: number
  gender?: string
  weight_kg?: number
  height_cm?: number
  head_cm?: number
  weight_pct?: number
  height_pct?: number
  head_pct?: number
  weight_z?: number
  height_z?: number
  head_z?: number
}

export interface CreateGrowthData {
  measured_at: string
  weight_kg: number
  height_cm: number
  head_cm: number
  note?: string
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
  get: (id: number) => api.get<Baby>(`/babies/${id}`),
  create: (data: CreateBabyData) => api.post<Baby>('/babies', data),
  update: (id: number, data: Partial<CreateBabyData>) => api.put<Baby>(`/babies/${id}`, data),
  delete: (id: number) => api.delete(`/babies/${id}`),
  stats: (id: number) => api.get<BabyStats>(`/babies/${id}/stats`),
  trend: (id: number, days?: number) => {
    const params = days ? { days } : {}
    return api.get<DailyStats[]>(`/babies/${id}/trend`, { params })
  },
  latestFeeding: (id: number) => api.get<FeedingRecord>(`/babies/${id}/latest-feeding`),
  latestTemperature: (id: number) => api.get<{ temperature: number; location: string; note: string }>(`/babies/${id}/latest-temperature`),
  latestSupplement: (id: number) => api.get<{ name: string; dosage_value: number; dosage_unit: string; note: string }>(`/babies/${id}/latest-supplement`),
  growth: (id: number) => api.get<GrowthRecord[]>(`/babies/${id}/growth`),
  growthStats: (id: number) => api.get<GrowthStats>(`/babies/${id}/growth/stats`),
  createGrowth: (id: number, data: CreateGrowthData) => api.post<{ id: number }>(`/babies/${id}/growth`, data),
  deleteGrowth: (id: number) => api.delete(`/growth/${id}`),
  exportUrl: (id: number, days?: number) => {
    const q = days ? `?days=${days}` : ''
    return `/api/babies/${id}/export${q}`
  },
}

export const recordAPI = {
  list: (babyId: number, type?: string, days?: number) => {
    const params: Record<string, string | number> = {}
    if (type) params.type = type
    if (days) params.days = days
    return api.get<Record[]>(`/babies/${babyId}/records`, { params })
  },
  count: (babyId: number) =>
    api.get<{ feeding_count: number; diaper_count: number; sleep_count: number; temperature_count: number; outdoor_count: number; supplement_count: number; total: number }>(`/babies/${babyId}/records/count`),
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
  createSupplement: (babyId: number, data: CreateSupplementData) =>
    api.post<Record>(`/babies/${babyId}/supplement`, data),
  createOutdoorStart: (babyId: number, data: CreateOutdoorData) =>
    api.post<Record>(`/babies/${babyId}/outdoor/start`, data),
  stopOutdoor: (babyId: number, outdoorId: number, data: { ended_at: string; note?: string }) =>
    api.put<Record>(`/babies/${babyId}/outdoor/${outdoorId}/stop`, data),
  getCurrentOutdoor: (babyId: number) =>
    api.get<OutdoorRecord | Record<string, never>>(`/babies/${babyId}/outdoor/current`),
  exportRecords: async (babyId: number, days?: number) => {
    const params: Record<string, string | number> = {}
    if (days) params.days = days
    return api.get(`/babies/${babyId}/export`, { params, responseType: 'blob' })
  },
  update: (id: number, type: string, data: UpdateRecordData) =>
    api.put<Record>(`/records/${id}?type=${type}`, data),
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

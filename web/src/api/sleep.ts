// web/src/api/sleep.ts
import api from './axios'

export interface SleepRecord {
  id: string
  child_id: string
  start_time: string
  end_time: string | null
  woke_up: boolean
  wake_count: number
  created_by: string
  created_at: string
}

export const listSleep = (childId: string) =>
  api.get<SleepRecord[]>(`/api/children/${childId}/sleep`).then(r => r.data)

export const createSleep = (
  childId: string,
  data: {
    start_time: string
    end_time?: string
    woke_up: boolean
    wake_count: number
  },
) =>
  api.post<SleepRecord>(`/api/children/${childId}/sleep`, data).then(r => r.data)

export const updateSleep = (
  childId: string,
  sid: string,
  data: {
    start_time: string
    end_time?: string
    woke_up: boolean
    wake_count: number
  },
) =>
  api.put<SleepRecord>(`/api/children/${childId}/sleep/${sid}`, data).then(r => r.data)

export const deleteSleep = (childId: string, sid: string) =>
  api.delete(`/api/children/${childId}/sleep/${sid}`)

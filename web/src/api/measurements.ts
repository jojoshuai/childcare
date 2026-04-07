// web/src/api/measurements.ts
import api from './axios'

export interface Measurement {
  id: string
  child_id: string
  type: 'weight' | 'height' | 'head_circumference'
  value: number
  measured_at: string
  note: string | null
  created_by: string
  created_at: string
}

export const listMeasurements = (childId: string, type?: string) =>
  api
    .get<Measurement[]>(`/api/children/${childId}/measurements`, {
      params: type ? { type } : {},
    })
    .then(r => r.data)

export const createMeasurement = (
  childId: string,
  data: {
    type: string
    value: number
    measured_at: string
    note: string | null
  },
) =>
  api
    .post<Measurement>(`/api/children/${childId}/measurements`, data)
    .then(r => r.data)

export const updateMeasurement = (
  childId: string,
  mid: string,
  data: {
    type: string
    value: number
    measured_at: string
    note: string | null
  },
) =>
  api
    .put<Measurement>(`/api/children/${childId}/measurements/${mid}`, data)
    .then(r => r.data)

export const deleteMeasurement = (childId: string, mid: string) =>
  api.delete(`/api/children/${childId}/measurements/${mid}`)

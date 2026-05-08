// web/src/api/supplement.ts
import api from './axios'

export interface SupplementRecord {
  id: string
  child_id: string
  supplement_name: string
  dose: string | null
  taken_at: string
  created_by: string
  created_at: string
}

export const listSupplements = (childId: string) =>
  api.get<SupplementRecord[]>(`/api/children/${childId}/supplements`).then(r => r.data)

export const createSupplement = (
  childId: string,
  data: {
    supplement_name: string
    dose?: string | null
    taken_at: string
  },
) =>
  api.post<SupplementRecord>(`/api/children/${childId}/supplements`, data).then(r => r.data)

export const updateSupplement = (
  childId: string,
  spid: string,
  data: {
    supplement_name: string
    dose?: string | null
    taken_at: string
  },
) =>
  api.put<SupplementRecord>(`/api/children/${childId}/supplements/${spid}`, data).then(r => r.data)

export const deleteSupplement = (childId: string, spid: string) =>
  api.delete(`/api/children/${childId}/supplements/${spid}`)

export const getSupplementNames = (childId: string) =>
  api.get<string[]>(`/api/children/${childId}/supplements/names`).then(r => r.data)

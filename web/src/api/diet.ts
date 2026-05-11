// web/src/api/diet.ts
import api from './axios'

export interface DietRecord {
  id: string
  child_id: string
  food_name: string
  food_type: string
  amount_level: number
  record_time: string
  meal_group_id: string | null
  meal_type: string
  notes: string | null
  created_by: string
  created_at: string
}

export interface FoodType {
  value: string
  label: string
}

export const listDiet = (childId: string) =>
  api.get<DietRecord[]>(`/api/children/${childId}/diet`).then(r => r.data)

export const createDiet = (
  childId: string,
  data: {
    food_name: string
    food_type: string
    amount_level: number
    record_time: string
    meal_group_id?: string | null
    meal_type?: string
    notes?: string | null
  },
) =>
  api.post<DietRecord>(`/api/children/${childId}/diet`, data).then(r => r.data)

export const updateDiet = (
  childId: string,
  did: string,
  data: {
    food_name: string
    food_type: string
    amount_level: number
    record_time: string
    meal_group_id?: string | null
    meal_type?: string
    notes?: string | null
  },
) =>
  api.put<DietRecord>(`/api/children/${childId}/diet/${did}`, data).then(r => r.data)

export const deleteDiet = (childId: string, did: string) =>
  api.delete(`/api/children/${childId}/diet/${did}`)

export const getFoodTypes = (childId: string) =>
  api.get<FoodType[]>(`/api/children/${childId}/diet/types`).then(r => r.data)

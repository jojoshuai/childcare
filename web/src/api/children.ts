// web/src/api/children.ts
import api from './axios'

export interface Child {
  id: string
  name: string
  gender: 'male' | 'female'
  birth_date: string
  created_at: string
}

export const listChildren = () =>
  api.get<Child[]>('/api/children').then(r => r.data)

export const createChild = (data: {
  name: string
  gender: string
  birth_date: string
}) => api.post<Child>('/api/children', data).then(r => r.data)

export const deleteChild = (id: string) =>
  api.delete(`/api/children/${id}`)

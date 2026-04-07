// web/src/api/family.ts
import api from './axios'

export interface FamilyMember {
  id: string
  nickname: string
  role: string
}

export interface FamilyInfo {
  id: string
  name: string
  members: FamilyMember[]
}

export const getFamily = () =>
  api.get<FamilyInfo>('/api/family').then(r => r.data)

export const generateInvite = () =>
  api
    .post<{ code: string; expires_at: string }>('/api/family/invite')
    .then(r => r.data)

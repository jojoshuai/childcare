// web/src/api/who.ts
import api from './axios'

export interface WHOPoint {
  month: number
  p3: number
  p50: number
  p97: number
}

export const getWHOStandards = (gender: string, type: string) =>
  api
    .get<{ data: WHOPoint[] }>('/api/who-standards', {
      params: { gender, type },
    })
    .then(r => r.data.data)

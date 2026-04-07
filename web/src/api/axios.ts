// web/src/api/axios.ts
import axios from 'axios'

const api = axios.create({ baseURL: '' })

// 自动附加 token
api.interceptors.request.use(config => {
  const token = localStorage.getItem('token')
  if (token) config.headers.Authorization = `Bearer ${token}`
  return config
})

// 401 自动刷新 token
let refreshing = false
let queue: Array<(token: string) => void> = []

api.interceptors.response.use(
  res => res,
  async err => {
    const orig = err.config
    if (err.response?.status !== 401 || orig._retry) {
      return Promise.reject(err)
    }
    orig._retry = true

    if (refreshing) {
      return new Promise(resolve => {
        queue.push(token => {
          orig.headers.Authorization = `Bearer ${token}`
          resolve(api(orig))
        })
      })
    }

    refreshing = true
    try {
      const refreshToken = localStorage.getItem('refresh_token')
      if (!refreshToken) throw new Error('no refresh token')
      const { data } = await axios.post('/api/auth/refresh', {
        refresh_token: refreshToken,
      })
      localStorage.setItem('token', data.token)
      localStorage.setItem('refresh_token', data.refresh_token)
      queue.forEach(cb => cb(data.token))
      queue = []
      orig.headers.Authorization = `Bearer ${data.token}`
      return api(orig)
    } catch {
      localStorage.clear()
      window.location.href = '/login'
      return Promise.reject(err)
    } finally {
      refreshing = false
    }
  },
)

export default api

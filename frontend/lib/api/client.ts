import axios, { AxiosInstance } from 'axios'

const API_URL = process.env.NEXT_PUBLIC_API_URL || 'http://localhost:8080'

class APIClient {
  private client: AxiosInstance

  constructor() {
    this.client = axios.create({
      baseURL: `${API_URL}/api/v1`,
      headers: {
        'Content-Type': 'application/json',
      },
    })

    // Add auth token to requests
    this.client.interceptors.request.use((config) => {
      if (typeof window !== 'undefined') {
        const token = localStorage.getItem('token')
        if (token) {
          config.headers.Authorization = `Bearer ${token}`
        }
      }
      return config
    })

    // Handle 401 errors
    this.client.interceptors.response.use(
      (response) => response,
      (error) => {
        if (error.response?.status === 401 && typeof window !== 'undefined') {
          // Only redirect if we're not already on an auth page
          const currentPath = window.location.pathname
          const isAuthPage = currentPath === '/login' || currentPath === '/register'
          
          if (!isAuthPage) {
            // User's token expired, redirect to login
            localStorage.removeItem('token')
            localStorage.removeItem('user')
            window.location.href = '/login'
          }
          // If we're already on auth page, just let the error propagate normally
        }
        return Promise.reject(error)
      }
    )
  }

  getClient() {
    return this.client
  }
}

export const apiClient = new APIClient().getClient()
export default apiClient


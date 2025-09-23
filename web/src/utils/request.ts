import http from './http'

// 统一的请求响应接口
export interface ApiResponse<T = any> {
  code: number
  message: string
  data: T
}

// 请求工具类
export const request = {
  get: <T = any>(url: string, config?: any): Promise<ApiResponse<T>> => {
    return http.get(url, config)
  },

  post: <T = any>(url: string, data?: any, config?: any): Promise<ApiResponse<T>> => {
    return http.post(url, data, config)
  },

  put: <T = any>(url: string, data?: any, config?: any): Promise<ApiResponse<T>> => {
    return http.put(url, data, config)
  },

  delete: <T = any>(url: string, config?: any): Promise<ApiResponse<T>> => {
    return http.delete(url, config)
  },

  patch: <T = any>(url: string, data?: any, config?: any): Promise<ApiResponse<T>> => {
    return http.patch(url, data, config)
  }
}

export default request
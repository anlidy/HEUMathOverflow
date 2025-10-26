// 登录表单数据类型
export interface LoginForm {
  username: string
  password: string
}

// 登录接口响应数据类型
export interface LoginResponse {
  code: number
  message: string
  data: {
    token: string
    userInfo: {
      id: number
      username: string
      email: string
      role: 'student' | 'teacher' | 'admin'
      avatar?: string
    }
  }
}

// 用户信息类型
export interface UserInfo {
  id: number
  username: string
  email: string
  role: 'student' | 'teacher' | 'admin'
  avatar?: string
}

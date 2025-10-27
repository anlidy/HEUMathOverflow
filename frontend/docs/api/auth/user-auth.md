# 用户认证模块 API 文档

## 1. 用户登录

### 接口信息

- **URL**: `POST /api/v1/auth/login`
- **Content-Type**: `application/json`
- **描述**: 用户登录接口，登录成功后通过Set-Cookie设置sessionID

### Request 类型

```typescript
interface LoginRequest {
  email: string;        // 用户邮箱
  password: string;     // 用户密码
  remember?: boolean;   // 是否记住登录状态，可选
}
```

### Response 类型

```typescript
interface LoginResponse {
  code: number;         // 状态码，200表示成功
  message: string;      // 响应消息
  data: {
    userInfo: {
      id: number;       // 用户ID
      username: string; // 用户名
      email: string;    // 用户邮箱
      role: 'student' | 'teacher' | 'admin'; // 用户角色
      avatar?: string;  // 用户头像URL，可选
    }
  }
}
```

## 2. 用户注册

### 接口信息

- **URL**: `POST /api/v1/auth/register`
- **Content-Type**: `application/json`
- **描述**: 用户注册接口，注册成功后直接进入主页

### Request 类型

```typescript
interface RegisterRequest {
  username: string;     // 用户名
  email: string;        // 用户邮箱
  password: string;     // 用户密码
  confirmPassword: string; // 确认密码
}
```

### Response 类型

```typescript
interface RegisterResponse {
  code: number;         // 状态码，200表示成功
  message: string;      // 响应消息
  data: {
    userInfo: {
      id: number;       // 用户ID
      username: string; // 用户名
      email: string;    // 用户邮箱
      role: 'student';  // 默认角色为学生
      avatar?: string;  // 用户头像URL，默认为空
    }
  }
}
```

## 3. 用户登出

### 接口信息

- **URL**: `POST /api/v1/auth/logout`
- **Content-Type**: `application/json`
- **描述**: 用户登出接口，清除服务端session

### Request 类型

```typescript
// 无需请求体，通过Cookie中的sessionID识别用户
```

### Response 类型

```typescript
interface LogoutResponse {
  code: number;         // 状态码，200表示成功
  message: string;      // 响应消息
  data: null;
}
```

## 4. 获取当前用户信息

### 接口信息

- **URL**: `GET /api/v1/auth/me`
- **Content-Type**: `application/json`
- **描述**: 获取当前登录用户的信息

### Request 类型

```typescript
// 无需请求体，通过Cookie中的sessionID识别用户
```

### Response 类型

```typescript
interface GetUserInfoResponse {
  code: number;         // 状态码，200表示成功
  message: string;      // 响应消息
  data: {
    userInfo: {
      id: number;       // 用户ID
      username: string; // 用户名
      email: string;    // 用户邮箱
      role: 'student' | 'teacher' | 'admin'; // 用户角色
      avatar?: string;  // 用户头像URL
      createdAt: string; // 注册时间
      lastLoginAt: string; // 最后登录时间
    }
  }
}
```

## 5. 修改密码

### 接口信息

- **URL**: `PUT /api/v1/auth/password`
- **Content-Type**: `application/json`
- **描述**: 修改用户密码

### Request 类型

```typescript
interface ChangePasswordRequest {
  oldPassword: string;  // 原密码
  newPassword: string;  // 新密码
  confirmPassword: string; // 确认新密码
}
```

### Response 类型

```typescript
interface ChangePasswordResponse {
  code: number;         // 状态码，200表示成功
  message: string;      // 响应消息
  data: null;
}
```

## 6. 重置密码

### 接口信息

- **URL**: `POST /api/v1/auth/reset-password`
- **Content-Type**: `application/json`
- **描述**: 发送密码重置邮件

### Request 类型

```typescript
interface ResetPasswordRequest {
  email: string;        // 用户邮箱
}
```

### Response 类型

```typescript
interface ResetPasswordResponse {
  code: number;         // 状态码，200表示成功
  message: string;      // 响应消息
  data: null;
}
```

## 7. 确认重置密码

### 接口信息

- **URL**: `POST /api/v1/auth/confirm-reset-password`
- **Content-Type**: `application/json`
- **描述**: 通过重置令牌确认重置密码

### Request 类型

```typescript
interface ConfirmResetPasswordRequest {
  token: string;        // 重置令牌
  newPassword: string;  // 新密码
  confirmPassword: string; // 确认新密码
}
```

### Response 类型

```typescript
interface ConfirmResetPasswordResponse {
  code: number;         // 状态码，200表示成功
  message: string;      // 响应消息
  data: null;
}
```

## 8. 上传用户头像

### 接口信息

- **URL**: `POST /api/v1/auth/avatar`
- **Content-Type**: `multipart/form-data`
- **描述**: 上传用户头像

### Request 类型

```typescript
// FormData格式
interface UploadAvatarRequest {
  avatar: File;         // 头像文件
}
```

### Response 类型

```typescript
interface UploadAvatarResponse {
  code: number;         // 状态码，200表示成功
  message: string;      // 响应消息
  data: {
    avatarUrl: string;  // 头像URL
  }
}
```

## 错误码说明

| 错误码 | 说明               |
| ------ | ------------------ |
| 200    | 成功               |
| 400    | 请求参数错误       |
| 401    | 未授权，需要登录   |
| 403    | 禁止访问，权限不足 |
| 404    | 资源不存在         |
| 409    | 冲突，如邮箱已存在 |
| 422    | 参数验证失败       |
| 500    | 服务器内部错误     |

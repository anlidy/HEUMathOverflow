# 登录认证模块文档

## 1. 界面设计

## 2. 主要功能

- 用户登录/注册
- 权限控制
- 会话管理
- 记住登录状态

## 3. 需要的API 接口

### 3.1 登录接口

登录成功后端会通过Set-Cookie为请求添加sessionID字段, 无需携带token.

```ts
POST /api/auth/login
Content-Type: application/json

Request:
{
  "email": string,
  "password": string
  "remember":bool
}

Response:
{
  "code": number,
  "message":string,
  "data": {
    "userInfo": {
      "id": number,
      "username": string,
      "role": string,
      "avatar_url":string
    }
  }
}
```

### 3.2 注册接口

注册成功后直接进入主页, 无需重新登录

```ts
POST /api/auth/register
Content-Type: application/json

Request:
{
  "username": string,
  "password": string,
  "email": string
}

Response:
{
  "code": number,
  "message": string,
  "data": {
    "userInfo": {
      "id": number,
      "username": string,
      "role": string, // 默认角色为学生
      "avatar_url":string // 默认为空,前端可以显示一个默认头像
    }
  }
}

```

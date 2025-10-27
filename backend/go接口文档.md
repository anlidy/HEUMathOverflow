#### 一、UserService：用户相关服务模块

#### 根路径: /api/v1/user

#### 1.1. 用户登录

- 请求方法: POST
- 相对路径: /login

登录成功后端会通过Set-Cookie为请求添加sessionID字段, 无需携带token.

```ts
POST /api/v1/user/login
Content-Type: application/json

Request:
{
  "email": string,
  "password": string
  "remember":bool
}

Response:
{
  "code": int,
  "message":string,
  "data": {
    "userInfo": {
      "id": int,
      "username": string,
      "role": string,
      "avatar_url":string 
    }
  }
}
```

#### 1.2. 用户注册

- 请求方法: POST
- 相对路径: /register

注册成功后直接进入主页, 无需重新登录

```ts
POST /api/v1/user/register
Content-Type: application/json

Request:
{
  "username": string,
  "password": string,
  "email": string
}

Response:
{
  "code": int,
  "message": string,
  "data": {
    "userInfo": {
      "id": int,
      "username": string,
      "role": string,  // 默认角色为学生
      "avatar_url":string // 默认为空,前端可以显示一个默认头像
    }
}
```

#### 1.3. 上传用户头像

- 请求方法: POST
- 相对路径: /avatar

```ts
POST /api/v1/user/avatar
Content-Type: multipart/form-data

file: (binary)

Response:
{
  "code": int,
  "message": string,
  "data": {
    "avatar_url": string
  }
}
```

#### 1.4.显示用户头像

- 请求方法: GET

可直接内嵌链接到网页中, 浏览器进行请求

```ts
GET /api/v1/user/avatar/{filename}	// 请求链接即为avatar_url

Response:
Content-Type: image/png  // 允许jpg,jpeg,png,bmp格式
Content-Length: 123456
Content-Disposition: inline; filename="20251027.png"
```

#### 1.5. 修改用户信息

- 请求方法: PATCH
- 相对路径: /profile

```ts
PATCH /api/v1/user/profile
Content-Type: application/json

Request:
{
  "username": string,
  ...	// 后续可添加字段
}
  
Response:
{
  "code": int,
  "message": string
}
```

#### 1.6. 修改用户密码 (需验证旧密码)

- 请求方法: PATCH
- 相对路径: /password

```ts
PATCH /api/v1/user/password
Content-Type: application/json

Request:
{
  "old_password": string,
  "new_password": string
}

Response:
{
  "code": int,
  "message": string
}
```


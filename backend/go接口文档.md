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
    "user_info": {
      "id": string,
      "username": string,
      "role": int,   // 1 -> 学生, 2 -> "助教", 3 -> "教师", 4 ->"管理员"
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
    "user_info": {
      "id": string,
      "username": string,
      "role": int,  // 默认角色为学生
      "avatar_url":string // 默认为空,前端可以显示一个默认头像
    }
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


#### ForumService：论坛相关服务模块

#### 根路径: /api/v1/forum

#### 2.1. 上传文件

- 请求方法: POST
- 相对路径: /upload

包括图片、语音等文件都可使用此接口

```ts
POST /api/v1/forum/upload
Content-Type: multipart/form-data

file: (binary)

Response:
{
    "code": int,
    "data": {
        "image_url": string
    },
    "message": string
}
```

#### 2.2.下载文件

- 请求方法: GET

```ts
GET /api/v1/forum/file/{filename}	 // 请求链接即为image_url

Response:
Content-Type: image/png  // 允许（正则式） jpg|jpeg|png|bmp|mp3|docx?|pptx?|xlsx?|pdf|zip|rar
Content-Length: 123456
Content-Disposition: inline; filename="20251027.png" //格式为inline，前端来决定是展示还是下载
```

#### 2.3. 创建帖子

- 请求方法: POST
- 相对路径: /posts

```ts
POST /api/v1/forum/posts
Content-Type: application/json

Request:
{
    "title":string,
    "content":string,
    "image_urls":string[],
    "tags":string[]
}

Response:
{
    "code": int,
    "data": {
        "post_id": string
    },
    "message": string
}
```

#### 2.4. 获取单条帖子

- 请求方法: GET
- 相对路径: /posts/{post_id}

```ts
GET /api/v1/forum/posts/{post_id}

Response:
{
    "code": int,
    "data": {
        "post_data": {  // 帖子数据
            "post_id": string,
            "title": string,
            "tags": string[],
            "status": int,  // 1:未解决 2:已解决 3:已认证
            "last_reply_at": string | null, // "2025-11-20T12:12:40.33807Z"
            "created_at": string,
            "updated_at": string,
            "content": string,
            "image_urls": string[]
        },
        "user_info": {  // 发帖人的用户信息
            "user_id": string,
            "username": string,
            "role": int,
            "avatar_url": string
        }
    },
    "message": string
}
```

#### 2.5. 创建回复(评论)

- 请求方法: POST
- 相对路径: /replies

```ts
POST /api/v1/forum/replies
Content-Type: application/json

Request:
{
    "post_id":string,
    "parent_reply_id":string|null,
    "content":string,
    "voice_url":string,
    "image_urls":string[]
}

Response:
{
    "code": int,
    "data": {
        "reply_id": string
    },
    "message": string
}
```

#### 2.4. 获取帖子下的回复

- 请求方法: GET
- 相对路径: /posts/{postID}/replies

```ts
// offset: 偏移量
// limit: 数量
GET /api/v1/forum/posts/{postID}/replies?offset=0&limit=10  

Response:
{
    "code": 200,
    "data": [
        {
            "user_info": {  // 回帖人信息
                "user_id": string,
                "username": string,
                "role": int,
                "avatar_url": string
            },
            "reply_data": { // 回帖数据
                "reply_id": string,
                "post_id": string,
                "parent_reply_id": string | null,
                "status": int,  // 1:未精选 2:作者精选 3:教师精选
                "certified_by": string | null,
                "created_at": string,
                "content": string,
                "voice_url": string,
                "voice_text": string,
                "image_urls": string[],
                "ai_answered": bool
            }
        },
        ...
    ],
    "message": string
}
```

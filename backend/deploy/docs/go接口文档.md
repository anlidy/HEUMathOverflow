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
GET /api/v1/user/avatar/{filename} // 请求链接即为avatar_url

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
  ... // 后续可添加字段
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

#### 1.7. 退出登录

- 请求方法: POST
- 相对路径: /logout

需要已登录状态, 后端从 Cookie 中读取 `session-id` 并清除会话与 Cookie。

```ts
POST /api/v1/user/logout

Response:
{
  "code": int,
  "message": string
}
```

#### 1.8. 注销账号

- 请求方法: DELETE
- 相对路径: /account

需要已登录状态, 并校验当前账号密码, 注销成功后会同时删除账号与当前登录会话。

```ts
DELETE /api/v1/user/account
Content-Type: application/json

Request:
{
  "password": string
}

Response:
{
  "code": int,
  "message": string
}
```

### ForumService：论坛相关服务模块

#### 根路径: /api/v1/forum

#### 2.1.1 上传文件

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

#### 2.1.2下载文件

- 请求方法: GET

```ts
GET /api/v1/forum/file/{filename}  // 请求链接即为image_url

Response:
Content-Type: image/png  // 允许（正则式） jpg|jpeg|png|bmp|mp3|docx?|pptx?|xlsx?|pdf|zip|rar
Content-Length: 123456
Content-Disposition: inline; filename="20251027.png" //格式为inline，前端来决定是展示还是下载
```

#### 2.2.1 创建帖子

- 请求方法: POST
- 相对路径: /posts
- 前置步骤: 进入发帖页面时先获取一次 `client_token`（临时 token，不需要持久化存储）

```ts
GET /api/v1/forum/posts/client-token

Response:
{
  "code": int,
  "data": { "client_token": string },
  "message": string
}
```

```ts
POST /api/v1/forum/posts
Content-Type: application/json

Request:
{
    "client_token": string,
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

备注:
- `client_token` 在有效期内用于防重复提交：重复发帖会直接返回首次创建的 `post_id`
- 若返回 `202 Accepted` 表示首次请求仍在处理中，前端可携带同一个 `client_token` 重试
- 若返回 `410 Gone` 表示 token 过期，前端需重新获取 `client_token`
```

#### 2.2.2 获取单条帖子

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
            "content": string,
            "image_urls": string[],
            "tags": string[],
            "status": int,  // 1:未解决 2:已解决 3:已认证
            "views": int, // 浏览量
            "likes": int, // 点赞数
            "stars": int, // 收藏数
            "replies": int, // 回复数
            "last_reply_at": string | null, // "2025-11-20T12:12:40.33807Z"
            "created_at": string,
            "updated_at": string,
            "liked":bool, // 当前用户是否点赞该贴
            "starred":bool, // 当前用户是否收藏该贴
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

#### 2.2.3 获取多条帖子

- 请求方法: GET
- 相对路径: /posts

```ts
// page: 页码 (默认:1)
// page_size: 每页帖子数(默认 20)
// order: 排序方式 (0:推荐 1:最热 2:最新),默认值:0
GET /api/v1/forum/posts?page=1&page_size=20&order=0  

// 此data字段返回一个post列表,每个元素的内容与2.2.2的data字段一致
Response:
{
    "code": 200,
    "data": [
      {
        "post_data": {  // 帖子数据
            "post_id": string,
            "title": string,
            "content": string,
            "image_urls": string[],
            "tags": string[],
            "status": int,  // 1:未解决 2:已解决 3:已认证
            "views": int, // 浏览量
            "likes": int, // 点赞数
            "stars": int, // 收藏数
            "replies": int, // 回复数
            "last_reply_at": string | null, // "2025-11-20T12:12:40.33807Z"
            "created_at": string,
            "updated_at": string,
            // 不显示状态,进入详情页再单独查询一次帖子数据
        },
        "user_info": {  // 发帖人的用户信息
            "user_id": string,
            "username": string,
            "role": int,
            "avatar_url": string
        }
      },
        ...
    ],
    "pagination":{   // COUNT(*) 非常耗时,不返回total帖子总数
        "page":int,
        "page_size":int    
    },
    "message": string
}
```

#### 2.2.4 更新帖子内容

- 请求方法: PATCH
- 相对路径: /posts/{post_id}

```ts
PATCH /api/v1/forum/posts/{post_id}
Content-Type: application/json

Request:
{
    "title": string,  // 必选字段
    "content": string,  // 必选字段
    "tags": string[]  // 必选字段,置空会覆盖之前的tags
    "add_image_urls": string[], // 可选字段: 新增的图片url
    "delete_image_urls": string[],  // 可选字段: 要删除的图片url
}

Response:
{
    "code": int,
    "message": string
}
```

#### 2.2.5 删除一条帖子

- 请求方法: DELETE
- 相对路径: /posts/{post_id}

```ts
DELETE /api/v1/forum/posts/{post_id}

Response:
{
    "code": int,
    "message": string
}
```

#### 2.2.6 帖子点赞

- 请求方法: POST
- 相对路径: /posts/like/{post_id}

```ts
POST /api/v1/forum/posts/like/{post_id}
Content-Type: application/json

Response:
{
    "code": int,
    "message": string
}
```

#### 2.2.7 取消帖子点赞

- 请求方法: DELETE
- 相对路径: /posts/like/{post_id}

```ts
DELETE /api/v1/forum/posts/like/{post_id}
Content-Type: application/json

Response:
{
    "code": int,
    "message": string
}
```

#### 2.2.8 收藏帖子

- 请求方法: POST
- 相对路径: /posts/star/{post_id}

```ts
POST /api/v1/forum/posts/star/{post_id}
Content-Type: application/json

Response:
{
    "code": int,
    "message": string
}
```

#### 2.2.9 取消收藏

- 请求方法: DELETE
- 相对路径: /posts/star/{post_id}

```ts
DELETE /api/v1/forum/posts/star/{post_id}
Content-Type: application/json
Response:
{
    "code": int,
    "message": string
}
```

#### 2.2.10 查询当前用户收藏的帖子（分页）

- 请求方法: GET
- 相对路径: /posts/starred

```ts
// page: 页码（默认 1）
// page_size: 每页帖子数(默认 20)
GET /api/v1/forum/posts/starred?page=1&page_size=20

Response:
{
    "code": 200,
    "data": [
        {
            "post_data": {  // 帖子数据
                "post_id": string,
                "title": string,
                "content": string,
                "image_urls": string[],
                "tags": string[],
                "status": int,
                "views": int,
                "likes": int,
                "stars": int,
                "replies": int,
                "last_reply_at": string | null,
                "created_at": string,
                "updated_at": string,
                // 进入详情页再查询一次帖子获取状态
            },
            "user_info": {  // 发帖人的用户信息
                "user_id": string,
                "username": string,
                "role": int,
                "avatar_url": string
            }
        },
        ...
    ],
    "pagination":{
        "page":int,
        "page_size":int,
        "total":int
    },
    "message": string
}
```

#### 2.3.1 创建回复(评论)

- 请求方法: POST
- 相对路径: /replies
- 前置步骤: 进入回复页面时先获取一次 `client_token`（临时 token，不需要持久化存储）

```ts
GET /api/v1/forum/replies/client-token

Response:
{
  "code": int,
  "data": { "client_token": string },
  "message": string
}
```

```ts
POST /api/v1/forum/replies
Content-Type: application/json

Request:
{
    "client_token": string,
    "post_id":string,
    "parent_reply_id":string|null,    // 为空代表回复一条帖子,不为空代表评论一条回复
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

备注:
- `client_token` 在有效期内用于防重复提交：重复回帖会直接返回首次创建的 `reply_id`
- 若返回 `202 Accepted` 表示首次请求仍在处理中，前端可携带同一个 `client_token` 重试
- 若返回 `410 Gone` 表示 token 过期，前端需重新获取 `client_token`
```

#### 2.3.2 获取帖子下的回复

- 请求方法: GET
- 相对路径: /posts/{postID}/replies

```ts
// page: 页码
// page_size: 每页帖子数
GET /api/v1/forum/posts/{postID}/replies?page=1&page_size=20  

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
                "likes": int,   // 点赞数
                "certified_by": string | null,
                "content": string,
                "image_urls": string[],
                "voice_url": string,
                "voice_text": string,
                "ai_answered": bool,
                "created_at": string,
                "liked":bool, // 当前用户是否点赞该回复
            }
        },
        ...
    ],
    "pagination":{
        "page":int,
        "page_size":int,
        "total":int,
    },
    "message": string
}
```

#### 2.3.3 给回复点赞

- 请求方法: POST
- 相对路径: /replies/like/{reply_id}

```ts
POST /api/v1/forum/replies/like/{reply_id}
Content-Type: application/json


Response:
{
    "code": int,
    "message": string
}
```

#### 2.3.4 取消回复点赞

- 请求方法: DELETE
- 相对路径: /replies/like/{reply_id}

```ts
DELETE /api/v1/forum/replies/like/{reply_id}
Content-Type: application/json

Response:
{
    "code": int,
    "message": string
}
```

#### 2.4. 帖子搜索

- 请求方法: POST
- 相对路径: /search

```ts
POST /api/v1/forum/search

Request:
{
    "query":"test2",
    "tags":[],
    "page":1,   // 页码,从1开始
    "page_size":20, // 每页帖子数
    "sort":1    // 1:默认排序 2:热度高 3:新发布 4:浏览多 5:评论多
}

Response:
{
    "code": 200,
    "data": [
        {
            "user_info": {
                "user_id": string,
                "username": string,
                "role": int,
                "avatar_url": string
            },
            "post_data": {
                "post_id": string,
                "title": string,
                "content": string,
                "image_urls": string[],
                "tags": string[],
                "status": int,
                "views": int,
                "likes": int,
                "stars": int,
                "replies": int,
                "last_reply_at": string | null,
                "created_at": string,
                "updated_at": string,
            }
        }
    ],
    "message": string,
    "pagination":{
        "page": int,
        "page_size": int,
        "total": int,
    },
}
```

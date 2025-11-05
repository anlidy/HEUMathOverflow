#### 一、数据库字段设计

#### 1. PostgreSQL

##### 1.1. user_db库：用户与会话管理

##### `users`表: 存储用户的基本信息

| 字段名        | 类型               | 说明                                                    |
| ------------- | ------------------ | ------------------------------------------------------- |
| id            | BIGINT PRIMARY KEY | 用户ID, 使用雪花算法生成                                |
| email         | VARCHAR(100)       | 邮箱（唯一）                                            |
| username      | VARCHAR(50)        | 用户名（唯一）                                          |
| password_hash | TEXT               | 加密密码                                                |
| role          | INT        | 枚举值：1: `student`  2: `assistant`   3: `teacher`  4: `admin` |
| avatar_url    | TEXT               | 用户头像url, 文件存入MinOS                              |
| created_at    | TIMESTAMP          | 注册时间                                                |
| last_login    | TIMESTAMP          | 上次登录时间                                            |

#####  `sessions`表：存储登录的会话记录

**已迁移到Redis数据库**, 采用sessionID -> {userID, roleKey,remember}和roleKey->role键值映射, 并根据以下情况设置过期时间:

-  前端remember为true, 设置30天过期, 每当用户活动则重置为30天;
-  前端remember为false 或 新用户注册, 设置1天过期, 每当用户活动则重置为1天.

##### 1.2. forum_db库: 帖子与回复索引

- MongoDB 保存 “正文” 信息，这里只保存结构化“ 索引” 和元信息, 便于首页展示和排序

##### `posts`表：存储原帖的基本信息

| 字段名        | 类型         | 说明                                                    |
| ------------- | ------------ | ------------------------------------------------------- |
| id            | BIGSERIAL PK | 帖子ID                                                  |
| author_id     | BIGINT       | 发帖人ID                                                |
| title         | VARCHAR(255) | 帖子标题                                                |
| tags          | TEXT[]       | 标签数组（如 ['微积分', '难题']）                       |
| status        | INT          | 帖子状态, 枚举值：1:`未解答`  2:`已解答` 3:  `教师认证` |
| doc_id        | CHAR(24)     | 对应 MongoDB 文档的 ObjectID                            |
| last_reply_at | TIMESTAMP    | 最后回复时间                                            |
| created_at    | TIMESTAMP    | 发布时间                                                |
| updated_at    | TIMESTAMP    | 更新时间                                                |

##### `replies`表：存储回帖的基本信息

| 字段名          | 类型         | 说明                                                       |
| --------------- | ------------ | ---------------------------------------------------------- |
| id              | BIGSERIAL PK | 回复ID                                                     |
| post_id         | BIGINT       | 所属原贴的ID                                               |
| replier_id      | BIGINT       | 回复者ID（学生/AI/教师/管理员）                            |
| parent_reply_id | BIGINT NULL  | 若为回复他人评论，则指向该评论ID；否则为 NULL              |
| mongo_doc_id    | CHAR(24)     | 回复内容在 MongoDB 中的文档ID                              |
| status          | INT          | 评论状态,枚举值: 1: `未被精选` 2: `作者精选` 3: `教师精选` |
| certified_by    | BIGINT NULL  | 精选该评论的教师ID（可空）                                 |
| created_at      | TIMESTAMP    | 回复时间                                                   |
| updated_at      | TIMESTAMP    | 更新时间                                                   |

##### 1.3. audit_db库: 内容审核和日志

##### `audit_logs`表：

| 字段名      | 类型        | 说明                                                         |
| ----------- | ----------- | ------------------------------------------------------------ |
| id          | BIGINT PK   | 主键, 雪花算法生成                                           |
| action      | VARCHAR(20) | 操作类型：`post_banned`(内容违规) , `post_useless`(无关水贴) |
| target_id   | BIGINT      | 被操作对象ID（原贴ID或回帖ID）                               |
| author_id   | BIGINT      | 发帖人ID                                                     |
| reason      | TEXT        | 审核原因                                                     |
| operator_id | BIGINT      | 操作人ID（AI助教/教师/管理员）                               |
| created_at  | TIMESTAMP   | 创建时间                                                     |

##### 1.4. teacher_db库: 教师认证与高质量问答存档

##### `certified_answers`表：

| 字段名       | 类型      | 说明               |
| ------------ | --------- | ------------------ |
| id           | BIGINT PK | 主键, 雪花算法生成 |
| post_id      | BIGINT    | 帖子ID             |
| reply_id     | BIGINT    | 回复ID             |
| teacher_id   | BIGINT    | 认证教师ID         |
| certified_at | TIMESTAMP | 认证时间           |

#### 2. MongoDB

**2.1 数据库：`forum_content`**

##### 集合：`posts`

- `_id`: 自动生成的内部id
- `post_id`: 原贴的id
- `content`: 纯文本内容
- `images`: 原贴附带的图片, 只保存url, 文件存入MinOS

```json
{
  "_id": ObjectId("..."),
  "post_id": 1001,
  "content": "<p>请问这道极限题如何计算？</p>",
  "images": [
    "https://cdn.math-overflow.edu/forum/1001/image1.jpg"
  ]
}
```

##### 集合：`replies`

- `_id`: 自动生成的内部id
- `reply_id`: 回复id
- `parent_reply_id`: 若回复他人评论，则指向该评论ID; 否则为null
- `content`: 纯文本内容
- `voice`:  用户发的语音, 只存储url
- `voice_text`: 用户语音转成的文字
- `images`: 回贴附带的图片, 只保存url, 文件存入MinOS
- `ai_answered`: 是否为Ai生成的回答
- `teracher_answered`: 是否为老师的回答

```json
{
  "_id": ObjectId("..."),
  "reply_id": 2005,
  "parent_reply_id":2001,	// 可空字段
  "content": "<p>这道题可以使用洛必达法则...</p>",
  "voice": "https://cdn.math-overflow.edu/forum/2005/voice1.mp3",
  "voice_text":"这道题可以使用洛必达法则...",
  "images": [],
  "ai_answered": false,
  "teacher_answered": true
}
```

#### 


#### 一、数据库字段设计

#### 1. PostgreSQL

##### 1.1. `user_db`：用户与会话管理

`users` 表（使用雪花 ID，非自增）：

| 字段名        | 类型               | 说明                                                    |
| ------------- | ------------------ | ------------------------------------------------------- |
| id            | BIGINT PK          | 用户ID，雪花算法生成                                    |
| email         | VARCHAR(100)       | 邮箱（唯一）                                            |
| username      | VARCHAR(50)        | 用户名（唯一）                                          |
| password_hash | TEXT               | 加密密码                                                |
| role          | INT                | 枚举：1 student / 2 assistant / 3 teacher / 4 admin     |
| avatar_url    | TEXT               | 头像 URL（文件存 MinIO）                                |
| last_login    | TIMESTAMP          | 上次登录时间                                            |
| created_at    | TIMESTAMP          | 创建时间                                                |
| updated_at    | TIMESTAMP          | 更新时间                                                |

会话数据已迁移至 Redis（sessionID → {userID, role, remember}），过期策略：
- remember=true：30 天，活跃即续期；
- remember=false：1 天，活跃即续期。

##### 1.2. `forum_db`：帖子与回复

`posts` 表（雪花 ID，非自增，正文直接存 PG）：

| 字段名        | 类型                | 说明                                                    |
| ------------- | ------------------- | ------------------------------------------------------- |
| id            | BIGINT PK           | 帖子ID                                                  |
| author_id     | BIGINT              | 作者ID                                                  |
| title         | VARCHAR(255)        | 标题                                                    |
| content       | TEXT                | 正文                                                    |
| image_urls    | VARCHAR(128)[]      | 图片 URL 数组（MinIO）                                  |
| tags          | VARCHAR(32)[]       | 标签数组                                                |
| status        | INT                 | 1 未解答 / 2 已解答 / 3 教师认证                        |
| views         | INT                 | 浏览数                                                  |
| likes         | INT                 | 点赞数（冗余计数）                                      |
| stars         | INT                 | 收藏数（冗余计数）                                      |
| replies       | INT                 | 回复数（冗余计数）                                      |
| last_reply_at | TIMESTAMP NULL      | 最后回复时间                                            |
| created_at    | TIMESTAMP           | 创建时间                                                |
| updated_at    | TIMESTAMP           | 更新时间                                                |

`post_likes` 表：用户对帖子的点赞

| 字段名   | 类型      | 说明              |
| -------- | --------- | ----------------- |
| post_id  | BIGINT PK | 帖子ID            |
| user_id  | BIGINT PK | 用户ID            |
| created_at | TIMESTAMP | 点赞时间        |

`post_stars` 表：用户对帖子的收藏

| 字段名   | 类型      | 说明              |
| -------- | --------- | ----------------- |
| post_id  | BIGINT PK | 帖子ID            |
| user_id  | BIGINT PK | 用户ID            |
| created_at | TIMESTAMP | 收藏时间        |

`replies` 表（雪花 ID，非自增）：

| 字段名         | 类型            | 说明                                                         |
| -------------- | --------------- | ------------------------------------------------------------ |
| id             | BIGINT PK       | 回复ID                                                      |
| post_id        | BIGINT          | 所属帖子ID（级联删除）                                       |
| replier_id     | BIGINT          | 回复者ID                                                     |
| content        | TEXT            | 回复正文                                                     |
| status         | INT             | 1 未被精选 / 2 作者精选 / 3 教师精选                         |
| likes          | INT             | 点赞数（冗余计数）                                           |
| image_urls     | VARCHAR(128)[]  | 图片URL数组（MinIO）                                         |
| voice_url      | TEXT            | 语音URL（MinIO，可空）                                      |
| voice_text     | TEXT            | 语音转文字（可空）                                           |
| ai_answered    | BOOL            | 是否AI回答                                                   |
| parent_reply_id| BIGINT NULL     | 父回复ID（级联置空）                                         |
| certified_by   | BIGINT NULL     | 认证教师ID（可空）                                           |
| created_at     | TIMESTAMP       | 创建时间                                                     |
| updated_at     | TIMESTAMP       | 更新时间                                                     |

`reply_likes` 表：用户对回复的点赞

| 字段名  | 类型      | 说明      |
| ------- | --------- | --------- |
| reply_id| BIGINT PK | 回复ID    |
| user_id | BIGINT PK | 用户ID    |
| created_at | TIMESTAMP | 点赞时间 |

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

##### `certified_answers`表

| 字段名       | 类型      | 说明               |
| ------------ | --------- | ------------------ |
| id           | BIGINT PK | 主键, 雪花算法生成 |
| post_id      | BIGINT    | 帖子ID             |
| reply_id     | BIGINT    | 回复ID             |
| teacher_id   | BIGINT    | 认证教师ID         |
| certified_at | TIMESTAMP | 认证时间           |

#### 2. Redis

- 存储登录 Session；键：sessionID，值：{userID, role, remember, TTL}。

#### 3. MinIO

- 存储用户头像、帖子/回复图片、语音等文件，URL 持久化在 PG。

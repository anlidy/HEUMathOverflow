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

##### 2.1. forum-service 缓存与计数

forum 当前只保留 Redis 缓存，不再使用进程内 LRU。本层 Redis 主要承担两类职责：

- 读缓存：减少热点帖子详情、热点帖子首页回复列表的数据库读取压力。
- 计数与异步写回：承接浏览、点赞、收藏、回复数等高频增量，再由后台 worker 定时刷回 PostgreSQL。

##### 2.2. 帖子详情缓存

用途：

- 给热点帖子详情接口提供缓存命中。
- 避免同一时间大量请求同时穿透到数据库。

主要键：

- `post:{postID}`：帖子详情缓存，值为帖子 JSON。
- `lock:post:{postID}`：回源加载帖子详情时使用的分布式锁。
- `post:visit:{postID}`：短时间访问计数，用于判断帖子是否属于热点。
- `post:hot:{yyyyMMddHH}`：按小时划分的热点 ZSet 桶，用于统计 24 小时内的热门帖子。

工作原理：

1. 访问帖子详情时，先读取 `post:{postID}`。
2. 若命中，直接返回缓存内容。
3. 若未命中，先通过进程内 `singleflight` 合并同实例内重复请求。
4. 随后尝试获取 `lock:post:{postID}` 分布式锁，避免多实例同时回源查询同一帖子。
5. 拿到锁的请求负责查 PostgreSQL，并在帖子被判定为热点时写入 `post:{postID}`。
6. 没拿到锁的请求会短暂等待其他请求回填缓存；若在等待窗口内仍未命中，再自行回源查询。

热点判定：

- 每次帖子详情访问都会增加 `post:visit:{postID}` 计数，并写入当前小时的 `post:hot:{yyyyMMddHH}` 桶。
- `post:visit:{postID}` 使用短 TTL 窗口判断“最近是否足够热”。
- 热点预热 worker 会定期聚合近 24 小时热点桶，挑出仍然处于热点窗口内的帖子，提前把详情写入 Redis。

##### 2.3. 回复列表与回复详情缓存

用途：

- 缓存热点帖子首页回复列表，减少第一页高频访问时的数据库压力。
- 缓存回复详情，避免已经拿到回复 ID 后再次逐条查库。

主要键：

- `comments:post:{postID}`：热点帖首页回复列表缓存，值为 `{ids, total}` JSON。
- `comment:{replyID}`：单条回复详情缓存，值为回复 JSON。
- `lock:comments:post:{postID}`：回源加载首页回复列表时使用的分布式锁。

工作原理：

1. 当前只对“热点帖子 + 第 1 页 + page_size=20”的回复列表启用缓存。
2. 读取首页回复时先查 `comments:post:{postID}`，一次拿到回复 ID 列表和总数 `total`。
3. 命中后，再通过 `MGET comment:{replyID}` 批量读取回复详情。
4. 若部分回复详情缺失，仅对缺失的 replyID 回源数据库，再回填对应 `comment:{replyID}`。
5. 列表缓存未命中时，同样先走 `singleflight`，再尝试获取 `lock:comments:post:{postID}` 分布式锁。
6. 没拿到锁的请求会短暂等待其他请求回填列表缓存，超过等待窗口才自行回源。

一致性策略：

- 回复列表缓存将 `ids` 和 `total` 存在同一个 key 中，避免“列表命中但总数丢失”的不一致问题。
- 回复新增、更新、删除后会主动失效对应帖子下的列表缓存，以及相关回复详情缓存。

##### 2.4. 计数缓存与异步写回

用途：

- 把浏览、点赞、收藏、回复数等高频小增量先落到 Redis，降低数据库直接写压力。
- 由后台 worker 周期性批量写回 PostgreSQL，再继续向下游发送所需事件。

主要键：

- `post:{postID}:stat`：帖子统计增量。
- `post:{postID}:stat:processing`：帖子统计刷库时的处理中 key。
- `reply:{replyID}:stat`：回复统计增量，目前主要用于回复点赞数。
- `reply:{replyID}:stat:processing`：回复统计刷库时的处理中 key。

工作原理：

1. 用户点赞、收藏、浏览、回复等操作先写 Redis 统计 key，而不是立即更新 PostgreSQL 冗余计数字段。
2. worker 扫描统计 key 后，会先把 `:stat` 原子搬运到 `:stat:processing`，冻结这一批增量。
3. worker 读取冻结快照并写回 PostgreSQL 中的冗余计数字段，例如：

- `posts.views`
- `posts.likes`
- `posts.stars`
- `posts.replies`
- `replies.likes`

4. 数据库写回成功后删除 `:processing` key，失败则保留，等待后续重试或排查。

这样做的作用：

- 降低高频计数直接写库造成的行竞争。
- 让接口主链路更轻，优先保证写操作响应速度。
- 为后续统一事件投递、搜索同步提供更稳定的统计快照来源。

#### 3. MinIO

- 存储用户头像、帖子/回复图片、语音等文件，URL 持久化在 PG。

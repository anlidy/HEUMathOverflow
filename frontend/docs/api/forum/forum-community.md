# 论坛社区模块 API 文档

## 1. 帖子管理

### 1.1 创建帖子

#### 接口信息

- **URL**: `POST /api/v1/forum/posts`
- **Content-Type**: `application/json`
- **描述**: 创建新的问题帖子

#### Request 类型

```typescript
interface CreatePostRequest {
  title: string;        // 帖子标题
  content: string;      // 帖子内容（支持LaTeX数学公式）
  tags: string[];       // 标签数组，如["微积分", "线性代数", "难题"]
  category?: string;    // 分类，可选
  isAnonymous?: boolean; // 是否匿名发布，可选，默认false
}
```

#### Response 类型

```typescript
interface CreatePostResponse {
  code: number;         // 状态码，200表示成功
  message: string;      // 响应消息
  data: {
    post: {
      id: number;       // 帖子ID
      title: string;    // 帖子标题
      content: string;  // 帖子内容
      tags: string[];   // 标签数组
      category: string; // 分类
      author: {
        id: number;     // 作者ID
        username: string; // 作者用户名
        avatar?: string; // 作者头像
        role: string;   // 作者角色
      };
      status: 'pending' | 'approved' | 'rejected'; // 审核状态
      isAnonymous: boolean; // 是否匿名
      createdAt: string; // 创建时间
      updatedAt: string; // 更新时间
    }
  }
}
```

### 1.2 获取帖子列表

#### 接口信息

- **URL**: `GET /api/v1/forum/posts`
- **Content-Type**: `application/json`
- **描述**: 获取帖子列表，支持分页和筛选

#### Request 类型

```typescript
// Query参数
interface GetPostsRequest {
  page?: number;        // 页码，默认1
  limit?: number;       // 每页数量，默认20
  tags?: string[];      // 标签筛选，多个用逗号分隔
  category?: string;    // 分类筛选
  status?: 'pending' | 'approved' | 'rejected' | 'all'; // 状态筛选
  sortBy?: 'createdAt' | 'updatedAt' | 'replyCount' | 'likeCount'; // 排序字段
  sortOrder?: 'asc' | 'desc'; // 排序方向
  search?: string;      // 搜索关键词
}
```

#### Response 类型

```typescript
interface GetPostsResponse {
  code: number;         // 状态码，200表示成功
  message: string;      // 响应消息
  data: {
    posts: {
      id: number;       // 帖子ID
      title: string;    // 帖子标题
      content: string;  // 帖子内容（截取前200字符）
      tags: string[];   // 标签数组
      category: string; // 分类
      author: {
        id: number;     // 作者ID
        username: string; // 作者用户名
        avatar?: string; // 作者头像
        role: string;   // 作者角色
      };
      status: string;   // 审核状态
      isAnonymous: boolean; // 是否匿名
      replyCount: number; // 回复数量
      likeCount: number; // 点赞数量
      viewCount: number; // 浏览数量
      hasCertifiedAnswer: boolean; // 是否有认证答案
      createdAt: string; // 创建时间
      updatedAt: string; // 更新时间
    }[];
    pagination: {
      page: number;     // 当前页码
      limit: number;    // 每页数量
      total: number;    // 总数量
      totalPages: number; // 总页数
    }
  }
}
```

### 1.3 获取帖子详情

#### 接口信息

- **URL**: `GET /api/v1/forum/posts/{id}`
- **Content-Type**: `application/json`
- **描述**: 获取指定帖子的详细信息

#### Request 类型

```typescript
// 路径参数
interface GetPostDetailRequest {
  id: number;           // 帖子ID
}
```

#### Response 类型

```typescript
interface GetPostDetailResponse {
  code: number;         // 状态码，200表示成功
  message: string;      // 响应消息
  data: {
    post: {
      id: number;       // 帖子ID
      title: string;    // 帖子标题
      content: string;  // 帖子内容
      tags: string[];   // 标签数组
      category: string; // 分类
      author: {
        id: number;     // 作者ID
        username: string; // 作者用户名
        avatar?: string; // 作者头像
        role: string;   // 作者角色
      };
      status: string;   // 审核状态
      isAnonymous: boolean; // 是否匿名
      replyCount: number; // 回复数量
      likeCount: number; // 点赞数量
      viewCount: number; // 浏览数量
      hasCertifiedAnswer: boolean; // 是否有认证答案
      isLiked: boolean; // 当前用户是否已点赞
      isBookmarked: boolean; // 当前用户是否已收藏
      createdAt: string; // 创建时间
      updatedAt: string; // 更新时间
    }
  }
}
```

### 1.4 更新帖子

#### 接口信息

- **URL**: `PUT /api/v1/forum/posts/{id}`
- **Content-Type**: `application/json`
- **描述**: 更新帖子内容（仅作者可操作）

#### Request 类型

```typescript
interface UpdatePostRequest {
  title?: string;       // 帖子标题
  content?: string;     // 帖子内容
  tags?: string[];      // 标签数组
  category?: string;    // 分类
}
```

#### Response 类型

```typescript
interface UpdatePostResponse {
  code: number;         // 状态码，200表示成功
  message: string;      // 响应消息
  data: {
    post: {
      id: number;       // 帖子ID
      title: string;    // 帖子标题
      content: string;  // 帖子内容
      tags: string[];   // 标签数组
      category: string; // 分类
      updatedAt: string; // 更新时间
    }
  }
}
```

### 1.5 删除帖子

#### 接口信息

- **URL**: `DELETE /api/v1/forum/posts/{id}`
- **Content-Type**: `application/json`
- **描述**: 删除帖子（仅作者或管理员可操作）

#### Request 类型

```typescript
// 路径参数
interface DeletePostRequest {
  id: number;           // 帖子ID
}
```

#### Response 类型

```typescript
interface DeletePostResponse {
  code: number;         // 状态码，200表示成功
  message: string;      // 响应消息
  data: null;
}
```

## 2. 回复管理

### 2.1 创建回复

#### 接口信息

- **URL**: `POST /api/v1/forum/posts/{postId}/replies`
- **Content-Type**: `application/json`
- **描述**: 对帖子进行回复

#### Request 类型

```typescript
interface CreateReplyRequest {
  content: string;      // 回复内容（支持LaTeX数学公式）
  parentId?: number;    // 父回复ID，用于嵌套回复
  isAnonymous?: boolean; // 是否匿名回复，可选，默认false
}
```

#### Response 类型

```typescript
interface CreateReplyResponse {
  code: number;         // 状态码，200表示成功
  message: string;      // 响应消息
  data: {
    reply: {
      id: number;       // 回复ID
      content: string;  // 回复内容
      author: {
        id: number;     // 作者ID
        username: string; // 作者用户名
        avatar?: string; // 作者头像
        role: string;   // 作者角色
      };
      parentId?: number; // 父回复ID
      postId: number;   // 所属帖子ID
      status: 'pending' | 'approved' | 'rejected'; // 审核状态
      isAnonymous: boolean; // 是否匿名
      isCertified: boolean; // 是否被认证为优质答案
      likeCount: number; // 点赞数量
      isLiked: boolean; // 当前用户是否已点赞
      createdAt: string; // 创建时间
      updatedAt: string; // 更新时间
    }
  }
}
```

### 2.2 获取回复列表

#### 接口信息

- **URL**: `GET /api/v1/forum/posts/{postId}/replies`
- **Content-Type**: `application/json`
- **描述**: 获取指定帖子的回复列表

#### Request 类型

```typescript
// Query参数
interface GetRepliesRequest {
  postId: number;       // 帖子ID（路径参数）
  page?: number;        // 页码，默认1
  limit?: number;       // 每页数量，默认20
  sortBy?: 'createdAt' | 'likeCount'; // 排序字段
  sortOrder?: 'asc' | 'desc'; // 排序方向
  showCertified?: boolean; // 是否只显示认证回复
}
```

#### Response 类型

```typescript
interface GetRepliesResponse {
  code: number;         // 状态码，200表示成功
  message: string;      // 响应消息
  data: {
    replies: {
      id: number;       // 回复ID
      content: string;  // 回复内容
      author: {
        id: number;     // 作者ID
        username: string; // 作者用户名
        avatar?: string; // 作者头像
        role: string;   // 作者角色
      };
      parentId?: number; // 父回复ID
      postId: number;   // 所属帖子ID
      status: string;   // 审核状态
      isAnonymous: boolean; // 是否匿名
      isCertified: boolean; // 是否被认证为优质答案
      likeCount: number; // 点赞数量
      isLiked: boolean; // 当前用户是否已点赞
      createdAt: string; // 创建时间
      updatedAt: string; // 更新时间
    }[];
    pagination: {
      page: number;     // 当前页码
      limit: number;    // 每页数量
      total: number;    // 总数量
      totalPages: number; // 总页数
    }
  }
}
```

### 2.3 更新回复

#### 接口信息

- **URL**: `PUT /api/v1/forum/replies/{id}`
- **Content-Type**: `application/json`
- **描述**: 更新回复内容（仅作者可操作）

#### Request 类型

```typescript
interface UpdateReplyRequest {
  content: string;      // 回复内容
}
```

#### Response 类型

```typescript
interface UpdateReplyResponse {
  code: number;         // 状态码，200表示成功
  message: string;      // 响应消息
  data: {
    reply: {
      id: number;       // 回复ID
      content: string;  // 回复内容
      updatedAt: string; // 更新时间
    }
  }
}
```

### 2.4 删除回复

#### 接口信息

- **URL**: `DELETE /api/v1/forum/replies/{id}`
- **Content-Type**: `application/json`
- **描述**: 删除回复（仅作者或管理员可操作）

#### Request 类型

```typescript
// 路径参数
interface DeleteReplyRequest {
  id: number;           // 回复ID
}
```

#### Response 类型

```typescript
interface DeleteReplyResponse {
  code: number;         // 状态码，200表示成功
  message: string;      // 响应消息
  data: null;
}
```

## 3. 互动功能

### 3.1 点赞/取消点赞

#### 接口信息

- **URL**: `POST /api/v1/forum/{type}/{id}/like`
- **Content-Type**: `application/json`
- **描述**: 对帖子或回复进行点赞/取消点赞

#### Request 类型

```typescript
// 路径参数：type为'posts'或'replies'，id为对应ID
interface LikeRequest {
  // 无需请求体
}
```

#### Response 类型

```typescript
interface LikeResponse {
  code: number;         // 状态码，200表示成功
  message: string;      // 响应消息
  data: {
    isLiked: boolean;   // 当前点赞状态
    likeCount: number;  // 总点赞数
  }
}
```

### 3.2 收藏/取消收藏

#### 接口信息

- **URL**: `POST /api/v1/forum/posts/{id}/bookmark`
- **Content-Type**: `application/json`
- **描述**: 对帖子进行收藏/取消收藏

#### Request 类型

```typescript
// 路径参数
interface BookmarkRequest {
  id: number;           // 帖子ID
}
```

#### Response 类型

```typescript
interface BookmarkResponse {
  code: number;         // 状态码，200表示成功
  message: string;      // 响应消息
  data: {
    isBookmarked: boolean; // 当前收藏状态
  }
}
```

### 3.3 获取用户收藏列表

#### 接口信息

- **URL**: `GET /api/v1/forum/bookmarks`
- **Content-Type**: `application/json`
- **描述**: 获取当前用户的收藏列表

#### Request 类型

```typescript
// Query参数
interface GetBookmarksRequest {
  page?: number;        // 页码，默认1
  limit?: number;       // 每页数量，默认20
}
```

#### Response 类型

```typescript
interface GetBookmarksResponse {
  code: number;         // 状态码，200表示成功
  message: string;      // 响应消息
  data: {
    bookmarks: {
      id: number;       // 收藏ID
      post: {
        id: number;     // 帖子ID
        title: string;  // 帖子标题
        content: string; // 帖子内容（截取）
        tags: string[]; // 标签数组
        author: {
          username: string; // 作者用户名
          avatar?: string; // 作者头像
        };
        replyCount: number; // 回复数量
        likeCount: number; // 点赞数量
        hasCertifiedAnswer: boolean; // 是否有认证答案
        createdAt: string; // 创建时间
      };
      bookmarkedAt: string; // 收藏时间
    }[];
    pagination: {
      page: number;     // 当前页码
      limit: number;    // 每页数量
      total: number;    // 总数量
      totalPages: number; // 总页数
    }
  }
}
```

## 4. 标签管理

### 4.1 获取所有标签

#### 接口信息

- **URL**: `GET /api/v1/forum/tags`
- **Content-Type**: `application/json`
- **描述**: 获取系统中所有可用的标签

#### Request 类型

```typescript
// 无需请求体
```

#### Response 类型

```typescript
interface GetTagsResponse {
  code: number;         // 状态码，200表示成功
  message: string;      // 响应消息
  data: {
    tags: {
      name: string;     // 标签名称
      count: number;    // 使用次数
      color?: string;   // 标签颜色
    }[];
  }
}
```

### 4.2 创建标签

#### 接口信息

- **URL**: `POST /api/v1/forum/tags`
- **Content-Type**: `application/json`
- **描述**: 创建新标签（仅管理员可操作）

#### Request 类型

```typescript
interface CreateTagRequest {
  name: string;         // 标签名称
  color?: string;       // 标签颜色，可选
  description?: string; // 标签描述，可选
}
```

#### Response 类型

```typescript
interface CreateTagResponse {
  code: number;         // 状态码，200表示成功
  message: string;      // 响应消息
  data: {
    tag: {
      name: string;     // 标签名称
      color?: string;   // 标签颜色
      description?: string; // 标签描述
      createdAt: string; // 创建时间
    }
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
| 409    | 冲突，如重复操作   |
| 422    | 参数验证失败       |
| 500    | 服务器内部错误     |

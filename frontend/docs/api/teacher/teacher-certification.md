# 教师认证模块 API 文档

## 1. 答案认证管理

### 1.1 认证优质答案

#### 接口信息

- **URL**: `POST /api/v1/teacher/certify-answer`
- **Content-Type**: `application/json`
- **描述**: 教师认证回复为优质答案

#### Request 类型

```typescript
interface CertifyAnswerRequest {
  replyId: number;      // 回复ID
  postId: number;       // 所属帖子ID
  certificationLevel: 'excellent' | 'good' | 'standard'; // 认证等级
  comment?: string;     // 认证评语，可选
  tags?: string[];      // 标签，如["重点", "难点", "易错点"]
}
```

#### Response 类型

```typescript
interface CertifyAnswerResponse {
  code: number;         // 状态码，200表示成功
  message: string;      // 响应消息
  data: {
    certification: {
      id: number;       // 认证ID
      replyId: number;  // 回复ID
      postId: number;   // 帖子ID
      certifier: {
        id: number;     // 认证教师ID
        username: string; // 认证教师用户名
        role: string;   // 教师角色
      };
      level: string;    // 认证等级
      comment?: string; // 认证评语
      tags: string[];   // 标签
      certifiedAt: string; // 认证时间
      isActive: boolean; // 是否有效
    }
  }
}
```

### 1.2 取消答案认证

#### 接口信息

- **URL**: `DELETE /api/v1/teacher/certify-answer/{certificationId}`
- **Content-Type**: `application/json`
- **描述**: 取消对答案的认证（仅认证教师可操作）

#### Request 类型

```typescript
// 路径参数
interface CancelCertificationRequest {
  certificationId: number; // 认证ID
}
```

#### Response 类型

```typescript
interface CancelCertificationResponse {
  code: number;         // 状态码，200表示成功
  message: string;      // 响应消息
  data: null;
}
```

### 1.3 获取认证答案列表

#### 接口信息

- **URL**: `GET /api/v1/teacher/certified-answers`
- **Content-Type**: `application/json`
- **描述**: 获取所有已认证的优质答案列表

#### Request 类型

```typescript
// Query参数
interface GetCertifiedAnswersRequest {
  page?: number;        // 页码，默认1
  limit?: number;       // 每页数量，默认20
  level?: 'excellent' | 'good' | 'standard' | 'all'; // 认证等级筛选
  tags?: string[];      // 标签筛选
  subject?: string;     // 学科筛选
  certifierId?: number; // 认证教师ID筛选
  startDate?: string;   // 开始日期
  endDate?: string;     // 结束日期
  sortBy?: 'certifiedAt' | 'level' | 'likeCount'; // 排序字段
  sortOrder?: 'asc' | 'desc'; // 排序方向
}
```

#### Response 类型

```typescript
interface GetCertifiedAnswersResponse {
  code: number;         // 状态码，200表示成功
  message: string;      // 响应消息
  data: {
    answers: {
      id: number;       // 认证ID
      reply: {
        id: number;     // 回复ID
        content: string; // 回复内容
        author: {
          id: number;   // 作者ID
          username: string; // 作者用户名
          avatar?: string; // 作者头像
          role: string; // 作者角色
        };
        likeCount: number; // 点赞数
        createdAt: string; // 回复时间
      };
      post: {
        id: number;     // 帖子ID
        title: string;  // 帖子标题
        tags: string[]; // 帖子标签
        category: string; // 帖子分类
      };
      certification: {
        level: string;  // 认证等级
        comment?: string; // 认证评语
        tags: string[]; // 认证标签
        certifier: {
          id: number;   // 认证教师ID
          username: string; // 认证教师用户名
          avatar?: string; // 认证教师头像
        };
        certifiedAt: string; // 认证时间
      };
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

### 1.4 获取单个认证答案详情

#### 接口信息

- **URL**: `GET /api/v1/teacher/certified-answers/{certificationId}`
- **Content-Type**: `application/json`
- **描述**: 获取指定认证答案的详细信息

#### Request 类型

```typescript
// 路径参数
interface GetCertifiedAnswerDetailRequest {
  certificationId: number; // 认证ID
}
```

#### Response 类型

```typescript
interface GetCertifiedAnswerDetailResponse {
  code: number;         // 状态码，200表示成功
  message: string;      // 响应消息
  data: {
    answer: {
      id: number;       // 认证ID
      reply: {
        id: number;     // 回复ID
        content: string; // 回复内容
        author: {
          id: number;   // 作者ID
          username: string; // 作者用户名
          avatar?: string; // 作者头像
          role: string; // 作者角色
        };
        likeCount: number; // 点赞数
        isLiked: boolean; // 当前用户是否已点赞
        createdAt: string; // 回复时间
        updatedAt: string; // 更新时间
      };
      post: {
        id: number;     // 帖子ID
        title: string;  // 帖子标题
        content: string; // 帖子内容
        tags: string[]; // 帖子标签
        category: string; // 帖子分类
        author: {
          id: number;   // 作者ID
          username: string; // 作者用户名
        };
        createdAt: string; // 帖子创建时间
      };
      certification: {
        level: string;  // 认证等级
        comment?: string; // 认证评语
        tags: string[]; // 认证标签
        certifier: {
          id: number;   // 认证教师ID
          username: string; // 认证教师用户名
          avatar?: string; // 认证教师头像
          role: string; // 认证教师角色
        };
        certifiedAt: string; // 认证时间
        isActive: boolean; // 是否有效
      };
    }
  }
}
```

## 2. 认证数据导出

### 2.1 导出认证数据

#### 接口信息

- **URL**: `POST /api/v1/teacher/export-certified-data`
- **Content-Type**: `application/json`
- **描述**: 导出认证的问答数据（仅管理员可操作）

#### Request 类型

```typescript
interface ExportCertifiedDataRequest {
  startDate?: string;   // 开始日期，格式：YYYY-MM-DD
  endDate?: string;     // 结束日期，格式：YYYY-MM-DD
  level?: 'excellent' | 'good' | 'standard' | 'all'; // 认证等级筛选
  tags?: string[];      // 标签筛选
  subject?: string;     // 学科筛选
  format: 'json' | 'csv' | 'excel'; // 导出格式
  includeMetadata?: boolean; // 是否包含元数据
}
```

#### Response 类型

```typescript
interface ExportCertifiedDataResponse {
  code: number;         // 状态码，200表示成功
  message: string;      // 响应消息
  data: {
    exportId: string;   // 导出任务ID
    status: 'processing' | 'completed' | 'failed'; // 导出状态
    downloadUrl?: string; // 下载链接（完成时提供）
    estimatedTime?: number; // 预计完成时间（秒）
    createdAt: string; // 创建时间
  }
}
```

### 2.2 获取导出状态

#### 接口信息

- **URL**: `GET /api/v1/teacher/export/{exportId}/status`
- **Content-Type**: `application/json`
- **描述**: 查询导出任务状态

#### Request 类型

```typescript
// 路径参数
interface GetExportStatusRequest {
  exportId: string;     // 导出任务ID
}
```

#### Response 类型

```typescript
interface GetExportStatusResponse {
  code: number;         // 状态码，200表示成功
  message: string;      // 响应消息
  data: {
    export: {
      id: string;       // 导出任务ID
      status: 'processing' | 'completed' | 'failed'; // 导出状态
      progress: number; // 导出进度（0-100）
      downloadUrl?: string; // 下载链接
      fileSize?: number; // 文件大小（字节）
      recordCount: number; // 记录数量
      createdAt: string; // 创建时间
      completedAt?: string; // 完成时间
      error?: string;   // 错误信息（如果失败）
    }
  }
}
```

### 2.3 获取导出历史

#### 接口信息

- **URL**: `GET /api/v1/teacher/export/history`
- **Content-Type**: `application/json`
- **描述**: 获取导出历史记录（仅管理员可访问）

#### Request 类型

```typescript
// Query参数
interface GetExportHistoryRequest {
  page?: number;        // 页码，默认1
  limit?: number;       // 每页数量，默认20
  startDate?: string;   // 开始日期
  endDate?: string;     // 结束日期
  status?: string;      // 导出状态筛选
}
```

#### Response 类型

```typescript
interface GetExportHistoryResponse {
  code: number;         // 状态码，200表示成功
  message: string;      // 响应消息
  data: {
    exports: {
      id: string;       // 导出任务ID
      status: string;   // 导出状态
      format: string;   // 导出格式
      recordCount: number; // 记录数量
      fileSize?: number; // 文件大小
      downloadUrl?: string; // 下载链接
      filters: {
        startDate?: string; // 筛选条件
        endDate?: string;
        level?: string;
        tags?: string[];
      };
      createdAt: string; // 创建时间
      completedAt?: string; // 完成时间
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

## 3. 教师管理

### 3.1 获取教师列表

#### 接口信息

- **URL**: `GET /api/v1/teacher/teachers`
- **Content-Type**: `application/json`
- **描述**: 获取系统中所有教师列表（仅管理员可访问）

#### Request 类型

```typescript
// Query参数
interface GetTeachersRequest {
  page?: number;        // 页码，默认1
  limit?: number;       // 每页数量，默认20
  role?: 'teacher' | 'admin' | 'all'; // 角色筛选
  status?: 'active' | 'inactive' | 'all'; // 状态筛选
  search?: string;      // 搜索关键词
}
```

#### Response 类型

```typescript
interface GetTeachersResponse {
  code: number;         // 状态码，200表示成功
  message: string;      // 响应消息
  data: {
    teachers: {
      id: number;       // 教师ID
      username: string; // 用户名
      email: string;    // 邮箱
      role: string;     // 角色
      status: string;   // 状态
      avatar?: string;  // 头像
      certificationCount: number; // 认证答案数量
      lastActiveAt: string; // 最后活跃时间
      createdAt: string; // 注册时间
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

### 3.2 更新教师权限

#### 接口信息

- **URL**: `PUT /api/v1/teacher/teachers/{teacherId}/permissions`
- **Content-Type**: `application/json`
- **描述**: 更新教师权限（仅管理员可操作）

#### Request 类型

```typescript
interface UpdateTeacherPermissionsRequest {
  role: 'teacher' | 'admin'; // 教师角色
  permissions: {
    canCertify: boolean; // 是否可以认证答案
    canModerate: boolean; // 是否可以审核内容
    canManageUsers: boolean; // 是否可以管理用户
    canExportData: boolean; // 是否可以导出数据
    canManageSystem: boolean; // 是否可以管理系统
  };
}
```

#### Response 类型

```typescript
interface UpdateTeacherPermissionsResponse {
  code: number;         // 状态码，200表示成功
  message: string;      // 响应消息
  data: {
    teacher: {
      id: number;       // 教师ID
      username: string; // 用户名
      role: string;     // 角色
      permissions: {
        canCertify: boolean;
        canModerate: boolean;
        canManageUsers: boolean;
        canExportData: boolean;
        canManageSystem: boolean;
      };
      updatedAt: string; // 更新时间
    }
  }
}
```

## 4. 认证统计

### 4.1 获取认证统计信息

#### 接口信息

- **URL**: `GET /api/v1/teacher/certification/stats`
- **Content-Type**: `application/json`
- **描述**: 获取认证相关统计信息

#### Request 类型

```typescript
// Query参数
interface GetCertificationStatsRequest {
  startDate?: string;   // 开始日期
  endDate?: string;     // 结束日期
  groupBy?: 'day' | 'week' | 'month'; // 分组方式
  teacherId?: number;   // 教师ID筛选（可选）
}
```

#### Response 类型

```typescript
interface GetCertificationStatsResponse {
  code: number;         // 状态码，200表示成功
  message: string;      // 响应消息
  data: {
    stats: {
      date: string;     // 日期
      totalCertifications: number; // 总认证数
      excellentCount: number; // 优秀认证数
      goodCount: number; // 良好认证数
      standardCount: number; // 标准认证数
      avgCertificationTime: number; // 平均认证时间（小时）
    }[];
    summary: {
      totalCertifications: number; // 总认证数
      totalTeachers: number; // 参与认证的教师数
      avgCertificationsPerTeacher: number; // 平均每教师认证数
      topTags: {
        tag: string;    // 标签
        count: number;  // 数量
        percentage: number; // 占比
      }[];
      topTeachers: {
        teacherId: number; // 教师ID
        username: string; // 教师用户名
        certificationCount: number; // 认证数量
      }[];
    }
  }
}
```

### 4.2 获取教师认证统计

#### 接口信息

- **URL**: `GET /api/v1/teacher/teachers/{teacherId}/certification-stats`
- **Content-Type**: `application/json`
- **描述**: 获取指定教师的认证统计信息

#### Request 类型

```typescript
// 路径参数
interface GetTeacherCertificationStatsRequest {
  teacherId: number;    // 教师ID
  startDate?: string;   // 开始日期
  endDate?: string;     // 结束日期
}
```

#### Response 类型

```typescript
interface GetTeacherCertificationStatsResponse {
  code: number;         // 状态码，200表示成功
  message: string;      // 响应消息
  data: {
    teacher: {
      id: number;       // 教师ID
      username: string; // 教师用户名
      avatar?: string;  // 教师头像
    };
    stats: {
      totalCertifications: number; // 总认证数
      excellentCount: number; // 优秀认证数
      goodCount: number; // 良好认证数
      standardCount: number; // 标准认证数
      avgCertificationTime: number; // 平均认证时间
      recentCertifications: {
        id: number;     // 认证ID
        replyId: number; // 回复ID
        postTitle: string; // 帖子标题
        level: string;  // 认证等级
        certifiedAt: string; // 认证时间
      }[];
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
| 409    | 冲突，如重复认证   |
| 422    | 参数验证失败       |
| 500    | 服务器内部错误     |

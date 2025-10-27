# 内容审核模块 API 文档

## 1. 内容安全审核

### 1.1 自动内容审核

#### 接口信息

- **URL**: `POST /api/v1/content/audit`
- **Content-Type**: `application/json`
- **描述**: 对用户提交的内容进行自动安全审核

#### Request 类型

```typescript
interface ContentAuditRequest {
  content: string;      // 待审核的内容
  contentType: 'post' | 'reply' | 'comment'; // 内容类型
  userId: number;       // 用户ID
  postId?: number;      // 相关帖子ID（如果是回复或评论）
  replyId?: number;     // 相关回复ID（如果是评论）
}
```

#### Response 类型

```typescript
interface ContentAuditResponse {
  code: number;         // 状态码，200表示成功
  message: string;      // 响应消息
  data: {
    auditId: string;    // 审核ID
    status: 'approved' | 'rejected' | 'pending'; // 审核状态
    confidence: number; // 审核置信度（0-1）
    violations: {
      type: 'spam' | 'abuse' | 'hate' | 'violence' | 'adult' | 'illegal'; // 违规类型
      severity: 'low' | 'medium' | 'high'; // 严重程度
      description: string; // 违规描述
      position?: {
        start: number;  // 违规内容开始位置
        end: number;    // 违规内容结束位置
      };
    }[];
    suggestions?: string[]; // 修改建议
    processingTime: number; // 处理时间（毫秒）
    createdAt: string; // 审核时间
  }
}
```

### 1.2 批量内容审核

#### 接口信息

- **URL**: `POST /api/v1/content/audit/batch`
- **Content-Type**: `application/json`
- **描述**: 批量审核多个内容

#### Request 类型

```typescript
interface BatchContentAuditRequest {
  contents: {
    id: string;         // 内容唯一标识
    content: string;    // 待审核的内容
    contentType: 'post' | 'reply' | 'comment'; // 内容类型
    userId: number;     // 用户ID
  }[];
}
```

#### Response 类型

```typescript
interface BatchContentAuditResponse {
  code: number;         // 状态码，200表示成功
  message: string;      // 响应消息
  data: {
    results: {
      id: string;       // 内容标识
      status: 'approved' | 'rejected' | 'pending'; // 审核状态
      confidence: number; // 审核置信度
      violations: {
        type: string;   // 违规类型
        severity: string; // 严重程度
        description: string; // 违规描述
      }[];
      processingTime: number; // 处理时间
    }[];
    summary: {
      total: number;    // 总数量
      approved: number; // 通过数量
      rejected: number; // 拒绝数量
      pending: number;  // 待审核数量
      avgProcessingTime: number; // 平均处理时间
    }
  }
}
```

### 1.3 获取审核结果

#### 接口信息

- **URL**: `GET /api/v1/content/audit/{auditId}`
- **Content-Type**: `application/json`
- **描述**: 获取指定审核的详细结果

#### Request 类型

```typescript
// 路径参数
interface GetAuditResultRequest {
  auditId: string;      // 审核ID
}
```

#### Response 类型

```typescript
interface GetAuditResultResponse {
  code: number;         // 状态码，200表示成功
  message: string;      // 响应消息
  data: {
    audit: {
      id: string;       // 审核ID
      content: string;  // 审核内容
      contentType: string; // 内容类型
      userId: number;   // 用户ID
      status: string;   // 审核状态
      confidence: number; // 审核置信度
      violations: {
        type: string;   // 违规类型
        severity: string; // 严重程度
        description: string; // 违规描述
        position?: {
          start: number; // 违规内容开始位置
          end: number;   // 违规内容结束位置
        };
      }[];
      suggestions?: string[]; // 修改建议
      processingTime: number; // 处理时间
      createdAt: string; // 审核时间
      reviewedAt?: string; // 人工审核时间
      reviewedBy?: number; // 审核人员ID
    }
  }
}
```

## 2. 人工审核管理

### 2.1 获取待审核内容列表

#### 接口信息

- **URL**: `GET /api/v1/content/audit/pending`
- **Content-Type**: `application/json`
- **描述**: 获取需要人工审核的内容列表（仅教师/管理员可访问）

#### Request 类型

```typescript
// Query参数
interface GetPendingAuditRequest {
  page?: number;        // 页码，默认1
  limit?: number;       // 每页数量，默认20
  contentType?: 'post' | 'reply' | 'comment' | 'all'; // 内容类型筛选
  severity?: 'low' | 'medium' | 'high' | 'all'; // 严重程度筛选
  startDate?: string;   // 开始日期
  endDate?: string;     // 结束日期
}
```

#### Response 类型

```typescript
interface GetPendingAuditResponse {
  code: number;         // 状态码，200表示成功
  message: string;      // 响应消息
  data: {
    audits: {
      id: string;       // 审核ID
      content: string;  // 审核内容
      contentType: string; // 内容类型
      author: {
        id: number;     // 作者ID
        username: string; // 作者用户名
        avatar?: string; // 作者头像
      };
      violations: {
        type: string;   // 违规类型
        severity: string; // 严重程度
        description: string; // 违规描述
      }[];
      confidence: number; // 审核置信度
      createdAt: string; // 提交时间
      priority: 'low' | 'medium' | 'high'; // 优先级
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

### 2.2 人工审核决策

#### 接口信息

- **URL**: `POST /api/v1/content/audit/{auditId}/review`
- **Content-Type**: `application/json`
- **描述**: 教师/管理员对审核内容进行最终决策

#### Request 类型

```typescript
interface ReviewAuditRequest {
  decision: 'approve' | 'reject' | 'request_revision'; // 审核决策
  reason?: string;     // 决策理由
  suggestions?: string[]; // 修改建议（如果要求修改）
  severity?: 'low' | 'medium' | 'high'; // 严重程度评估
}
```

#### Response 类型

```typescript
interface ReviewAuditResponse {
  code: number;         // 状态码，200表示成功
  message: string;      // 响应消息
  data: {
    audit: {
      id: string;       // 审核ID
      status: 'approved' | 'rejected' | 'pending_revision'; // 最终状态
      decision: string; // 审核决策
      reason?: string;  // 决策理由
      reviewedBy: number; // 审核人员ID
      reviewedAt: string; // 审核时间
    }
  }
}
```

### 2.3 获取审核历史

#### 接口信息

- **URL**: `GET /api/v1/content/audit/history`
- **Content-Type**: `application/json`
- **描述**: 获取内容审核历史记录（仅教师/管理员可访问）

#### Request 类型

```typescript
// Query参数
interface GetAuditHistoryRequest {
  page?: number;        // 页码，默认1
  limit?: number;       // 每页数量，默认20
  userId?: number;      // 用户ID筛选
  contentType?: string; // 内容类型筛选
  status?: string;      // 审核状态筛选
  reviewerId?: number;  // 审核人员ID筛选
  startDate?: string;   // 开始日期
  endDate?: string;     // 结束日期
}
```

#### Response 类型

```typescript
interface GetAuditHistoryResponse {
  code: number;         // 状态码，200表示成功
  message: string;      // 响应消息
  data: {
    audits: {
      id: string;       // 审核ID
      content: string;  // 审核内容
      contentType: string; // 内容类型
      author: {
        id: number;     // 作者ID
        username: string; // 作者用户名
      };
      status: string;   // 最终状态
      violations: {
        type: string;   // 违规类型
        severity: string; // 严重程度
        description: string; // 违规描述
      }[];
      decision?: string; // 审核决策
      reason?: string;  // 决策理由
      reviewer?: {
        id: number;     // 审核人员ID
        username: string; // 审核人员用户名
      };
      createdAt: string; // 提交时间
      reviewedAt?: string; // 审核时间
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

## 3. 敏感词管理

### 3.1 获取敏感词列表

#### 接口信息

- **URL**: `GET /api/v1/content/sensitive-words`
- **Content-Type**: `application/json`
- **描述**: 获取系统敏感词列表（仅管理员可访问）

#### Request 类型

```typescript
// Query参数
interface GetSensitiveWordsRequest {
  page?: number;        // 页码，默认1
  limit?: number;       // 每页数量，默认50
  category?: string;    // 敏感词分类
  search?: string;      // 搜索关键词
}
```

#### Response 类型

```typescript
interface GetSensitiveWordsResponse {
  code: number;         // 状态码，200表示成功
  message: string;      // 响应消息
  data: {
    words: {
      id: number;       // 敏感词ID
      word: string;     // 敏感词内容
      category: string; // 分类
      severity: 'low' | 'medium' | 'high'; // 严重程度
      isActive: boolean; // 是否启用
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

### 3.2 添加敏感词

#### 接口信息

- **URL**: `POST /api/v1/content/sensitive-words`
- **Content-Type**: `application/json`
- **描述**: 添加新的敏感词（仅管理员可操作）

#### Request 类型

```typescript
interface AddSensitiveWordRequest {
  word: string;         // 敏感词内容
  category: string;     // 分类
  severity: 'low' | 'medium' | 'high'; // 严重程度
  description?: string; // 描述
}
```

#### Response 类型

```typescript
interface AddSensitiveWordResponse {
  code: number;         // 状态码，200表示成功
  message: string;      // 响应消息
  data: {
    word: {
      id: number;       // 敏感词ID
      word: string;     // 敏感词内容
      category: string; // 分类
      severity: string; // 严重程度
      description?: string; // 描述
      isActive: boolean; // 是否启用
      createdAt: string; // 创建时间
    }
  }
}
```

### 3.3 更新敏感词

#### 接口信息

- **URL**: `PUT /api/v1/content/sensitive-words/{id}`
- **Content-Type**: `application/json`
- **描述**: 更新敏感词信息（仅管理员可操作）

#### Request 类型

```typescript
interface UpdateSensitiveWordRequest {
  word?: string;        // 敏感词内容
  category?: string;    // 分类
  severity?: 'low' | 'medium' | 'high'; // 严重程度
  description?: string; // 描述
  isActive?: boolean;   // 是否启用
}
```

#### Response 类型

```typescript
interface UpdateSensitiveWordResponse {
  code: number;         // 状态码，200表示成功
  message: string;      // 响应消息
  data: {
    word: {
      id: number;       // 敏感词ID
      word: string;     // 敏感词内容
      category: string; // 分类
      severity: string; // 严重程度
      description?: string; // 描述
      isActive: boolean; // 是否启用
      updatedAt: string; // 更新时间
    }
  }
}
```

### 3.4 删除敏感词

#### 接口信息

- **URL**: `DELETE /api/v1/content/sensitive-words/{id}`
- **Content-Type**: `application/json`
- **描述**: 删除敏感词（仅管理员可操作）

#### Request 类型

```typescript
// 路径参数
interface DeleteSensitiveWordRequest {
  id: number;           // 敏感词ID
}
```

#### Response 类型

```typescript
interface DeleteSensitiveWordResponse {
  code: number;         // 状态码，200表示成功
  message: string;      // 响应消息
  data: null;
}
```

## 4. 审核统计

### 4.1 获取审核统计信息

#### 接口信息

- **URL**: `GET /api/v1/content/audit/stats`
- **Content-Type**: `application/json`
- **描述**: 获取内容审核统计信息（仅教师/管理员可访问）

#### Request 类型

```typescript
// Query参数
interface GetAuditStatsRequest {
  startDate?: string;   // 开始日期
  endDate?: string;     // 结束日期
  groupBy?: 'day' | 'week' | 'month'; // 分组方式
}
```

#### Response 类型

```typescript
interface GetAuditStatsResponse {
  code: number;         // 状态码，200表示成功
  message: string;      // 响应消息
  data: {
    stats: {
      date: string;     // 日期
      totalAudits: number; // 总审核数
      approved: number; // 通过数量
      rejected: number; // 拒绝数量
      pending: number;  // 待审核数量
      avgProcessingTime: number; // 平均处理时间
      violationTypes: {
        type: string;   // 违规类型
        count: number;  // 数量
      }[];
    }[];
    summary: {
      totalAudits: number; // 总审核数
      approvalRate: number; // 通过率
      avgProcessingTime: number; // 平均处理时间
      topViolations: {
        type: string;   // 违规类型
        count: number;  // 数量
        percentage: number; // 占比
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
| 422    | 参数验证失败       |
| 429    | 请求频率限制       |
| 500    | 服务器内部错误     |
| 502    | 审核服务不可用     |
| 503    | 审核服务超时       |

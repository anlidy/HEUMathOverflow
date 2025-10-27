# AI集成模块 API 文档

## 1. AI自动回答

### 1.1 触发AI回答

#### 接口信息

- **URL**: `POST /api/v1/ai/generate-answer`
- **Content-Type**: `application/json`
- **描述**: 当新问题发布后，自动调用AI生成回答

#### Request 类型

```typescript
interface GenerateAnswerRequest {
  postId: number;       // 问题帖子ID
  question: string;     // 问题内容
  tags: string[];       // 问题标签
  context?: string;     // 上下文信息，可选
  model?: string;       // 指定AI模型，可选，默认使用配置的模型
}
```

#### Response 类型

```typescript
interface GenerateAnswerResponse {
  code: number;         // 状态码，200表示成功
  message: string;      // 响应消息
  data: {
    answer: {
      id: number;       // AI回答ID
      content: string;  // AI生成的回答内容
      postId: number;   // 所属问题ID
      model: string;    // 使用的AI模型
      confidence: number; // 回答置信度（0-1）
      processingTime: number; // 处理时间（毫秒）
      createdAt: string; // 创建时间
    };
    status: 'success' | 'processing' | 'failed'; // 处理状态
  }
}
```

### 1.2 获取AI回答状态

#### 接口信息

- **URL**: `GET /api/v1/ai/answers/{answerId}/status`
- **Content-Type**: `application/json`
- **描述**: 查询AI回答的生成状态

#### Request 类型

```typescript
// 路径参数
interface GetAnswerStatusRequest {
  answerId: number;     // AI回答ID
}
```

#### Response 类型

```typescript
interface GetAnswerStatusResponse {
  code: number;         // 状态码，200表示成功
  message: string;      // 响应消息
  data: {
    status: 'pending' | 'processing' | 'completed' | 'failed'; // 处理状态
    progress: number;   // 处理进度（0-100）
    estimatedTime?: number; // 预计剩余时间（秒）
    error?: string;     // 错误信息（如果失败）
  }
}
```

### 1.3 重新生成AI回答

#### 接口信息

- **URL**: `POST /api/v1/ai/answers/{answerId}/regenerate`
- **Content-Type**: `application/json`
- **描述**: 重新生成AI回答

#### Request 类型

```typescript
interface RegenerateAnswerRequest {
  answerId: number;     // AI回答ID
  model?: string;       // 指定AI模型，可选
  prompt?: string;      // 自定义提示词，可选
}
```

#### Response 类型

```typescript
interface RegenerateAnswerResponse {
  code: number;         // 状态码，200表示成功
  message: string;      // 响应消息
  data: {
    answer: {
      id: number;       // 新的AI回答ID
      content: string;  // 重新生成的回答内容
      model: string;    // 使用的AI模型
      confidence: number; // 回答置信度
      processingTime: number; // 处理时间
      createdAt: string; // 创建时间
    };
    status: 'success' | 'processing' | 'failed';
  }
}
```

## 2. AI模型管理

### 2.1 获取可用AI模型列表

#### 接口信息

- **URL**: `GET /api/v1/ai/models`
- **Content-Type**: `application/json`
- **描述**: 获取系统中配置的可用AI模型列表

#### Request 类型

```typescript
// 无需请求体
```

#### Response 类型

```typescript
interface GetModelsResponse {
  code: number;         // 状态码，200表示成功
  message: string;      // 响应消息
  data: {
    models: {
      id: string;       // 模型ID
      name: string;     // 模型名称
      provider: string; // 提供商（如豆包、Kimi、智谱清言）
      description: string; // 模型描述
      isActive: boolean; // 是否启用
      maxTokens: number; // 最大token数
      costPerToken: number; // 每token成本
      capabilities: string[]; // 能力列表
    }[];
    defaultModel: string; // 默认模型ID
  }
}
```

### 2.2 更新AI模型配置

#### 接口信息

- **URL**: `PUT /api/v1/ai/models/{modelId}`
- **Content-Type**: `application/json`
- **描述**: 更新AI模型配置（仅管理员可操作）

#### Request 类型

```typescript
interface UpdateModelRequest {
  name?: string;        // 模型名称
  isActive?: boolean;   // 是否启用
  apiKey?: string;      // API密钥
  apiUrl?: string;      // API地址
  maxTokens?: number;   // 最大token数
  temperature?: number; // 温度参数
  topP?: number;       // Top-p参数
  frequencyPenalty?: number; // 频率惩罚
  presencePenalty?: number;  // 存在惩罚
}
```

#### Response 类型

```typescript
interface UpdateModelResponse {
  code: number;         // 状态码，200表示成功
  message: string;      // 响应消息
  data: {
    model: {
      id: string;       // 模型ID
      name: string;     // 模型名称
      provider: string; // 提供商
      isActive: boolean; // 是否启用
      updatedAt: string; // 更新时间
    }
  }
}
```

### 2.3 测试AI模型

#### 接口信息

- **URL**: `POST /api/v1/ai/models/{modelId}/test`
- **Content-Type**: `application/json`
- **描述**: 测试AI模型是否正常工作

#### Request 类型

```typescript
interface TestModelRequest {
  prompt: string;       // 测试提示词
  maxTokens?: number;   // 最大token数
  temperature?: number; // 温度参数
}
```

#### Response 类型

```typescript
interface TestModelResponse {
  code: number;         // 状态码，200表示成功
  message: string;      // 响应消息
  data: {
    response: string;   // AI响应内容
    tokensUsed: number; // 使用的token数
    processingTime: number; // 处理时间（毫秒）
    cost: number;       // 成本
  }
}
```

## 3. 提示词管理

### 3.1 获取系统提示词

#### 接口信息

- **URL**: `GET /api/v1/ai/prompts`
- **Content-Type**: `application/json`
- **描述**: 获取系统配置的AI提示词

#### Request 类型

```typescript
// Query参数
interface GetPromptsRequest {
  type?: 'default' | 'math' | 'physics' | 'chemistry'; // 提示词类型
}
```

#### Response 类型

```typescript
interface GetPromptsResponse {
  code: number;         // 状态码，200表示成功
  message: string;      // 响应消息
  data: {
    prompts: {
      id: string;       // 提示词ID
      type: string;     // 提示词类型
      title: string;    // 提示词标题
      content: string;  // 提示词内容
      isActive: boolean; // 是否启用
      variables: string[]; // 可替换变量列表
      createdAt: string; // 创建时间
      updatedAt: string; // 更新时间
    }[];
  }
}
```

### 3.2 更新系统提示词

#### 接口信息

- **URL**: `PUT /api/v1/ai/prompts/{promptId}`
- **Content-Type**: `application/json`
- **描述**: 更新系统提示词（仅管理员可操作）

#### Request 类型

```typescript
interface UpdatePromptRequest {
  title?: string;       // 提示词标题
  content: string;      // 提示词内容
  isActive?: boolean;   // 是否启用
  variables?: string[]; // 可替换变量列表
}
```

#### Response 类型

```typescript
interface UpdatePromptResponse {
  code: number;         // 状态码，200表示成功
  message: string;      // 响应消息
  data: {
    prompt: {
      id: string;       // 提示词ID
      type: string;     // 提示词类型
      title: string;    // 提示词标题
      content: string;  // 提示词内容
      isActive: boolean; // 是否启用
      variables: string[]; // 可替换变量列表
      updatedAt: string; // 更新时间
    }
  }
}
```

## 4. 知识库集成

### 4.1 搜索相关知识

#### 接口信息

- **URL**: `POST /api/v1/ai/knowledge/search`
- **Content-Type**: `application/json`
- **描述**: 在知识库中搜索与问题相关的知识片段

#### Request 类型

```typescript
interface SearchKnowledgeRequest {
  query: string;        // 搜索查询
  tags?: string[];      // 相关标签
  limit?: number;       // 返回结果数量限制，默认5
  threshold?: number;   // 相似度阈值，默认0.7
}
```

#### Response 类型

```typescript
interface SearchKnowledgeResponse {
  code: number;         // 状态码，200表示成功
  message: string;      // 响应消息
  data: {
    results: {
      id: string;       // 知识片段ID
      title: string;    // 知识片段标题
      content: string;  // 知识片段内容
      source: string;   // 来源文档
      similarity: number; // 相似度分数
      tags: string[];   // 相关标签
    }[];
    totalFound: number; // 找到的总数量
  }
}
```

### 4.2 获取知识库统计

#### 接口信息

- **URL**: `GET /api/v1/ai/knowledge/stats`
- **Content-Type**: `application/json`
- **描述**: 获取知识库的统计信息

#### Request 类型

```typescript
// 无需请求体
```

#### Response 类型

```typescript
interface GetKnowledgeStatsResponse {
  code: number;         // 状态码，200表示成功
  message: string;      // 响应消息
  data: {
    totalDocuments: number; // 总文档数
    totalFragments: number; // 总知识片段数
    totalSize: number;  // 总大小（字节）
    lastUpdated: string; // 最后更新时间
    categories: {
      name: string;     // 分类名称
      count: number;    // 文档数量
      size: number;     // 大小（字节）
    }[];
  }
}
```

## 5. AI使用统计

### 5.1 获取AI使用统计

#### 接口信息

- **URL**: `GET /api/v1/ai/usage/stats`
- **Content-Type**: `application/json`
- **描述**: 获取AI使用统计信息（仅管理员可查看）

#### Request 类型

```typescript
// Query参数
interface GetUsageStatsRequest {
  startDate?: string;   // 开始日期，格式：YYYY-MM-DD
  endDate?: string;     // 结束日期，格式：YYYY-MM-DD
  modelId?: string;     // 模型ID筛选
  groupBy?: 'day' | 'week' | 'month'; // 分组方式
}
```

#### Response 类型

```typescript
interface GetUsageStatsResponse {
  code: number;         // 状态码，200表示成功
  message: string;      // 响应消息
  data: {
    stats: {
      date: string;     // 日期
      totalRequests: number; // 总请求数
      successfulRequests: number; // 成功请求数
      failedRequests: number; // 失败请求数
      totalTokens: number; // 总token数
      totalCost: number; // 总成本
      avgResponseTime: number; // 平均响应时间
      models: {
        modelId: string; // 模型ID
        requests: number; // 请求数
        tokens: number; // token数
        cost: number; // 成本
      }[];
    }[];
    summary: {
      totalRequests: number; // 总请求数
      totalTokens: number; // 总token数
      totalCost: number; // 总成本
      avgResponseTime: number; // 平均响应时间
    }
  }
}
```

### 5.2 获取用户AI使用记录

#### 接口信息

- **URL**: `GET /api/v1/ai/usage/user`
- **Content-Type**: `application/json`
- **描述**: 获取当前用户的AI使用记录

#### Request 类型

```typescript
// Query参数
interface GetUserUsageRequest {
  page?: number;        // 页码，默认1
  limit?: number;       // 每页数量，默认20
  startDate?: string;   // 开始日期
  endDate?: string;     // 结束日期
}
```

#### Response 类型

```typescript
interface GetUserUsageResponse {
  code: number;         // 状态码，200表示成功
  message: string;      // 响应消息
  data: {
    usage: {
      id: number;       // 使用记录ID
      postId: number;   // 相关帖子ID
      model: string;    // 使用的模型
      tokensUsed: number; // 使用的token数
      cost: number;     // 成本
      responseTime: number; // 响应时间
      status: 'success' | 'failed'; // 状态
      createdAt: string; // 创建时间
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
| 502    | AI服务不可用       |
| 503    | AI服务超时         |

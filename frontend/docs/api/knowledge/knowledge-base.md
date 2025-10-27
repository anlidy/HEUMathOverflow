# 知识库管理模块 API 文档

## 1. 文档管理

### 1.1 上传文档

#### 接口信息

- **URL**: `POST /api/v1/knowledge/documents`
- **Content-Type**: `multipart/form-data`
- **描述**: 上传知识库文档（仅管理员可操作）

#### Request 类型

```typescript
// FormData格式
interface UploadDocumentRequest {
  file: File;           // 文档文件（支持PDF, Word, TXT等格式）
  title: string;        // 文档标题
  description?: string; // 文档描述
  category: string;     // 文档分类
  tags: string[];       // 文档标签
  isPublic: boolean;    // 是否公开
  subject: string;      // 学科分类
}
```

#### Response 类型

```typescript
interface UploadDocumentResponse {
  code: number;         // 状态码，200表示成功
  message: string;      // 响应消息
  data: {
    document: {
      id: number;       // 文档ID
      title: string;    // 文档标题
      description?: string; // 文档描述
      category: string; // 文档分类
      tags: string[];   // 文档标签
      subject: string;  // 学科分类
      fileName: string; // 原始文件名
      fileSize: number; // 文件大小（字节）
      fileType: string; // 文件类型
      isPublic: boolean; // 是否公开
      status: 'processing' | 'completed' | 'failed'; // 处理状态
      uploader: {
        id: number;     // 上传者ID
        username: string; // 上传者用户名
      };
      createdAt: string; // 上传时间
      processedAt?: string; // 处理完成时间
    }
  }
}
```

### 1.2 获取文档列表

#### 接口信息

- **URL**: `GET /api/v1/knowledge/documents`
- **Content-Type**: `application/json`
- **描述**: 获取知识库文档列表

#### Request 类型

```typescript
// Query参数
interface GetDocumentsRequest {
  page?: number;        // 页码，默认1
  limit?: number;       // 每页数量，默认20
  category?: string;    // 分类筛选
  subject?: string;     // 学科筛选
  tags?: string[];      // 标签筛选
  status?: 'processing' | 'completed' | 'failed' | 'all'; // 状态筛选
  isPublic?: boolean;   // 是否公开筛选
  search?: string;      // 搜索关键词
  sortBy?: 'createdAt' | 'title' | 'fileSize' | 'processedAt'; // 排序字段
  sortOrder?: 'asc' | 'desc'; // 排序方向
}
```

#### Response 类型

```typescript
interface GetDocumentsResponse {
  code: number;         // 状态码，200表示成功
  message: string;      // 响应消息
  data: {
    documents: {
      id: number;       // 文档ID
      title: string;    // 文档标题
      description?: string; // 文档描述
      category: string; // 文档分类
      tags: string[];   // 文档标签
      subject: string;  // 学科分类
      fileName: string; // 原始文件名
      fileSize: number; // 文件大小
      fileType: string; // 文件类型
      isPublic: boolean; // 是否公开
      status: string;   // 处理状态
      fragmentCount: number; // 知识片段数量
      uploader: {
        id: number;     // 上传者ID
        username: string; // 上传者用户名
      };
      createdAt: string; // 上传时间
      processedAt?: string; // 处理完成时间
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

### 1.3 获取文档详情

#### 接口信息

- **URL**: `GET /api/v1/knowledge/documents/{id}`
- **Content-Type**: `application/json`
- **描述**: 获取指定文档的详细信息

#### Request 类型

```typescript
// 路径参数
interface GetDocumentDetailRequest {
  id: number;           // 文档ID
}
```

#### Response 类型

```typescript
interface GetDocumentDetailResponse {
  code: number;         // 状态码，200表示成功
  message: string;      // 响应消息
  data: {
    document: {
      id: number;       // 文档ID
      title: string;    // 文档标题
      description?: string; // 文档描述
      category: string; // 文档分类
      tags: string[];   // 文档标签
      subject: string;  // 学科分类
      fileName: string; // 原始文件名
      fileSize: number; // 文件大小
      fileType: string; // 文件类型
      isPublic: boolean; // 是否公开
      status: string;   // 处理状态
      fragmentCount: number; // 知识片段数量
      downloadUrl: string; // 下载链接
      uploader: {
        id: number;     // 上传者ID
        username: string; // 上传者用户名
        avatar?: string; // 上传者头像
      };
      createdAt: string; // 上传时间
      processedAt?: string; // 处理完成时间
      updatedAt: string; // 更新时间
    }
  }
}
```

### 1.4 更新文档信息

#### 接口信息

- **URL**: `PUT /api/v1/knowledge/documents/{id}`
- **Content-Type**: `application/json`
- **描述**: 更新文档信息（仅上传者或管理员可操作）

#### Request 类型

```typescript
interface UpdateDocumentRequest {
  title?: string;       // 文档标题
  description?: string; // 文档描述
  category?: string;    // 文档分类
  tags?: string[];      // 文档标签
  subject?: string;     // 学科分类
  isPublic?: boolean;   // 是否公开
}
```

#### Response 类型

```typescript
interface UpdateDocumentResponse {
  code: number;         // 状态码，200表示成功
  message: string;      // 响应消息
  data: {
    document: {
      id: number;       // 文档ID
      title: string;    // 文档标题
      description?: string; // 文档描述
      category: string; // 文档分类
      tags: string[];   // 文档标签
      subject: string;  // 学科分类
      isPublic: boolean; // 是否公开
      updatedAt: string; // 更新时间
    }
  }
}
```

### 1.5 删除文档

#### 接口信息

- **URL**: `DELETE /api/v1/knowledge/documents/{id}`
- **Content-Type**: `application/json`
- **描述**: 删除文档（仅上传者或管理员可操作）

#### Request 类型

```typescript
// 路径参数
interface DeleteDocumentRequest {
  id: number;           // 文档ID
}
```

#### Response 类型

```typescript
interface DeleteDocumentResponse {
  code: number;         // 状态码，200表示成功
  message: string;      // 响应消息
  data: null;
}
```

## 2. 知识片段管理

### 2.1 获取文档知识片段

#### 接口信息

- **URL**: `GET /api/v1/knowledge/documents/{documentId}/fragments`
- **Content-Type**: `application/json`
- **描述**: 获取指定文档的知识片段列表

#### Request 类型

```typescript
// Query参数
interface GetDocumentFragmentsRequest {
  documentId: number;   // 文档ID（路径参数）
  page?: number;        // 页码，默认1
  limit?: number;       // 每页数量，默认20
  search?: string;      // 搜索关键词
  sortBy?: 'position' | 'relevance' | 'createdAt'; // 排序字段
  sortOrder?: 'asc' | 'desc'; // 排序方向
}
```

#### Response 类型

```typescript
interface GetDocumentFragmentsResponse {
  code: number;         // 状态码，200表示成功
  message: string;      // 响应消息
  data: {
    fragments: {
      id: string;       // 片段ID
      content: string;  // 片段内容
      title?: string;   // 片段标题
      position: number; // 在文档中的位置
      pageNumber?: number; // 页码（如果是PDF）
      relevance: number; // 相关性分数
      tags: string[];   // 片段标签
      metadata: {
        [key: string]: any; // 元数据
      };
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

### 2.2 获取知识片段详情

#### 接口信息

- **URL**: `GET /api/v1/knowledge/fragments/{fragmentId}`
- **Content-Type**: `application/json`
- **描述**: 获取指定知识片段的详细信息

#### Request 类型

```typescript
// 路径参数
interface GetFragmentDetailRequest {
  fragmentId: string;   // 片段ID
}
```

#### Response 类型

```typescript
interface GetFragmentDetailResponse {
  code: number;         // 状态码，200表示成功
  message: string;      // 响应消息
  data: {
    fragment: {
      id: string;       // 片段ID
      content: string;  // 片段内容
      title?: string;   // 片段标题
      position: number; // 在文档中的位置
      pageNumber?: number; // 页码
      relevance: number; // 相关性分数
      tags: string[];   // 片段标签
      metadata: {
        [key: string]: any; // 元数据
      };
      document: {
        id: number;     // 所属文档ID
        title: string;  // 文档标题
        category: string; // 文档分类
        subject: string; // 学科分类
      };
      createdAt: string; // 创建时间
    }
  }
}
```

### 2.3 更新知识片段

#### 接口信息

- **URL**: `PUT /api/v1/knowledge/fragments/{fragmentId}`
- **Content-Type**: `application/json`
- **描述**: 更新知识片段信息（仅管理员可操作）

#### Request 类型

```typescript
interface UpdateFragmentRequest {
  title?: string;       // 片段标题
  tags?: string[];      // 片段标签
  metadata?: {
    [key: string]: any; // 元数据
  };
}
```

#### Response 类型

```typescript
interface UpdateFragmentResponse {
  code: number;         // 状态码，200表示成功
  message: string;      // 响应消息
  data: {
    fragment: {
      id: string;       // 片段ID
      title?: string;   // 片段标题
      tags: string[];   // 片段标签
      metadata: {
        [key: string]: any; // 元数据
      };
      updatedAt: string; // 更新时间
    }
  }
}
```

## 3. 知识搜索

### 3.1 搜索知识片段

#### 接口信息

- **URL**: `POST /api/v1/knowledge/search`
- **Content-Type**: `application/json`
- **描述**: 在知识库中搜索相关知识片段

#### Request 类型

```typescript
interface SearchKnowledgeRequest {
  query: string;        // 搜索查询
  filters?: {
    categories?: string[]; // 分类筛选
    subjects?: string[];   // 学科筛选
    tags?: string[];       // 标签筛选
    documents?: number[];  // 文档ID筛选
  };
  options?: {
    limit?: number;     // 返回结果数量限制，默认10
    threshold?: number; // 相似度阈值，默认0.7
    includeMetadata?: boolean; // 是否包含元数据
    highlight?: boolean; // 是否高亮匹配内容
  };
}
```

#### Response 类型

```typescript
interface SearchKnowledgeResponse {
  code: number;         // 状态码，200表示成功
  message: string;      // 响应消息
  data: {
    results: {
      id: string;       // 片段ID
      content: string;  // 片段内容
      title?: string;   // 片段标题
      similarity: number; // 相似度分数
      highlights?: string[]; // 高亮内容
      document: {
        id: number;     // 所属文档ID
        title: string;  // 文档标题
        category: string; // 文档分类
        subject: string; // 学科分类
      };
      tags: string[];   // 片段标签
      metadata: {
        [key: string]: any; // 元数据
      };
    }[];
    totalFound: number; // 找到的总数量
    searchTime: number; // 搜索耗时（毫秒）
    suggestions?: string[]; // 搜索建议
  }
}
```

### 3.2 获取搜索建议

#### 接口信息

- **URL**: `GET /api/v1/knowledge/search/suggestions`
- **Content-Type**: `application/json`
- **描述**: 获取搜索建议

#### Request 类型

```typescript
// Query参数
interface GetSearchSuggestionsRequest {
  query: string;        // 搜索查询
  limit?: number;       // 建议数量限制，默认5
}
```

#### Response 类型

```typescript
interface GetSearchSuggestionsResponse {
  code: number;         // 状态码，200表示成功
  message: string;      // 响应消息
  data: {
    suggestions: {
      text: string;     // 建议文本
      type: 'tag' | 'category' | 'subject' | 'title'; // 建议类型
      count: number;    // 相关数量
    }[];
  }
}
```

## 4. 分类管理

### 4.1 获取分类列表

#### 接口信息

- **URL**: `GET /api/v1/knowledge/categories`
- **Content-Type**: `application/json`
- **描述**: 获取知识库分类列表

#### Request 类型

```typescript
// Query参数
interface GetCategoriesRequest {
  type?: 'category' | 'subject' | 'tag' | 'all'; // 分类类型
  parentId?: number;   // 父分类ID（用于层级分类）
}
```

#### Response 类型

```typescript
interface GetCategoriesResponse {
  code: number;         // 状态码，200表示成功
  message: string;      // 响应消息
  data: {
    categories: {
      id: number;       // 分类ID
      name: string;     // 分类名称
      type: string;     // 分类类型
      description?: string; // 分类描述
      parentId?: number; // 父分类ID
      documentCount: number; // 文档数量
      fragmentCount: number; // 片段数量
      children?: {
        id: number;     // 子分类ID
        name: string;   // 子分类名称
        documentCount: number; // 文档数量
      }[];
    }[];
  }
}
```

### 4.2 创建分类

#### 接口信息

- **URL**: `POST /api/v1/knowledge/categories`
- **Content-Type**: `application/json`
- **描述**: 创建新分类（仅管理员可操作）

#### Request 类型

```typescript
interface CreateCategoryRequest {
  name: string;         // 分类名称
  type: 'category' | 'subject' | 'tag'; // 分类类型
  description?: string; // 分类描述
  parentId?: number;    // 父分类ID
}
```

#### Response 类型

```typescript
interface CreateCategoryResponse {
  code: number;         // 状态码，200表示成功
  message: string;      // 响应消息
  data: {
    category: {
      id: number;       // 分类ID
      name: string;     // 分类名称
      type: string;     // 分类类型
      description?: string; // 分类描述
      parentId?: number; // 父分类ID
      createdAt: string; // 创建时间
    }
  }
}
```

### 4.3 更新分类

#### 接口信息

- **URL**: `PUT /api/v1/knowledge/categories/{id}`
- **Content-Type**: `application/json`
- **描述**: 更新分类信息（仅管理员可操作）

#### Request 类型

```typescript
interface UpdateCategoryRequest {
  name?: string;        // 分类名称
  description?: string; // 分类描述
  parentId?: number;    // 父分类ID
}
```

#### Response 类型

```typescript
interface UpdateCategoryResponse {
  code: number;         // 状态码，200表示成功
  message: string;      // 响应消息
  data: {
    category: {
      id: number;       // 分类ID
      name: string;     // 分类名称
      type: string;     // 分类类型
      description?: string; // 分类描述
      parentId?: number; // 父分类ID
      updatedAt: string; // 更新时间
    }
  }
}
```

### 4.4 删除分类

#### 接口信息

- **URL**: `DELETE /api/v1/knowledge/categories/{id}`
- **Content-Type**: `application/json`
- **描述**: 删除分类（仅管理员可操作）

#### Request 类型

```typescript
// 路径参数
interface DeleteCategoryRequest {
  id: number;           // 分类ID
}
```

#### Response 类型

```typescript
interface DeleteCategoryResponse {
  code: number;         // 状态码，200表示成功
  message: string;      // 响应消息
  data: null;
}
```

## 5. 知识库统计

### 5.1 获取知识库统计信息

#### 接口信息

- **URL**: `GET /api/v1/knowledge/stats`
- **Content-Type**: `application/json`
- **描述**: 获取知识库统计信息

#### Request 类型

```typescript
// Query参数
interface GetKnowledgeStatsRequest {
  startDate?: string;   // 开始日期
  endDate?: string;     // 结束日期
}
```

#### Response 类型

```typescript
interface GetKnowledgeStatsResponse {
  code: number;         // 状态码，200表示成功
  message: string;      // 响应消息
  data: {
    overview: {
      totalDocuments: number; // 总文档数
      totalFragments: number; // 总片段数
      totalSize: number; // 总大小（字节）
      publicDocuments: number; // 公开文档数
      privateDocuments: number; // 私有文档数
    };
    categories: {
      name: string;     // 分类名称
      documentCount: number; // 文档数量
      fragmentCount: number; // 片段数量
      size: number;     // 大小（字节）
    }[];
    subjects: {
      name: string;     // 学科名称
      documentCount: number; // 文档数量
      fragmentCount: number; // 片段数量
    }[];
    recentActivity: {
      date: string;     // 日期
      uploads: number;  // 上传数量
      searches: number; // 搜索次数
    }[];
    topTags: {
      tag: string;      // 标签
      count: number;    // 使用次数
    }[];
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
| 413    | 文件过大           |
| 415    | 不支持的文件类型   |
| 422    | 参数验证失败       |
| 500    | 服务器内部错误     |
| 502    | 文档处理服务不可用 |
| 503    | 文档处理服务超时   |

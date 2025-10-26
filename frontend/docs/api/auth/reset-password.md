# 密码重置 API 文档

## 接口描述

用户密码重置接口，通过邮箱验证重置密码。

## 请求信息

- 请求路径：`/api/auth/reset-password`
- 请求方法：`POST`
- 请求头：
    ```
    Content-Type: application/json
    ```

## 请求参数

| 参数名      | 类型   | 必填 | 描述   | 示例           |
| ----------- | ------ | ---- | ------ | -------------- |
| email       | string | 是   | 邮箱   | "john@doe.com" |
| verifyCode  | string | 是   | 验证码 | "123456"       |
| newPassword | string | 是   | 新密码 | "newpass123"   |

## 响应信息

### 成功响应

```json
{
    "code": 200,
    "message": "密码重置成功"
}
```

### 错误响应

```json
{
    "code": 1009,
    "message": "验证码无效"
}
```

## 错误码说明

| 错误码 | 说明       |
| ------ | ---------- |
| 1009   | 验证码无效 |
| 1010   | 验证码过期 |
| 1011   | 邮箱不存在 |

## 示例代码

```typescript
// 密码重置请求示例
const resetData = {
  email: "john@doe.com",
  verifyCode: "123456",
  newPassword: "newpass123"
};

const response = await fetch('/api/auth/reset-password', {
  method: 'POST',
  headers: {
    'Content-Type': 'application/json'
  },
  body: JSON.stringify(resetData)
});

const result = await response.json();
```

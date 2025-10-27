#### Go版本: 1.23.5

#### 安装依赖库: 
```bash
 cd backend/
 go mod download
```

#### 项目结构:

 `/cmd `  各个go服务的启动入口(main.go)

`/docker` docker配置文件

`/internal` 程序的内部实现

- `/common` 各服务共享的一些模块
- `/common/config` 共享的配置文件
- `/common/db` 共享的数据库实例化函数
- `/common/middleware` 共享的中间件函数
- `/common/utils` 共享的工具函数

`/internal/user-service` 用户服务

- `/controller` 服务控制层, 负责解析请求, 调用service层的方法进行处理, 返回请求
- `/service` 服务层, 负责完成业务逻辑, 调用resository层的方法来控制数据库, 返回处理结果
- `/repository` 数据层, 封装了数据库操作逻辑, 对service层暴露接口
- `/router` 路由层, 负责定义服务的请求路径和处理函数, 调用中间件
- `model` 数据模型, 包括用户表结构, 请求/响应体结构等 

# Telegrafana设计文档

Telegrafana是一个管理Telegraf的Docker实例的可视化工具，包括的功能有：

+ 新建配置Telegraf实例

+ Telegraf运行状态监控

+ Telegraf实例的启停

## 详细设计

### Web UI

#### 实例列表界面

#### 实例详细信息（包括配置信息）

### API Server

#### 获取实例列表

#### 获取实例信息

#### 新建实例

#### 删除实例

#### 更新实例配置

### Telegraf的Docker实例管理

#### 运行新的实例

+ 输入参数：Telegraf配置的URL

+ 返回结果：Docker实例的ID和错误信息

#### 启动实例

+ 输入参数：Docker实例ID

+ 返回结果：错误信息

#### 停止实例

+ 输入参数：Docker实例ID

+ 返回结果：错误信息

#### 重启实例

+ 输入参数：Docker实例ID

+ 返回结果：错误信息

#### 获取实例状态

+ 输入参数：Docker实例ID

+ 返回结果：错误信息

### Telegraf配置管理
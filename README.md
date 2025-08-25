# UDP Punch

WireGuard UDP 打洞工具，灵感来源于 [natpunch-go](https://github.com/malcolmseyd/natpunch-go)

## 功能特色

- ✨ **Web 监控界面** - 实时查看连接的客户端状态
- 🔐 **安全认证** - Web 界面支持 HTTP Basic 认证保护
- 📊 **JSON API** - 提供 RESTful API 接口查询客户端信息
- 🔄 **自动刷新** - 监控页面支持自动和手动刷新
- 🌐 **IPv4 优化** - 专用 IPv4 UDP 监听，提升兼容性

## 使用方法

### 服务器端

```bash
# 基本启动（UDP端口19993，Web端口8080）
./punch-server-linux-amd64 -port 19993

# 自定义Web端口和密码
./punch-server-linux-amd64 -port 19993 -web-port 8080 -web-pass mypassword
```

**服务器参数说明：**
- `-port`: UDP 服务端口（默认: 19993）
- `-web-port`: Web 监控界面端口（默认: 8080）
- `-web-pass`: Web 界面密码（默认: admin）

### 客户端

> 确保 WireGuard 接口已启动

```bash
./punch-client-linux-amd64 -server xxxx:19993 -iface wg0
```

### Web 监控界面

启动服务器后，访问监控界面：
- 地址: `http://服务器IP:8080`
- 用户名: `admin`
- 密码: 启动时设置的密码（默认: `admin`）

监控界面功能：
- 查看服务器状态和端口信息
- 实时显示活跃客户端列表
- 显示客户端公钥、IP地址、端口和最后活跃时间
- 每10秒自动刷新，支持手动刷新

### API 接口

JSON API 端点: `http://服务器IP:8080/api`

返回格式：
```json
{
  "status": "running",
  "udp_port": 19993,
  "web_port": 8080,
  "client_count": 2,
  "clients": [
    {
      "key": "客户端公钥",
      "address": "客户端IP:端口",
      "timestamp": "2024-01-01T12:00:00Z"
    }
  ],
  "timestamp": "2024-01-01T12:00:00Z"
}
```

## 编译构建

```bash
# 编译本地平台版本
make build

# 交叉编译所有平台
make build_all

# 清理构建文件
make clean
```

## 相关资源

- [natpunch-go](https://github.com/malcolmseyd/natpunch-go) - 原始项目灵感来源
- [wireguard-vanity-address](https://github.com/yinheli/wireguard-vanity-address) - 生成指定前缀的密钥对
- [UDP hole punching](https://en.wikipedia.org/wiki/UDP_hole_punching) - UDP 打洞技术原理

## 更新日志

### v1.1.0 (当前版本)
- ✨ 新增 Web 监控界面
- 🔐 添加 HTTP Basic 认证
- 📊 提供 JSON API 接口
- 🔄 支持自动刷新功能
- 🌐 优化 IPv4 网络兼容性

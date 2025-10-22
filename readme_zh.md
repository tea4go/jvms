# JVMS - 全平台 JDK 版本管理器

[![GitHub release](https://img.shields.io/github/release/tea4go/jvms.svg)](https://github.com/tea4go/jvms/releases)
[![License](https://img.shields.io/github/license/tea4go/jvms.svg)](https://github.com/tea4go/jvms/blob/main/LICENSE)

> 在 windows/macos/linux 系统上轻松管理多个 JDK 版本

[JVMS](https://github.com/tea4go/jvms) 是一个专为 windows/macos/linux 设计的 JDK 版本管理工具，让您可以在同一台计算机上安装和切换多个 JDK 版本。

## 特性

- ✅ **多版本管理** - 在一台机器上安装和管理多个 JDK 版本
- 🔄 **快速切换** - 无需重启，即时切换 JDK 版本
- 📦 **多源支持** - 支持从多个镜像源下载 JDK
  - 清华大学软件镜像库
  - 兰州大学软件镜像库
  - injdk
  - Adoptium (Eclipse Temurin)
  - Azul Zulu
  - 华为云 OpenJDK 镜像
- 🎯 **智能匹配** - 自动识别系统架构，下载适配版本
- 🚀 **零依赖** - 使用 Go 编写，无需预装 JDK
- 🔗 **符号链接** - 使用符号链接技术，切换后所有终端立即生效
- 🏠 **本地版本** - 支持添加本地 JDK 版本
- 🌐 **自定义源** - 支持配置私有下载服务器

## 为什么需要 JVMS？

在开发过程中，您可能需要：
- 用不同 JDK 版本测试项目兼容性
- 某些项目需要特定 JDK 版本
- 体验最新的 JDK 特性而不影响稳定版本
- 在旧版本和新版本之间快速切换

JVMS 让这一切变得简单！

## 安装

### 快速安装

1. **下载最新版本**
   - 访问 [Releases 页面](https://github.com/tea4go/jvms/releases)
   - 下载最新的 `jvms.zip`

2. **解压并配置**
   ```cmd
   # 解压 zip 文件到你想要的位置，例如：C:\jvms
   # 注意：建议选择一个固定的位置，避免频繁移动
   ```

3. **初始化 JVMS**
   ```cmd
   # 以管理员身份运行
   cd C:\jvms
   jvms.exe init
   ```

4. **完成！**

   初始化完成后，JVMS 会自动配置环境变量。重新打开终端即可使用。

![安装示例](images/下载jdk.jpg)

## 使用指南

### 命令概览

```
NAME:
   jvms - 全平台 JDK 版本管理器

USAGE:
   jvms.exe [全局选项] 命令 [命令选项] [参数...]

VERSION:
   2.0.0

COMMANDS:
     init        初始化配置文件
     list, ls    列出已安装的 JDK 版本
     install, i  安装远程可用的 JDK 版本
     switch, s   切换到指定的 JDK 版本
     use, u      临时使用指定的 JDK 版本（当前会话）
     remove, rm  删除指定的 JDK 版本
     rls         显示可供下载的 JDK 版本列表
     proxy       设置下载代理
     help, h     显示命令列表或命令帮助

全局选项:
   --help, -h     显示帮助
   --version, -v  显示版本号
```

### 基本使用流程

#### 1. 查看可下载的 JDK 版本

```cmd
# 查看前 10 个可用版本
jvms rls

# 查看所有可用版本
jvms rls -a
```

输出示例：
```
从本地缓存加载版本列表...
  1) openjdk-25
  2) openjdk-24.0.2
  3) openjdk-23.0.2
  4) openjdk-22.0.2
  5) openjdk-21.0.2
  6) openjdk-20.0.2
  7) openjdk-19.0.2
  8) openjdk-18.0.2.1
  9) openjdk-17.0.2
 10) openjdk-16.0.2

使用 "jvm rls -a" 显示所有版本
```

#### 2. 安装 JDK

```cmd
# 以管理员身份运行
$ jvms install 17.0.6

# 或安装指定的版本
$ jvms install openjdk-21
从本地缓存加载版本列表...
正在下载 JDK 版本 openjdk-21.0.2...
 201.33 MB / 201.33 MB [=================================] 100.00% 6.19 MB/s 32s
完成
正在安装 JDK openjdk-21.0.2 ...
安装成功完成。如果您想使用此版本，请使用: jvms switch openjdk-21.0.2
```
> 下载速度从50+KB，提升到了6+MB（在同一条宽带，速度提升100倍）

#### 3. 查看已安装的版本

```cmd
jvms list
# 或简写
jvms ls
```

#### 4. 切换 JDK 版本

```cmd
# 全局切换（所有新终端生效）
jvms switch 17.0.6

# 或使用 use 命令临时切换（仅当前会话）
jvms use 21.0.4
```

#### 5. 验证切换结果

```cmd
java -version
```

![使用示例](images/安装jdk.jpg)

### 高级功能

#### 添加本地 JDK 版本

如果你已经下载了 JDK，可以手动添加到 JVMS：

1. 找到 JVMS 的安装目录（例如：`C:\jvms`）
2. 进入 `store` 子目录
3. 将 JDK 文件夹复制到 `store` 目录
4. 重命名为版本号（例如：`17.0.1`）
5. 运行 `jvms list` 确认
6. 运行 `jvms switch 17.0.1` 切换版本

**目录结构示例：**
```
C:\jvms\
  ├── jvms.exe
  ├── config.json
  └── store\
      ├── 17.0.1\
      │   ├── bin\
      │   ├── lib\
      │   └── ...
      └── 21.0.4\
          ├── bin\
          ├── lib\
          └── ...
```

![JDK目录示例](images/安装目录.jpg)

#### 配置下载代理

如果需要通过代理下载：

```cmd
jvms proxy http://proxy.example.com:8080
```

#### 搭建私有 JDK 下载服务器

适用于企业内网环境或需要自定义 JDK 版本的场景。

**1. 创建索引文件**

创建 `index.json` 文件：

```json
[
  {
    "version": "17.0.10-custom",
    "url": "http://192.168.1.100/jdk/jdk-17.0.10-custom.zip"
  },
  {
    "version": "21.0.2-internal",
    "url": "http://192.168.1.100/jdk/jdk-21.0.2-internal.zip"
  }
]
```

**2. 部署到 HTTP 服务器**

将 `jdkdlindex.json` 和 JDK zip 文件部署到 Nginx、Apache 或任何静态文件服务器。

**3. 配置 JVMS**

```cmd
jvms init
```

可选：自定义 JAVA_HOME 路径
```cmd
jvms init --java_home D:\MyJDK
```

**4. 使用私有源**

```cmd
jvms rls                      # 列出私有源中的 JDK 版本
jvms install 17.0.10-custom   # 安装私有源中的版本
```

**制作 JDK zip 包：**
1. 打开 JDK 安装目录（包含 `bin`、`lib` 等文件夹）
2. 选中所有文件和文件夹
3. 压缩为 `.zip` 格式（注意：不要包含外层文件夹）
4. 上传到你的服务器
5. 在 `index.json` 中添加对应的链接

## 工作原理

### 为什么选择符号链接？

管理多个 JDK 版本通常有两种方式：

1. **修改 PATH 环境变量** - 每次切换都修改系统 PATH，或使用批处理文件重定向
   - ❌ 需要重启终端才能生效
   - ❌ 实现复杂，容易出问题

2. **使用符号链接（JVMS 的方案）** - 在 PATH 中放置一个符号链接，切换时只需更新链接目标
   - ✅ 所有终端立即生效，无需重启
   - ✅ 实现简洁，稳定可靠
   - ✅ 系统重启后仍然有效

### JVMS 的实现

- **初始化时**：`jvms init` 创建符号链接并添加到系统 PATH
- **切换时**：`jvms switch x.x.x` 只需更新符号链接的目标
- **无需管理员**：切换操作需要管理员权限，但初始化后只需切换时运行一次

这种方式既保证了便利性，又最大化了性能和稳定性。

## 多镜像源支持

JVMS 自动从以下镜像源获取 JDK 版本：

### Adoptium (Eclipse Temurin) (移除)
- OpenJDK 的官方发行版
- 长期支持（LTS）版本
- 企业级质量保证

### Azul Zulu (移除)
- Azul Systems 提供的 OpenJDK 发行版
- 支持多种平台和架构
- 提供商业支持选项

### 华为云 OpenJDK 镜像（新增）
- 国内镜像，下载速度快
- 自动匹配系统架构（amd64/arm64）
- 支持 JDK 9 到 JDK 25

JVMS 会自动筛选适合你当前系统（Windows/amd64）的版本，无需手动选择。

## 常见问题

### Q: 是否需要预先安装 JDK？
A: 不需要。JVMS 使用 Go 语言编写，完全独立运行。

### Q: 切换版本需要重启电脑吗？
A: 不需要。使用符号链接技术，所有打开的终端会立即生效。

### Q: 可以同时使用多个 JDK 版本吗？
A: 可以。安装多个版本后，使用 `jvms switch` 全局切换，或 `jvms use` 在当前会话临时切换。

### Q: 如何卸载 JVMS？
A: 删除 JVMS 安装目录，并手动清理系统 PATH 中的 JAVA_HOME 相关配置即可。

### Q: JVMS 支持哪些 Windows 版本？
A: 支持 Windows 7 及以上版本（需要管理员权限创建符号链接）。

### Q: 为什么需要管理员权限？
A: Windows 创建符号链接需要管理员权限。只在 `jvms init` 和 `jvms switch` 时需要。

## 项目信息

### 技术栈

- 编程语言：Go
- 依赖管理：Go Modules
- HTML解析：golang.org/x/net/html

### 贡献指南

欢迎提交 Issue 和 Pull Request！

1. Fork 本仓库
2. 创建特性分支 (`git checkout -b feature/AmazingFeature`)
3. 提交更改 (`git commit -m 'Add some AmazingFeature'`)
4. 推送到分支 (`git push origin feature/AmazingFeature`)
5. 开启 Pull Request

### 更新日志

查看 [Releases](https://github.com/tea4go/jvms/releases) 了解版本更新历史。

### 致谢

本项目灵感来源于 Node.js 社区的 NVM（Node Version Manager），并针对 Windows 和 JDK 进行了优化。

## 许可证

MIT License

Copyright (c) 2024 JVMS Contributors

详见 [LICENSE](LICENSE) 文件。

---

## 快速链接

- [下载最新版本](https://github.com/tea4go/jvms/releases)
- [报告问题](https://github.com/tea4go/jvms/issues)
- [查看源码](https://github.com/tea4go/jvms)
- [参与贡献](https://github.com/tea4go/jvms/pulls)

**如果 JVMS 对你有帮助，请给我们一个 ⭐️ Star！**

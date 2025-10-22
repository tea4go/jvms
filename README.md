# JVMS - Cross-Platform JDK Version Manager

[![GitHub release](https://img.shields.io/github/release/tea4go/jvms.svg)](https://github.com/tea4go/jvms/releases)
[![License](https://img.shields.io/github/license/tea4go/jvms.svg)](https://github.com/tea4go/jvms/blob/main/LICENSE)

[简体中文 🇨🇳](./README_CN.md)

> Manage multiple JDK versions easily on Windows/macOS/Linux.

[JVMS](https://github.com/tea4go/jvms) is a JDK version manager designed for Windows, macOS, and Linux, making it simple to install and switch between multiple JDK versions on the same machine.

## Features

- ✅ **Multi-version management** - Install and manage multiple JDK versions on one machine
- 🔄 **Fast switching** - Switch JDK versions instantly without rebooting
- 📦 **Multi-source support** - Download JDK builds from multiple mirrors
  - Tsinghua University Open Source Software Mirror
  - Lanzhou University Mirror
  - injdk
  - Adoptium (Eclipse Temurin)
  - Azul Zulu
  - Huawei Cloud OpenJDK Mirror
- 🎯 **Smart detection** - Identify your system architecture and download the matching build automatically
- 🚀 **Zero dependencies** - Written in Go, no JDK required beforehand
- 🔗 **Symbolic links** - Use symlinks so every terminal picks up the change immediately
- 🏠 **Local versions** - Add locally downloaded JDK archives
- 🌐 **Custom sources** - Configure private download servers

## Why JVMS?

During development you may need to:
- Test project compatibility across different JDK versions
- Run projects that require a specific JDK
- Try new JDK features without disturbing your stable setup
- Switch quickly between legacy and cutting-edge versions

JVMS makes it all effortless!

## Installation

### Quick install

1. **Download the latest release**
   - Visit the [Releases page](https://github.com/tea4go/jvms/releases)
   - Download the newest `jvms.zip`

2. **Extract and configure**
   ```cmd
   # Extract the zip file to your preferred location, for example: C:\jvms
   # Tip: choose a stable location and avoid moving it frequently
   ```

3. **Initialize JVMS**
   ```cmd
   # Run as administrator
   cd C:\jvms
   jvms.exe init
   ```

4. **Done!**

   After initialization, JVMS configures the environment variables automatically. Reopen your terminal to start using it.

![Installation example](images/下载jdk.jpg)

## Usage Guide

### Command overview

```
NAME:
   jvms - Cross-platform JDK version manager

USAGE:
   jvms.exe [global options] command [command options] [arguments...]

VERSION:
   2.0.0

COMMANDS:
     init        Initialize configuration files
     list, ls    List installed JDK versions
     install, i  Install a remote JDK version
     switch, s   Switch to a specific JDK version
     use, u      Temporarily use a version for the current session
     remove, rm  Remove a JDK version
     rls         Show available JDK versions
     proxy       Configure a download proxy
     help, h     Show command list or help for a command

GLOBAL OPTIONS:
   --help, -h     Show help
   --version, -v  Print version information
```

### Basic workflow

#### 1. Browse downloadable JDK versions

```cmd
# Show the first 10 available versions
jvms rls

# Show all available versions
jvms rls -a
```

Output example:
```
Loading version list from local cache...
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

Use "jvm rls -a" to display all versions
```

#### 2. Install a JDK

```cmd
# Run as administrator
$ jvms install 17.0.6

# Or install a named version
$ jvms install openjdk-21
Loading version list from local cache...
Downloading JDK version openjdk-21.0.2...
 201.33 MB / 201.33 MB [=================================] 100.00% 6.19 MB/s 32s
Done
Installing JDK openjdk-21.0.2 ...
Installation finished successfully. To use this version, run: jvms switch openjdk-21.0.2
```
> Download speed jumped from 50+ KB/s to 6+ MB/s (same broadband, 100x faster)

#### 3. Check installed versions

```cmd
jvms list
# or shorthand
jvms ls
```

#### 4. Switch JDK versions

```cmd
# Global switch (affects all new terminals)
jvms switch 17.0.6

# Temporarily switch for the current session
jvms use 21.0.4
```

#### 5. Verify the switch

```cmd
java -version
```

![Usage example](images/安装jdk.jpg)

### Advanced features

#### Add a local JDK version

If you already have a JDK, you can add it manually:

1. Locate the JVMS installation directory (for example: `C:\jvms`)
2. Enter the `store` subdirectory
3. Copy the JDK folder into `store`
4. Rename it to the version number (for example: `17.0.1`)
5. Run `jvms list` to confirm
6. Run `jvms switch 17.0.1` to switch versions

**Directory structure example:**
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

![JDK directory example](images/安装目录.jpg)

#### Configure a download proxy

If you need to download through a proxy:

```cmd
jvms proxy http://proxy.example.com:8080
```

#### Set up a private JDK download server

Ideal for corporate intranets or customized JDK builds.

**1. Create an index file**

Create an `index.json` file:

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

**2. Deploy to an HTTP server**

Place `jdkdlindex.json` and the JDK zip files on Nginx, Apache, or any static file server.

**3. Configure JVMS**

```cmd
jvms init
```

Optional: customize the JAVA_HOME path
```cmd
jvms init --java_home D:\MyJDK
```

**4. Use the private source**

```cmd
jvms rls                      # List JDK versions provided by the private source
jvms install 17.0.10-custom   # Install a version from the private source
```

**Create a JDK zip package:**
1. Open the JDK installation directory (contains `bin`, `lib`, etc.)
2. Select all files and folders
3. Compress them into a `.zip` archive (do not include the outer folder)
4. Upload the archive to your server
5. Add the corresponding link in `index.json`

## How It Works

### Why symbolic links?

Managing multiple JDK versions usually comes down to two approaches:

1. **Modify the PATH environment variable** - Update PATH each time you switch or rely on batch scripts
   - ❌ Requires restarting the terminal to take effect
   - ❌ Complex to implement and error-prone

2. **Use symbolic links (the JVMS approach)** - Keep a symlink in PATH and update its target when switching
   - ✅ Every terminal reflects the change immediately, no restart needed
   - ✅ Simple and reliable implementation
   - ✅ Persists across system reboots

### How JVMS works

- **During initialization**: `jvms init` creates symbolic links and adds them to PATH
- **When switching**: `jvms switch x.x.x` updates the symlink target
- **No administrator afterward**: switching requires admin privileges, but only needs to run once per switch

This approach keeps things convenient while maximizing performance and stability.

## Mirror Support

JVMS automatically fetches JDK builds from the following mirrors:

### Adoptium (Eclipse Temurin) (removed)
- Official OpenJDK distribution
- Long-term support (LTS) releases
- Enterprise-grade quality assurance

### Azul Zulu (removed)
- OpenJDK builds provided by Azul Systems
- Supports multiple platforms and architectures
- Offers commercial support options

### Huawei Cloud OpenJDK Mirror (added)
- Domestic mirror with fast download speeds
- Automatically detects system architecture (amd64/arm64)
- Supports JDK 9 through JDK 25

JVMS automatically filters versions that match your current system (Windows/amd64), so there is no need to pick manually.

## FAQ

### Q: Do I need to pre-install a JDK?
A: No. JVMS is written in Go and runs independently.

### Q: Do I need to restart my computer after switching versions?
A: No. With symbolic links, every open terminal reflects the change instantly.

### Q: Can I use multiple JDK versions simultaneously?
A: Yes. Install multiple versions, then use `jvms switch` to change globally or `jvms use` for the current session.

### Q: How do I uninstall JVMS?
A: Delete the JVMS installation directory and manually clean any JAVA_HOME entries from your PATH.

### Q: Which Windows versions does JVMS support?
A: Windows 7 or later (administrator privileges are required to create symbolic links).

### Q: Why are administrator privileges needed?
A: Windows requires administrator rights to create symbolic links. You only need them for `jvms init` and `jvms switch`.

## Project Info

### Tech stack

- Language: Go
- Dependency management: Go Modules
- HTML parsing: golang.org/x/net/html

### Contribution guide

Contributions via Issues and Pull Requests are welcome!

1. Fork this repository
2. Create a feature branch (`git checkout -b feature/AmazingFeature`)
3. Commit your changes (`git commit -m 'Add some AmazingFeature'`)
4. Push the branch (`git push origin feature/AmazingFeature`)
5. Open a Pull Request

### Changelog

See [Releases](https://github.com/tea4go/jvms/releases) for the version history.

### Acknowledgements

This project is inspired by the Node.js community's NVM (Node Version Manager) and optimized for Windows and JDK workflows.

## License

MIT License

Copyright (c) 2024 JVMS Contributors

See the [LICENSE](LICENSE) file for details.

---

## Quick Links

- [Download the latest release](https://github.com/tea4go/jvms/releases)
- [Report issues](https://github.com/tea4go/jvms/issues)
- [Browse the source](https://github.com/tea4go/jvms)
- [Contribute](https://github.com/tea4go/jvms/pulls)

**If JVMS helps you, please give us a ⭐️ Star!**

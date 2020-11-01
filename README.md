# minit

[![BMC Donate](https://img.shields.io/badge/BMC-Donate-orange)](https://www.buymeacoffee.com/vFa5wfRq6)

一个用 Go 编写的命令行工具，用以在 Docker 容器内启动多个进程，并支持 Cron 类型的进程

## 获取镜像

`guoyk93/minit`

## 使用方法

使用多阶段 Dockerfile 来从上述镜像地址导入 `minit` 可执行程序

```dockerfile
FROM guoyk/minit:1.2.1 AS minit

FROM xxxxxxx

# 添加一份服务配置到 /etc/minit.d/
ADD my-service.yml /etc/minit.d/my-service.yml
# 这将从 minit 镜像中，将可执行文件 /minit 拷贝到最终镜像的 /minit 位置
COPY --from=minit /minit /minit
# 这将指定 /minit 作为主启动程序
CMD ["/minit"]
```

## 配置文件

配置文件默认从 `/etc/minit.d/*.yml` 读取

每个配置单元必须具有唯一的 `name`，控制台输出默认会分别记录在 `/var/log/minit` 文件夹内

允许使用 `---` 分割在单个 `yaml` 文件中，写入多条配置单元

当前支持以下类型

### `render`

`render` 类型配置单元最先运行，一般用于渲染配置文件

如下示例

`/etc/minit.d/render-test.yml`

```yaml
kind: render
name: render-test
files:
    - /tmp/*.txt
```

`/tmp/sample.txt`

```text
Hello, {{uppercase .Env.HOME}}
```

`minit` 启动时，会按照配置规则，渲染 `/tmp/sample.txt` 文件

由于容器用户默认为 `root`，因此 `/tmp/sample.txt` 文件会被渲染为

```text
Hello, /ROOT
```

可用渲染函数，参见 `github.com/guoyk93/tmplfuncs`

### `once`

`once` 类型的配置单元随后运行，用于执行一次性进程

`/etc/minit.d/sample.yml`

```yaml
kind: once
name: once-sample
dir: /work # 指定工作目录
command:
    - echo
    - once
```

### `daemon`

`daemon` 类型的配置单元，最后启动，用于执行常驻进程

```yaml
kind: daemon
name: daemon-sample
dir: /work # 指定工作目录
count: 3 # 如果指定了 count，会启动多个副本
command:
    - sleep
    - 9999
```

### `cron`

`cron` 类型的配置单元，最后启动，用于按照 cron 表达式，执行命令

```yaml
kind: cron
name: cron-sample
cron: "* * * * *"
dir: /work # 指定工作目录
command:
    - echo
    - cron
```

### `logrotate`

`logrotate` 类型的配置单元，最后启动

`logrotate` 会在每天凌晨执行以下动作

1. 寻找 `files` 字段指定的，不包含 `YYYY-MM-DD` 标记的文件，进行按日重命名
2. 按照 `keep` 字段删除过期日
3. 在 `dir` 目录执行 `command`

```yaml
kind: logrotate
name: logrotate-example
files:
  - /app/logs/*.log
  - /app/logs/*/*.log
  - /app/logs/*/*/*.log
  - /app/logs/*/*/*/*.log
mode: daily # 默认 daily， 可以设置为 filesize, 以 256 MB 为单元进行分割
keep: 4 # 保留 4 天，或者 4 个分割文件
# 完成 rotation 之后要执行的命令
dir: /tmp
command:
    - touch
    - xlog.reopen.txt
```

## 环境变量创建单元

如果只需要一个主单元，而且懒得写 `YAML` 文件，可以用以下环境变量创建一个主单元

```
MINIT_MAIN_COMMAND=redis-server /etc/redis.conf
MINIT_MAIN_NAME=main-program
MINIT_MAIN_GROUP=super-main
MINIT_MAIN_ONCE=false
```

## 打开/关闭单元

可以通过环境变量，打开/关闭特定的单元

* `MINIT_ENABLE`, 逗号分隔, 如果值存在，则为 `白名单模式`，只有指定名称的单元会执行
* `MINIT_DISABLE`, 逗号分隔, 如果值存在，则为 `黑名单模式`，除了指定名称外的单元会执行

可以为配置单元设置字段 `group`，然后在上述环境变量使用 `@group` ，设置一组单元的开启和关闭。

## 快速退出

默认情况下，即便是没有 L3 类型任务 (`daemon`, `cron`, `logrotate` 等)，`minit` 也会持续运行，以支撑起容器主进程。

如果要在 `initContainers` 中，或者容器外使用 `minit`，可以将环境变量 `MINIT_QUICK_EXIT` 设置为 `true`

此时，如果没有 L3 类型任务，`minit` 会自动退出

## 资源限制 (ulimit)

**注意，使用此功能可能需要容器运行在高权限 (Privileged) 模式**

使用环境变量 `MINIT_RLIMIT_XXXX` 来设置容器的资源限制，`unlimited` 代表无限制, `-` 表示不修改

比如:

```
MINIT_RLIMIT_NOFILE=unlimited       # 同时设置软硬限制为 unlimited
MINIT_RLIMIT_NOFILE=128:unlimited   # 设置软限制为 128，设置硬限制为 unlimited
MINIT_RLIMIT_NOFILE=128:-           # 设置软限制为 128，硬限制不变
MINIT_RLIMIT_NOFILE=-:unlimited     # 软限制不变，硬限制修改为 unlimited
```

可用的环境变量有:

```
MINIT_RLIMIT_AS
MINIT_RLIMIT_CORE
MINIT_RLIMIT_CPU
MINIT_RLIMIT_DATA
MINIT_RLIMIT_FSIZE
MINIT_RLIMIT_LOCKS
MINIT_RLIMIT_MEMLOCK
MINIT_RLIMIT_MSGQUEUE
MINIT_RLIMIT_NICE
MINIT_RLIMIT_NOFILE
MINIT_RLIMIT_NPROC
MINIT_RLIMIT_RTPRIO
MINIT_RLIMIT_SIGPENDING
MINIT_RLIMIT_STACK
```

## 内核参数 (sysctl)

**注意，使用此功能可能需要容器运行在高权限 (Privileged) 模式**

使用环境变量 `MINIT_SYSCTL` 来写入 `sysctl` 配置项，`minit` 会自动写入 `/proc/sys` 目录下对应的参数

使用 `,` 分隔多个值

比如:

```
MINIT_SYSCTL=vm.max_map_count=262144,vm.swappiness=60
```

## 透明大页 (THP)

**注意，使用此功能可能需要容器运行在高权限 (Privileged) 模式**

使用环境变量 `MINIT_THP` 修改 透明大页配置，可选值为 `never`, `madvise` 和 `always`

## 许可证

Guo Y.K., MIT License

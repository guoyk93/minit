# minit

[![BMC Donate](https://img.shields.io/badge/BMC-Donate-orange)](https://www.buymeacoffee.com/vFa5wfRq6)

一个用 Go 编写的命令行工具，用以在 Docker 容器内启动多个进程，并支持 Cron 类型的进程

## 获取镜像

`registry.cn-shenzhen.aliyuncs.com/landzero/minit`

## 使用方法

使用多阶段 Dockerfile 来从上述镜像地址导入 `minit` 可执行程序

```dockerfile
FROM registry.cn-shenzhen.aliyuncs.com/landzero/minit AS minit

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

**当前支持 `render`, `once`, `daemon` 和 `cron` 四种配置单元**

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

## 打开/关闭单元

可以通过环境变量，打开/关闭特定的单元

* `MINIT_ENABLE`, 逗号分隔, 如果值存在，则为 `白名单模式`，只有指定名称的单元会执行
* `MINIT_DISABLE`, 逗号分隔, 如果值存在，则为 `黑名单模式`，除了指定名称外的单元会执行

可以为配置单元设置字段 `group`，然后在上述环境变量使用 `@group` ，设置一组单元的开启和关闭。

## 许可证

Guo Y.K., MIT License

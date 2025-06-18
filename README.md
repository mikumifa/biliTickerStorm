# 🎫 BiliTickerStorm - B 站分布式抢票

> 本项目使用 **Docker Swarm** 构建，具备良好的分布式扩展能力，可实现多节点协作式抢票。

---

## 📦 项目结构

```bash
.
├── docker-compose.yml            # 兼容 Compose 和 Swarm 的服务定义
├── master.Dockerfile             # ticket-master 构建文件
├── worker.Dockerfile             # ticket-worker 构建文件
├── python.Dockerfile             # gt-python 图形验证服务
├── data/                         # 配置数据目录（挂载给 master）
└── README.md
```

---

## ⚙️ 服务组件说明

| 服务名          | 描述                    | 备注       |
| --------------- | ----------------------- | ---------- |
| `ticket-master` | 主控服务，负责调度任务  | 单实例部署 |
| `ticket-worker` | 抢票 worker，可横向扩展 | 支持多实例 |
| `gt-python`     | 图形验证处理服务        | 单实例部署 |

---

## 🚀 快速部署步骤（Docker Swarm）

### 0. 下载 or Clone 本项目

### 1. 配置 Swarm 集群

> 本项目暂只支持单个 master 节点

参考 https://learn.microsoft.com/zh-cn/virtualization/windowscontainers/manage-containers/swarm-mode

---

### 2. 创建 Overlay 网络

Swarm 集群间通信需要使用 `overlay` 网络：

```bash
docker network create --driver overlay bili-ticket-storm-network
```

---

### 3. 构建镜像

> 后续上传镜像到 Docker Hub

在 Docker Swarm 的 Stack 部署模式下（docker stack deploy），不能使用 build 来构建镜像，必须 先构建好镜像并打 tag，然后用 image: 指定。

```bash
docker build -t ticket-master:latest -f master.Dockerfile .
docker build -t ticket-worker:latest -f worker.Dockerfile .
docker build -t gt-python:latest -f python.Dockerfile .
```

---

### 4. 部署服务栈（Stack）

> 在 master 节点运行，可以在 docker-compose-swarm.ym 中更改相应配置

```bash
docker stack deploy -c docker-compose-swarm.yml bli-ticker-storm
```

> `ticket-system` 是 Stack 名称，服务会注册为 `ticket-system_ticket-master` 等。

---

## 📂 配置说明

将抢票配置文件放置在 `./data/` 目录下，会自动挂载至 master 容器 `/app/data`

抢票配置为 biliTickerBuy 生成的配置文件 https://github.com/mikumifa/biliTickerBuy

---

## 📌 环境变量

### ticket-worker 支持：

| 环境变量名          | 说明                 |
| ------------------- | -------------------- |
| `PUSHPLUS_TOKEN`    | plusplus 推送配置    |
| `TICKET_INTERVAL`   | 抢票间隔秒数（可选） |
| `TICKET_TIME_START` | 定时启动时间（可选） |

---

## 📄 License

[MIT License](LICENSE)

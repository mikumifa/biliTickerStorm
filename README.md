
# 🎫 BiliTickerStorm


## ⚙️ 服务说明

| 服务名             | 说明              | 备注    |
| --------------- | --------------- | ----- |
| `ticket-master` | 主控服务，负责调度任务     | 单实例部署 |
| `ticket-worker` | 抢票 worker，可横向扩展 | 支持多实例 |
| `gt-python`     | 图形验证码处理服务       | 单实例部署 |

---

## 🚀 快速部署步骤

<details> <summary><strong>📦 远程仓库安装（推荐）</strong></summary>

```bash
helm repo add bili-ticker-storm https://mikumifa.github.io/biliTickerStorm/
helm repo update
```

### 2. 安装 Chart

```bash
# 如果使用本地 Chart 目录
helm install bili-ticker-storm bili-ticker-storm/bili-ticker-storm \
  --set hostDataPath=/your/host/data/path \
  --set ticketWorker.pushplusToken="your_token" \
  --set ticketWorker.ticketInterval="300" \
  --set ticketWorker.ticketTimeStart="2025-05-20T13:14"
  
```

> - `hostDataPath` 是抢票配置文件目录，挂载给 `ticket-master` 容器用。
> - `ticketWorker.pushplusToken` 是plusplus 推送配置，设置后可以接收抢票结果通知。
> - `ticketWorker.ticketInterval` 是抢票间隔秒数，默认 300 秒。
> - `ticketWorker.ticketTimeStart` 是定时启动时间，格式为 `2025-05-20T13:14`，可选。

### 3. 升级 Chart

```bash
helm upgrade bili-ticker-storm bili-ticker-storm/bili-ticker-storm --reuse-values \
  --set ticketWorker.ticketInterval="600"
```
---
</details> 
<details> <summary><strong>📦 本地 Chart 安装</strong></summary>


### 1. 安装 Chart

```bash
# 克隆仓库
git clone https://github.com/mikumifa/biliTickerStorm
# 使用本地 Chart 包
helm install bili-ticker-storm bili-ticker-storm/bili-ticker-storm \
  --set hostDataPath=/your/host/data/path \
  --set ticketWorker.pushplusToken="your_token" \
  --set ticketWorker.ticketInterval="300" \
  --set ticketWorker.ticketTimeStart="2025-05-20T13:14"
```
### 2. 升级 Chart

```bash
helm upgrade bili-ticker-storm ./helm --reuse-values
```
</details>
<details>
<summary><strong>📌 通用命令</strong></summary>

### ⏹ 卸载
```bash
helm uninstall bili-ticker-storm
```
</details>

## 📄 License

[MIT License](LICENSE)


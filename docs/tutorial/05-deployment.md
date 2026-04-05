# 部署指南

## 方式一：Docker Compose（推荐）

### 前置条件
- Docker 和 Docker Compose 已安装
- 至少4GB内存

### 步骤

```bash
# 1. 创建环境变量文件
cat > .env << EOF
MINIMAX_API_KEY=your_api_key_here
MINIMAX_API_BASE=https://api.minimax.chat/v1
LLM_MODEL=MiniMax-M2.7
EOF

# 2. 启动所有服务
docker-compose up -d

# 3. 查看状态
docker-compose ps

# 4. 访问
# Python后端: http://localhost:8000
# Java后端:   http://localhost:8080
# Go后端:     http://localhost:8082
# 前端:       http://localhost:3000
```

### 停止和清理

```bash
docker-compose down        # 停止
docker-compose down -v     # 停止并删除数据卷
```

## 方式二：本地直接运行

### Python

```bash
cd python
python -m venv venv && source venv/bin/activate
pip install -r requirements.txt
cp .env.example .env  # 编辑配置
python -m uvicorn src.api.main:app --host 0.0.0.0 --port 8000 --reload
```

### Java

```bash
cd java
mvn spring-boot:run
```

### Go

```bash
cd golang
export MINIMAX_API_KEY=your_key
go run cmd/main.go
```

### 前端

```bash
cd frontend
npm install
npm run dev
```

## 方式三：云部署

### 阿里云/AWS
1. 使用云服务器（2核4G以上）
2. 安装Docker
3. 使用Docker Compose部署
4. 配置Nginx反向代理

### Vercel（仅前端）
1. Fork项目到自己的GitHub
2. 在Vercel导入frontend目录
3. 设置环境变量 `NEXT_PUBLIC_API_BASE`

## 环境变量说明

| 变量 | 说明 | 默认值 |
|------|------|-------|
| `MINIMAX_API_KEY` | MiniMax API密钥 | （必填） |
| `MINIMAX_API_BASE` | API基础URL | https://api.minimax.chat/v1 |
| `LLM_MODEL` | 模型名称 | MiniMax-M2.7 |
| `APP_PORT` | 服务端口 | 8000/8080/8082 |

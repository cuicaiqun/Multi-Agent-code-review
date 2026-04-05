# 多Agent智能合同审查系统

> Multi-Agent Intelligent Contract Review System

[![Python](https://img.shields.io/badge/Python-3.11+-blue.svg)](https://python.org)
[![Java](https://img.shields.io/badge/Java-17+-orange.svg)](https://openjdk.org)
[![Go](https://img.shields.io/badge/Go-1.21+-cyan.svg)](https://go.dev)
[![License](https://img.shields.io/badge/License-Apache%202.0-green.svg)](LICENSE)

一个面向企业级的多Agent智能合同审查系统，提供 **Python / Java / Go** 三种语言实现，配套完整的面试准备材料。从零开始，助你拿下大厂Offer。

## 系统架构

```
┌─────────────────────────────────────────────────────────────┐
│                        前端界面 (Next.js)                      │
│              合同上传 → 审查进度 → 结果展示 → 人工审核            │
└──────────────────────────┬──────────────────────────────────┘
                           │ REST API
┌──────────────────────────▼──────────────────────────────────┐
│                     多Agent流水线引擎                          │
│  ┌──────────┐  ┌──────────┐  ┌──────────┐  ┌──────────┐   │
│  │ 条款提取  │→│ 风险识别  │→│ 合规检查  │→│ 修改建议  │   │
│  │  Agent   │  │  Agent   │  │  Agent   │  │  Agent   │   │
│  └──────────┘  └──────────┘  └──────────┘  └──────────┘   │
│                    ↕ 人机协同 (Human-in-the-Loop)             │
└──────────────────────────┬──────────────────────────────────┘
                           │
┌──────────────────────────▼──────────────────────────────────┐
│                       基础设施层                               │
│    向量数据库 (FAISS/Milvus)  │  规则引擎  │  大模型API         │
└─────────────────────────────────────────────────────────────┘
```

## 核心特性

- **四Agent流水线**：条款提取 → 风险识别 → 合规检查 → 修改建议
- **三层校验架构**：规则硬校验 + Agent编排调度 + 大模型精审
- **人机协同**：关键决策保留人工判断，支持审核/驳回/修改
- **法律NLP**：条款分类、命名实体识别、语义相似度计算
- **规则引擎**：Python DSL / Java Drools / Go 自研规则匹配
- **版本对比**：合同条款级语义diff，Track Changes格式输出
- **三语言实现**：Python (LangGraph) / Java (Spring AI) / Go (Eino)

## 技术栈

| 语言 | Agent框架 | Web框架 | 规则引擎 | 向量数据库 |
|------|----------|---------|---------|-----------|
| Python | LangGraph | FastAPI | 自定义DSL | FAISS/ChromaDB |
| Java | Spring AI Alibaba | Spring Boot | Drools | - |
| Go | Eino (字节) | Hertz | 自研 | Milvus |

## 快速开始

### Python 版本（推荐入门）

```bash
cd python
pip install -r requirements.txt
cp .env.example .env  # 配置API Key
python -m src.api.main
```

### Java 版本

```bash
cd java
mvn spring-boot:run
```

### Go 版本

```bash
cd golang
go run cmd/main.go
```

### Docker 一键部署

```bash
docker-compose up -d
```

## 项目结构

```
├── python/              # Python版本 (LangGraph + FastAPI)
├── java/                # Java版本 (Spring AI Alibaba + Drools)
├── golang/              # Go版本 (Eino + Hertz)
├── frontend/            # Next.js前端
├── docs/                # 文档
│   ├── interview/       # 面试准备材料
│   ├── tutorial/        # 从0到1教程
│   ├── code-walkthrough/# 代码讲解
│   └── api/             # API文档
├── sample-contracts/    # 示例合同
└── docker-compose.yml   # 一键部署
```

## 面试准备

本项目配套完整的面试准备材料，适合求职者使用：

- [简历模板](docs/interview/resume-template.md) - 项目经历怎么写
- [STAR法话术](docs/interview/star-method.md) - 面试怎么讲项目
- [八股文题库](docs/interview/eight-part-essay.md) - 技术面试高频题
- [项目QA](docs/interview/project-qa.md) - 面试官常问的问题

## 教程

- [从0开始](docs/tutorial/01-getting-started.md)
- [Python实现详解](docs/tutorial/02-python-impl.md)
- [Java实现详解](docs/tutorial/03-java-impl.md)
- [Go实现详解](docs/tutorial/04-go-impl.md)
- [部署指南](docs/tutorial/05-deployment.md)

## 参考项目

- [Fan-Luo/multi-agent-contract-platform](https://github.com/Fan-Luo/multi-agent-contract-platform) — 企业级合同审查多Agent平台
- [alibaba/spring-ai-alibaba](https://github.com/alibaba/spring-ai-alibaba) — 阿里Java Agentic AI框架
- [cloudwego/eino](https://github.com/cloudwego/eino) — 字节跳动Go Agent框架
- [petrosrapto/PAKTON](https://github.com/petrosrapto/PAKTON) — EMNLP 2025 法律多Agent QA框架

## License

[Apache License 2.0](LICENSE)

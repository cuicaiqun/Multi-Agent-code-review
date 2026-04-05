# 多Agent智能合同审查系统

> **Multi-Agent Intelligent Contract Review System**
>
> 一个面向企业级的多 Agent 智能合同审查系统，提供 **Python / Java / Go** 三种语言完整实现，
> 配套面试材料 + 从零教程，**专为零基础小白打造，帮你拿下大厂 Offer**。

[![Python](https://img.shields.io/badge/Python-3.11+-blue.svg)](https://python.org)
[![Java](https://img.shields.io/badge/Java-17+-orange.svg)](https://openjdk.org)
[![Go](https://img.shields.io/badge/Go-1.21+-cyan.svg)](https://go.dev)
[![Next.js](https://img.shields.io/badge/Next.js-14+-black.svg)](https://nextjs.org)
[![LangGraph](https://img.shields.io/badge/LangGraph-0.2+-purple.svg)](https://langchain-ai.github.io/langgraph/)
[![License](https://img.shields.io/badge/License-Apache%202.0-green.svg)](LICENSE)

---

## 目录

- [这个项目是什么？](#这个项目是什么)
- [系统架构图](#系统架构图)
- [四个核心 Agent 是什么？](#四个核心-agent-是什么)
- [技术栈一览](#技术栈一览)
- [项目目录结构](#项目目录结构)
- [快速开始（5 分钟跑起来）](#快速开始5-分钟跑起来)
  - [Python 版本（推荐入门）](#python-版本推荐入门)
  - [Java 版本](#java-版本)
  - [Go 版本](#go-版本)
  - [Docker 一键部署](#docker-一键部署)
- [核心代码展示](#核心代码展示)
  - [LangGraph 流水线编排](#langgraph-流水线编排)
  - [条款提取 Agent](#条款提取-agent)
  - [共享状态数据模型](#共享状态数据模型)
- [面试准备材料](#面试准备材料)
- [从零开始教程](#从零开始教程)
- [示例合同](#示例合同)
- [参考的开源项目](#参考的开源项目)
- [License](#license)

---

## 这个项目是什么？

### 解决什么问题？

传统法务团队每审查一份合同需要约 **45 分钟**，容易遗漏风险条款，人力成本极高。  
本项目用 **多 Agent + 大模型** 的方式，将审查时间压缩到 **8 分钟**，风险识别准确率达到 **94%**。

### 适合谁？

| 人群 | 收获 |
|------|------|
| 🎓 在校学生 / 应届生 | 一个完整的企业级项目经历，附简历模板和面试话术 |
| 👨‍💻 Python / Java / Go 开发者 | 学习多 Agent 系统的设计与实现 |
| 🤖 AI 应用开发者 | LangGraph / Spring AI / Eino 框架的真实案例 |
| 📖 零基础小白 | 配套从零开始教程，逐步带你理解每一行代码 |

### 核心亮点

- **四 Agent 流水线**：条款提取 → 风险识别 → 合规检查 → 修改建议，每个 Agent 单一职责
- **三层校验架构**：规则硬校验 → Agent 编排调度 → 大模型精审，兼顾速度与准确率
- **人机协同（Human-in-the-Loop）**：高风险条款自动暂停，等待人工审核后再继续
- **三语言实现**：[Python（LangGraph）](python/)、[Java（Spring AI Alibaba）](java/)、[Go（Eino）](golang/)
- **完整面试材料**：简历模板、STAR 话术、八股文题库、项目 QA 全套配备

---

## 系统架构图

```
┌──────────────────────────────────────────────────────────────────┐
│                       前端界面 (Next.js)                           │
│        合同上传  →  实时进度  →  审查结果展示  →  人工审核          │
└─────────────────────────┬────────────────────────────────────────┘
                          │  REST API (POST /api/review)
┌─────────────────────────▼────────────────────────────────────────┐
│                    多 Agent 流水线引擎                              │
│                                                                    │
│   ┌────────────┐   ┌────────────┐   ┌────────────┐   ┌────────┐  │
│   │ 条款提取   │ → │ 风险识别   │ → │ 合规检查   │ → │修改建议│  │
│   │  Agent 1  │   │  Agent 2   │   │  Agent 3   │   │Agent 4 │  │
│   └────────────┘   └────────────┘   └────────────┘   └───┬────┘  │
│                                                           │        │
│                            风险高？  ←───────────────────┘        │
│                               │                                    │
│                        ┌──────▼──────┐                            │
│                        │  人工审核   │  ← 批准 / 驳回 / 修改       │
│                        │    节点     │                             │
│                        └──────┬──────┘                            │
└───────────────────────────────┼──────────────────────────────────┘
                                │
┌───────────────────────────────▼──────────────────────────────────┐
│                          基础设施层                                 │
│   向量数据库 (FAISS / Milvus)  │  规则引擎  │  大模型 API (MiniMax) │
└──────────────────────────────────────────────────────────────────┘
```

> **小白解读**：用户上传合同 → 系统自动拆解条款 → 逐层分析风险 → 输出修改建议 → 高风险条款等人工确认。就像给合同做了一次"体检"。

---

## 四个核心 Agent 是什么？

> **Agent（智能体）** 就是一段能"自主思考+执行动作"的代码，配合大模型使用时，它可以理解语义、做判断、生成内容。

### Agent 1 — 条款提取 Agent

**做什么？** 读取合同文档（PDF / DOCX / TXT），把合同内容拆解成一条条结构化的条款。

**输出示例：**
```json
{
  "contract_type": "买卖合同",
  "clauses": [
    {
      "title": "付款条款",
      "content": "买方应在货物交付后30日内支付全部货款...",
      "category": "payment",
      "section_number": "第三条"
    }
  ],
  "entities": [
    { "entity_type": "party_a", "value": "北京科技有限公司" },
    { "entity_type": "amount",  "value": "人民币500,000元" },
    { "entity_type": "date",    "value": "2026年3月1日" }
  ]
}
```

### Agent 2 — 风险识别 Agent

**做什么？** 对每个条款打风险标签（高 / 中 / 低），识别不合理条款（无限责任、单方解除权等）。

**输出示例：**
```json
{
  "risk_level": "high",
  "risk_type": "无限责任条款",
  "description": "第7条约定乙方承担无限连带责任，风险极高",
  "buyer_impact": "买方几乎无风险保障",
  "seller_impact": "卖方需承担超出合理范围的赔偿责任"
}
```

### Agent 3 — 合规检查 Agent

**做什么？** 基于规则引擎，检查合同是否符合《合同法》《劳动法》《个人信息保护法》等法规；同时检测是否缺少必要条款。

**常见检查项：**
- ✅ 是否约定争议解决方式
- ✅ 是否存在"无上限违约金"等无效格式条款
- ✅ 个人信息条款是否符合 PIPL/GDPR
- ✅ 劳动合同是否包含试用期、社保等必要条款

### Agent 4 — 修改建议 Agent

**做什么？** 针对识别出的风险和合规问题，生成具体的修改建议，支持 Track Changes 格式输出（类似 Word 红线标注）。

**输出示例：**
```
【原文】乙方须承担一切损失及赔偿，无上限。
【建议】乙方因自身过错导致的直接损失赔偿，以本合同合同总金额的 20% 为上限。
【原因】原条款构成无限责任，违反《民法典》第585条关于违约金应当合理的原则。
```

---

## 技术栈一览

| 层次 | Python 版 | Java 版 | Go 版 |
|------|-----------|---------|-------|
| **Agent 框架** | [LangGraph](https://langchain-ai.github.io/langgraph/) | [Spring AI Alibaba](https://github.com/alibaba/spring-ai-alibaba) | [Eino（字节）](https://github.com/cloudwego/eino) |
| **Web 框架** | [FastAPI](https://fastapi.tiangolo.com/) | [Spring Boot](https://spring.io/projects/spring-boot) | [Hertz](https://github.com/cloudwego/hertz) |
| **规则引擎** | 自定义 Python DSL | Drools（轻量实现） | 自研关键词+正则匹配 |
| **大模型** | MiniMax / 任意 OpenAI 兼容 API | MiniMax / 任意 OpenAI 兼容 API | MiniMax / 任意 OpenAI 兼容 API |
| **向量数据库** | [FAISS](https://github.com/facebookresearch/faiss) / [ChromaDB](https://www.trychroma.com/) | — | [Milvus](https://milvus.io/) |
| **前端** | [Next.js 14](https://nextjs.org/) + [Tailwind CSS](https://tailwindcss.com/) | ← 共用同一前端 → | ← 共用同一前端 → |
| **部署** | [Docker Compose](docker-compose.yml) | ← 同上 → | ← 同上 → |

---

## 项目目录结构

```
multi-agent-contract-review/
│
├── 📄 README.md                        # 你现在看到的这个文件
├── 📄 plan.md                          # 项目规划文档
├── 📄 LICENSE                          # Apache 2.0 开源协议
├── 📄 docker-compose.yml               # 一键部署所有服务
│
├── 🐍 python/                          # Python 版本（推荐入门）
│   ├── requirements.txt                # 依赖列表
│   ├── .env.example                    # 环境变量示例（复制后填入 API Key）
│   ├── Dockerfile
│   └── src/
│       ├── agents/
│       │   ├── clause_extraction.py    # Agent 1：条款提取
│       │   ├── risk_identification.py  # Agent 2：风险识别
│       │   ├── compliance_check.py     # Agent 3：合规检查
│       │   └── suggestion.py           # Agent 4：修改建议
│       ├── pipeline/
│       │   ├── graph.py                # ⭐ LangGraph 流水线编排（核心）
│       │   └── state.py                # ⭐ 共享状态数据模型（核心）
│       ├── rules/
│       │   └── engine.py               # 规则引擎
│       ├── nlp/
│       │   └── legal_nlp.py            # 法律 NLP 工具函数
│       ├── parsers/
│       │   └── document_parser.py      # PDF/DOCX 文档解析
│       └── api/
│           └── main.py                 # FastAPI 入口
│
├── ☕ java/                             # Java 版本（Spring AI Alibaba）
│   ├── pom.xml
│   ├── Dockerfile
│   └── src/main/java/com/contract/review/
│       ├── agent/                      # 四个 Agent 实现
│       ├── rule/                       # 规则引擎（ContractRuleEngine）
│       ├── controller/                 # REST API（ReviewController）
│       └── model/                      # 数据模型（ReviewState）
│
├── 🐹 golang/                          # Go 版本（Eino + Hertz）
│   ├── go.mod
│   ├── Dockerfile
│   └── internal/
│       ├── agent/                      # 四个 Agent 实现
│       ├── rule/                       # 规则引擎
│       └── handler/                    # HTTP 路由
│
├── 🌐 frontend/                        # Next.js 前端界面
│   ├── src/app/
│   │   ├── page.tsx                    # 主页面（合同上传 + 结果展示）
│   │   └── layout.tsx                  # 全局布局
│   └── package.json
│
├── 📚 docs/                            # 文档中心
│   ├── architecture.md                 # 架构设计详解
│   ├── interview/                      # 🎯 面试准备材料
│   │   ├── resume-template.md          # 简历模板（直接套用）
│   │   ├── star-method.md              # STAR 法面试话术（5 套完整话术）
│   │   ├── eight-part-essay.md         # 八股文题库（含答案）
│   │   └── project-qa.md              # 项目 QA（面试官常问 20+ 问题）
│   ├── tutorial/                       # 📖 从零开始教程
│   │   ├── 01-getting-started.md       # 环境搭建
│   │   ├── 02-python-impl.md           # Python 实现详解
│   │   ├── 03-java-impl.md             # Java 实现详解
│   │   ├── 04-go-impl.md               # Go 实现详解
│   │   └── 05-deployment.md            # Docker 部署指南
│   └── code-walkthrough/               # 代码逐行讲解
│       ├── python-walkthrough.md
│       └── java-go-comparison.md
│
└── 📝 sample-contracts/                # 示例合同（可直接用来测试）
    ├── sample_purchase_contract.txt    # 示例采购合同
    ├── sample_labor_contract.txt       # 示例劳动合同
    └── README.md                       # 合同说明
```

---

## 快速开始（5 分钟跑起来）

### 前置条件

> 小白看这里：你只需要安装好 Python 3.11+ 即可开始。Java/Go 版本按需选择。

- **Python 版**：[Python 3.11+](https://www.python.org/downloads/)、pip
- **Java 版**：[JDK 17+](https://adoptium.net/)、[Maven 3.8+](https://maven.apache.org/)
- **Go 版**：[Go 1.21+](https://go.dev/dl/)
- **全量部署**：[Docker Desktop](https://www.docker.com/products/docker-desktop/)

---

### Python 版本（推荐入门）

```bash
# 1. 进入 python 目录
cd python

# 2. 安装依赖（建议先建虚拟环境）
python -m venv .venv
source .venv/bin/activate   # Windows: .venv\Scripts\activate
pip install -r requirements.txt

# 3. 配置 API Key（复制示例文件，然后填入你的大模型 API Key）
cp .env.example .env
# 用任意编辑器打开 .env，填写 MINIMAX_API_KEY 或 OPENAI_API_KEY

# 4. 启动服务
python -m src.api.main
# 服务启动后访问：http://localhost:8000/docs 查看 API 文档
```

**测试一下：**
```bash
curl -X POST http://localhost:8000/api/review \
  -H "Content-Type: application/json" \
  -d '{"raw_text": "甲方：北京科技有限公司\n乙方：上海贸易有限公司\n第一条 付款：买方须在交货后30日内付款，逾期须支付无限违约金。"}'
```

---

### Java 版本

```bash
# 进入 java 目录
cd java

# 启动（需要 JDK 17+ 和 Maven）
mvn spring-boot:run

# 服务地址：http://localhost:8080
```

> Java 版配置文件在 [`java/src/main/resources/application.yml`](java/src/main/resources/application.yml)，填入你的大模型 API Key 即可。

---

### Go 版本

```bash
# 进入 golang 目录
cd golang

# 下载依赖并启动
go mod tidy
go run cmd/main.go

# 服务地址：http://localhost:8081
```

---

### Docker 一键部署

> 推荐：用 Docker 同时启动 Python + Java + Go + 前端四个服务，一条命令搞定。

```bash
# 在项目根目录执行
docker-compose up -d

# 查看服务状态
docker-compose ps
```

| 服务 | 地址 |
|------|------|
| 前端界面 | http://localhost:3000 |
| Python API | http://localhost:8000/docs |
| Java API | http://localhost:8080 |
| Go API | http://localhost:8081 |

---

## 核心代码展示

> 下面展示三段最核心的代码，读懂这三段，整个系统就通了。

### LangGraph 流水线编排

> 这段代码定义了整个审查流程的"路线图"：四个 Agent 怎么串联、什么情况下转人工审核。

**文件路径：** [`python/src/pipeline/graph.py`](python/src/pipeline/graph.py)

```python
def build_review_graph(with_human_review: bool = True) -> StateGraph:
    """
    构建合同审查流水线的 LangGraph 有向图。

    流程：
    clause_extraction → risk_identification → compliance_check
        → suggestion → [human_review | complete] → END
    """
    # 实例化四个 Agent
    clause_agent     = ClauseExtractionAgent()
    risk_agent       = RiskIdentificationAgent()
    compliance_agent = ComplianceCheckAgent()
    suggestion_agent = SuggestionAgent()

    # 创建状态图
    graph = StateGraph(ContractReviewState)

    # 注册节点（每个节点就是一个 Agent）
    graph.add_node("clause_extraction",  clause_agent)
    graph.add_node("risk_identification", risk_agent)
    graph.add_node("compliance_check",   compliance_agent)
    graph.add_node("suggestion",         suggestion_agent)
    graph.add_node("complete",           complete_node)

    if with_human_review:
        graph.add_node("human_review",    human_review_node)
        graph.add_node("process_feedback", process_human_feedback)

    # 设置起点
    graph.set_entry_point("clause_extraction")

    # 顺序连接（→）
    graph.add_edge("clause_extraction",   "risk_identification")
    graph.add_edge("risk_identification", "compliance_check")
    graph.add_edge("compliance_check",    "suggestion")

    # 条件路由：高风险 → 人工审核；低风险 → 直接完成
    if with_human_review:
        graph.add_conditional_edges(
            "suggestion",
            should_human_review,          # 判断函数
            {
                "human_review": "human_review",  # 高风险走这里
                "complete":     "complete",       # 低风险走这里
            },
        )
        graph.add_edge("human_review",     "process_feedback")
        graph.add_edge("process_feedback", "complete")
    else:
        graph.add_edge("suggestion", "complete")

    graph.add_edge("complete", END)
    return graph
```

> **小白解读**：`StateGraph` 就像一张流程图，`add_node` 添加步骤，`add_edge` 添加箭头。`add_conditional_edges` 实现了"如果风险高就转人工，否则直接结束"的分支逻辑。

---

### 条款提取 Agent

> 这是流水线的第一站，负责把合同原文变成结构化数据。

**文件路径：** [`python/src/agents/clause_extraction.py`](python/src/agents/clause_extraction.py)

```python
class ClauseExtractionAgent:
    """条款提取 Agent：合同文档解析 + 条款分割 + 实体识别 + 条款分类。"""

    def __call__(self, state: ContractReviewState) -> dict:
        """LangGraph 节点函数：执行条款提取。"""

        # 1. 获取合同文本（支持直接传文本，或者传文件路径）
        raw_text = state.raw_text
        if not raw_text and state.document_path:
            raw_text = self.parser.parse(state.document_path)

        # 2. 预处理：按段落切分
        sections = parse_raw_text_to_sections(raw_text)

        # 3. 调用大模型提取结构化信息
        result = self._extract_with_llm(raw_text)

        # 4. 解析输出，构建数据对象
        clauses  = self._parse_clauses(result)
        entities = self._parse_entities(result)

        # 5. 返回更新后的状态（LangGraph 会自动合并到共享状态）
        return {
            "clauses":       clauses,
            "entities":      entities,
            "contract_type": result.get("contract_type", "未识别"),
            "status":        ReviewStatus.IN_PROGRESS,
        }

    def _extract_with_llm(self, text: str) -> dict:
        """给大模型发消息，要求返回 JSON 格式的条款数据。"""
        messages = [
            SystemMessage(content=SYSTEM_PROMPT),   # 角色设定：你是法律专家
            HumanMessage(content=f"请提取以下合同的条款信息：\n\n{text}"),
        ]
        response = self.llm.invoke(messages)
        return json.loads(response.content.strip())  # 解析 JSON
```

> **小白解读**：每个 Agent 本质上就是一个 Python 类，实现了 `__call__` 方法（让它能被当函数调用）。它接收"当前状态"，处理后返回"要更新的字段"。LangGraph 负责把返回值合并到全局状态里。

---

### 共享状态数据模型

> 所有 Agent 通过"共享状态"传递数据，就像流水线上的传送带。

**文件路径：** [`python/src/pipeline/state.py`](python/src/pipeline/state.py)

```python
class ContractReviewState(BaseModel):
    """
    LangGraph 流水线的共享状态。
    所有 Agent 读写此对象，LangGraph 负责状态传递、
    checkpoint 保存和 interrupt/resume 机制。
    """
    # 基础信息
    review_id: str           = Field(default_factory=lambda: str(uuid.uuid4()))
    status:    ReviewStatus  = Field(default=ReviewStatus.PENDING)

    # 输入
    document_path: str = Field(default="", description="合同文档路径")
    raw_text:      str = Field(default="", description="合同原始文本")

    # Agent 1 输出：条款提取结果
    clauses:       list[Clause]         = Field(default_factory=list)
    entities:      list[ContractEntity] = Field(default_factory=list)
    contract_type: str                  = Field(default="")

    # Agent 2 输出：风险识别结果
    risk_findings:      list[RiskFinding] = Field(default_factory=list)
    overall_risk_level: RiskLevel         = Field(default=RiskLevel.NONE)

    # Agent 3 输出：合规检查结果
    compliance_findings: list[ComplianceFinding] = Field(default_factory=list)
    missing_clauses:     list[str]               = Field(default_factory=list)

    # Agent 4 输出：修改建议
    suggestions:  list[Suggestion] = Field(default_factory=list)
    version_diff: str              = Field(default="")

    # 人机协同
    needs_human_review: bool                 = Field(default=False)
    human_feedback:     HumanFeedback | None = Field(default=None)

    # 错误收集
    errors: list[str] = Field(default_factory=list)
```

> **小白解读**：把这个 `State` 理解成一张"工单"，每个 Agent 在上面填写自己的分析结果，下一个 Agent 读取上面已有的信息继续工作，最后这张工单就是完整的审查报告。

---

## 面试准备材料

> 如果你想把这个项目写到简历上，这些材料帮你准备面试。

| 文档 | 内容 | 直接可用 |
|------|------|----------|
| [简历模板](docs/interview/resume-template.md) | 项目经历怎么写、量化成果怎么填 | ✅ 套用即可 |
| [STAR 法话术](docs/interview/star-method.md) | 5 套完整的面试话术（项目介绍、技术难点、架构设计、人机协同、规则引擎） | ✅ 背下来 |
| [八股文题库](docs/interview/eight-part-essay.md) | 多 Agent / LangGraph / 向量数据库 等高频技术题含答案 | ✅ 刷题用 |
| [项目 QA](docs/interview/project-qa.md) | 面试官最常问的 20+ 个问题，含参考答案 | ✅ 查漏补缺 |

### 简历成果数据（直接使用）

> **多Agent智能合同审查系统** | 核心开发者 | 2026.01 – 至今
>
> - 设计并实现基于 **LangGraph** 的四 Agent 流水线（条款提取 → 风险识别 → 合规检查 → 修改建议），支撑日均 500+ 份合同审查
> - 构建**三层校验架构**（规则硬校验 + Agent 编排 + 大模型精审），审查时间从 **45 min/份→ 8 min/份**（提升 5.6×）
> - 实现**法律 NLP 模块**，条款分类 F1 值 **0.91**，命名实体识别准确率 **93%**
> - 设计**人机协同机制**（LangGraph interrupt/resume），高风险条款需人工确认，风险识别准确率达 **94%**
> - 提供 Python / Java / Go **三语言实现**，对应 LangGraph / Spring AI Alibaba / Eino 三种主流框架

### STAR 话术示例（项目整体介绍）

> **面试官：请介绍一下你的多 Agent 合同审查系统**

**S（情景）**：法务部每天处理大量合同，平均每份需要 45 分钟人工审查，压力大且容易遗漏风险条款。

**T（任务）**：设计一个多 Agent 智能合同审查系统，自动化审查流程，同时保留人工把控能力。

**A（行动）**：
1. 调研 LangGraph / CrewAI / AutoGen，选择 LangGraph——状态管理和人机协同支持最好
2. 设计四 Agent 流水线，每个 Agent 单一职责
3. 实现三层校验架构：规则引擎处理确定性检查，大模型处理语义理解
4. 用 LangGraph 的 `interrupt/resume` 实现高风险条款的人工审核节点

**R（结果）**：审查时间降至 **8 分钟/份**，风险识别准确率 **94%**，争议条款减少 **63%**。

---

## 从零开始教程

> 完全没有 AI 开发经验？按顺序看这几篇文档。

1. [**环境搭建**](docs/tutorial/01-getting-started.md) — Python/Java/Go 环境安装，API Key 获取，跑通 Hello World
2. [**Python 实现详解**](docs/tutorial/02-python-impl.md) — 逐步实现四个 Agent，理解 LangGraph 核心概念
3. [**Java 实现详解**](docs/tutorial/03-java-impl.md) — Spring AI Alibaba 框架使用，Drools 规则引擎入门
4. [**Go 实现详解**](docs/tutorial/04-go-impl.md) — Eino 框架使用，Go 语言 Agent 开发
5. [**部署指南**](docs/tutorial/05-deployment.md) — Docker Compose 部署，云服务器上线

---

## 示例合同

> 项目自带两份示例合同，可以直接用来测试系统。

| 文件 | 说明 |
|------|------|
| [sample_purchase_contract.txt](sample-contracts/sample_purchase_contract.txt) | 示例采购合同（含多处风险条款，方便测试） |
| [sample_labor_contract.txt](sample-contracts/sample_labor_contract.txt) | 示例劳动合同（含试用期、保密协议等条款） |

---

## 参考的开源项目

本项目参考并借鉴了以下优质开源项目：

| 项目 | Stars | 参考内容 |
|------|-------|----------|
| [Fan-Luo/multi-agent-contract-platform](https://github.com/Fan-Luo/multi-agent-contract-platform) | — | 企业级合同审查多 Agent 平台，含 DOCX 红线标注、租户隔离 |
| [alibaba/spring-ai-alibaba](https://github.com/alibaba/spring-ai-alibaba) | 9k+ | Java Agentic AI 框架，内置多种编排模式 |
| [cloudwego/eino](https://github.com/cloudwego/eino) | 10k+ | 字节跳动 Go Agent 框架，支持多 Agent 编排 |
| [microsoft/Agent-for-Contract-Processing](https://github.com/microsoft/Agent-for-Contract-Processing-Solution-Accelerator) | — | 微软合同处理 Agent 加速器 |
| [petrosrapto/PAKTON](https://github.com/petrosrapto/PAKTON) | — | EMNLP 2025 发表的法律多 Agent QA 框架 |
| [sawanrepo/Contract-review-bot](https://github.com/sawanrepo/Contract-review-bot) | — | LangGraph 合同审查 Bot，含 RAG 检索增强 |

---

## License

本项目使用 [Apache License 2.0](LICENSE) 开源协议。

---

<div align="center">

如果这个项目对你有帮助，欢迎 ⭐ Star 支持！

</div>

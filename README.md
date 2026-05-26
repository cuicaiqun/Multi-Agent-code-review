```markdown
# Multi-Agent 代码审查系统

> 面向 Python 代码审查场景，构建规则引擎 + LLM 混合检测的多 Agent 系统，实现安全审计与智能修复的全流程自动化。

[![Python 3.11+](https://img.shields.io/badge/Python-3.11+-blue.svg)](https://www.python.org/)
[![LangGraph](https://img.shields.io/badge/LangGraph-0.2+-green.svg)](https://www.langchain.com/langgraph)
[![FastAPI](https://img.shields.io/badge/FastAPI-0.115+-teal.svg)](https://fastapi.tiangolo.com/)
[![License: MIT](https://img.shields.io/badge/License-MIT-yellow.svg)](https://opensource.org/licenses/MIT)

---

## 📖 项目简介

在纯 LLM 代码审查场景中，存在**误报率高、无法直接落地**的问题。本项目通过 **规则引擎 + 大语言模型** 混合检测架构，结合多 Agent 协同工作，实现了安全审计与智能修复的自动化流程。

系统基于 LangGraph 将代码审查拆解为四个串行 Agent：结构解析 → 漏洞检测 → 规范校验 → 重构建议。通过共享状态对象实现数据传递，并引入人机协同机制（HITL）处理高危漏洞，同时设计降级容错策略保证系统稳定性。

## 🎯 核心功能与技术亮点

### 🤖 Multi-Agent 编排

基于 LangGraph 的 `StateGraph` 将代码审查流程拆解为四个专业 Agent：

| Agent | 职责 | 输入 | 输出 |
|-------|------|------|------|
| **结构解析** | AST 解析，提取函数、类、导入，计算圈复杂度 | 原始代码 | `code_units` |
| **漏洞检测** | 规则引擎 + LLM 双重扫描，识别安全漏洞 | `code_units` | `issue_findings`（含 CWE 分类） |
| **规范校验** | 检查 PEP8 及自定义编码规范 | `code_units` | `standard_findings` |
| **重构建议** | 生成可执行的修复方案（diff 格式） | 问题 + 规范 | `suggestions` + `version_diff` |

四个 Agent 通过共享的 `CodeReviewState` 对象传递数据，形成串行流水线，实现从代码输入到修复方案输出的全自动审查。

### 👥 人机协同审核（HITL）

纯 LLM 误报率高，无法直接落地。针对 **高危漏洞（Critical severity）** 触发人工审核闭环：

- 利用 LangGraph 的 `interrupt` 机制暂停流程，等待人工反馈
- 结合 `Checkpoint` 进行状态持久化，支持长流程恢复
- 人工介入率控制在 **15% 以内**

### ⚙️ LLM 工程化与降级容错

- **规则引擎**：自研覆盖 **9 类 CWE** 高危漏洞的规则引擎（基于正则 + DSL），作为第一道防线
- **LLM 补充**：利用 LLM 检测语义逻辑问题盲区
- **降级策略**：LLM 调用失败时自动切换纯规则模式，保证服务可用性
- **输出稳定性**：通过结构化 Prompt + Pydantic 模型约束输出格式

## 📊 评测结果

### 数据集：SecurityEval（121 个 Python 漏洞样本）

| 指标 | 数值 |
|------|------|
| **召回率 (Recall)** | **89.5%** |
| **精确率 (Precision)** | 56.7% |
| **F1** | 69.4% |

**对比纯 LLM 基线**（相同 MiniMax-M2.7 模型）：

- 纯 LLM 检出：15 个漏洞 → 召回率约 78.9%
- 混合方案检出：17 个漏洞 → **提升 11.8%**
- 耗时：22.9 秒（持平）

**按 CWE 类别召回（规则覆盖 7 类）**：

| CWE | 名称 | 召回率 |
|-----|------|--------|
| CWE-078 | 命令注入 | **100%** |
| CWE-089 | SQL 注入 | **100%** |
| CWE-095 | 代码注入 | **100%** |
| CWE-502 | 不安全反序列化 | **100%** |
| CWE-798 | 硬编码密钥 | **100%** |
| CWE-022 | 路径遍历 | 75% |
| CWE-327 | 弱加密 | 75% |

**漏检分析**：2 条漏检（CWE-022、CWE-327）原因是漏洞模式与规则不匹配，规则库待扩展。

**重构 Agent**：在 5 组高危用例中自动生成了 **41 条可执行的修复方案**（diff 格式）。

## 🏗️ 系统架构

```text
┌──────────┐
│ 代码输入  │
└─────┬────┘
      ▼
┌─────────────────────────────────────────────────┐
│            Agent 1: 结构解析 (AST)               │
│   输出: 函数/类/导入/圈复杂度 → code_units        │
└─────────────────────┬───────────────────────────┘
                      ▼
┌─────────────────────────────────────────────────┐
│      Agent 2: 漏洞检测（规则引擎 + LLM）          │
│  ┌─────────────┐      ┌─────────────┐          │
│  │ 规则引擎(9类CWE)│  +  │ LLM深度分析  │          │
│  └─────────────┘      └─────────────┘          │
│  输出: issue_findings (含 severity, CWE)        │
└─────────────────────┬───────────────────────────┘
                      ▼
              ┌───────┴───────┐
              │ 是否为高危漏洞？ │
              └───────┬───────┘
                    │
        ┌───────────┴───────────┐
        ▼                       ▼
   [是] 人机协同(HITL)       [否] 继续
        │                       │
        ▼                       ▼
   ┌──────────┐          ┌──────────────┐
   │人工审核+反馈│          │Agent 3: 规范校验│
   └─────┬────┘          │输出: standard │
         │               │    _findings  │
         └───────────────┴───────┬───────┘
                                 ▼
                    ┌────────────────────────┐
                    │ Agent 4: 重构建议       │
                    │ 输出: suggestions + diff│
                    └───────────┬────────────┘
                                ▼
                        ┌────────────┐
                        │ 修复方案输出 │
                        └────────────┘
```

## 🚀 快速开始

### 环境准备

1. **克隆仓库**

   ```bash
   git clone https://github.com/cuicaiqun-gif/Multi-Agent-code-review.git
   cd Multi-Agent-code-review/python
   ```

2. **安装依赖**

   建议使用 `conda` 或 `venv` 创建 Python 3.11+ 虚拟环境。

   ```bash
   pip install -r requirements.txt
   ```

3. **配置环境变量**

   复制 `.env.example` 为 `.env`，填入 LLM API 密钥（目前支持 MiniMax 兼容接口）。

   ```bash
   cp .env.example .env
   # 编辑 .env 设置 MINIMAX_API_KEY
   ```

### 运行评测（无需 API Key）

纯规则引擎模式，可直接复现召回率数据：

```bash
python eval_securityeval.py
```

预期输出：

```
总样本: 121
TP: 17, FP: 13, FN: 2, TN: 89
精确率: 56.7%, 召回率: 89.5%, F1: 69.4%
```

### 启动 API 服务

```bash
uvicorn src.api.main:app --reload
```

访问 `http://127.0.0.1:8000/docs` 查看 Swagger 文档。

### 发起代码审查请求

```bash
curl -X POST http://127.0.0.1:8000/review \
  -H "Content-Type: application/json" \
  -d '{"code": "import sqlite3\nconn = sqlite3.connect(\"db\")\ncursor = conn.cursor()\ncursor.execute(\"SELECT * FROM users WHERE id = \" + user_input)"}'
```

## 💡 设计哲学

- **混合检测优先**：规则引擎保证确定性漏洞的检出，LLM 补充语义盲区，两者互补而非替代。
- **可落地的人机协同**：只在真正高危且易误报的环节引入人工，控制成本。
- **降级容错能力**：LLM 不可用时系统仍能提供基础服务（纯规则模式）。
- **可评测、可迭代**：提供 SecurityEval 评测脚本，每次改动可量化召回率与精确率变化。

## 📁 项目结构

```
python/
├── src/
│   ├── agents/
│   │   ├── issue_detection.py      # 漏洞检测 Agent + 9 类 CWE 规则定义
│   │   ├── structure_extraction.py # AST 结构解析
│   │   ├── standard_check.py       # 编码规范检查
│   │   └── refactor_suggestion.py  # 重构建议生成
│   ├── rules/
│   │   └── engine.py               # 规则引擎（支持 AND/OR 组合、优先级）
│   ├── pipeline/
│   │   ├── graph.py                # LangGraph 流水线编排
│   │   └── state.py                # CodeReviewState 共享状态
│   ├── api/
│   │   └── main.py                 # FastAPI 服务入口
│   └── config.py                   # 配置管理（模型、温度、降级开关）
├── eval_securityeval.py            # 评测脚本
├── requirements.txt
└── .env.example
```

## 🔮 未来计划

- [ ] 扩展规则引擎，覆盖更多 CWE（如 CWE-022、CWE-327 盲区）
- [ ] 支持 Java / Go 代码审查
- [ ] 集成 CI/CD 插件（GitHub Action / GitLab CI）
- [ ] 基于人工反馈数据的主动学习，优化 LLM 提示词

## 📜 许可证

本项目采用 [MIT 许可证](LICENSE)。

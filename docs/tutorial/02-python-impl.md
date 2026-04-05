# Python版实现详解

## 技术栈

- **LangGraph** — Agent编排框架（LangChain团队出品）
- **FastAPI** — 高性能异步Web框架
- **Pydantic** — 数据模型和验证
- **MiniMax M2.7** — 大语言模型（通过OpenAI兼容接口）

## Step 1: 定义共享状态

所有Agent通过共享状态通信。`pipeline/state.py`定义了核心数据模型：

```python
class ContractReviewState(BaseModel):
    """流水线共享状态"""
    review_id: str
    raw_text: str               # 输入
    clauses: list[Clause]       # Agent 1 输出
    risk_findings: list[...]    # Agent 2 输出
    compliance_findings: ...    # Agent 3 输出
    suggestions: list[...]      # Agent 4 输出
    needs_human_review: bool    # 人机协同标志
```

**设计要点**：每个Agent只写自己的字段，读取前面Agent的输出。

## Step 2: 实现Agent

每个Agent是一个可调用类，`__call__`方法是入口：

```python
class ClauseExtractionAgent:
    def __call__(self, state: ContractReviewState) -> dict:
        # 1. 文档解析
        # 2. 调用LLM
        # 3. 返回更新字段
        return {"clauses": [...], "entities": [...]}
```

**关键**：返回dict而不是完整state，LangGraph会自动merge。

## Step 3: 构建LangGraph流水线

```python
graph = StateGraph(ContractReviewState)
graph.add_node("clause_extraction", clause_agent)
graph.add_node("risk_identification", risk_agent)
graph.add_node("compliance_check", compliance_agent)
graph.add_node("suggestion", suggestion_agent)

graph.set_entry_point("clause_extraction")
graph.add_edge("clause_extraction", "risk_identification")
# ... 更多边
```

## Step 4: 暴露API

```python
@app.post("/api/v1/review")
async def create_review(request: ReviewRequest):
    pipeline = create_review_pipeline()
    result = await pipeline.ainvoke(state.model_dump(), config)
    return ReviewResponse(...)
```

## Step 5: 运行测试

```bash
cd python
python -m src.api.main
# 访问 http://localhost:8000/docs 查看Swagger文档
```

## 练习题

1. 给规则引擎添加一条新规则：检测"竞业禁止"条款
2. 修改条款提取Agent，支持英文合同
3. 在风险识别Agent中实现并行调用（规则引擎和LLM同时执行）

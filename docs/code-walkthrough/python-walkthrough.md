# Python版代码讲解

## 代码结构总览

```
python/
├── src/
│   ├── config.py              # 应用配置和LLM客户端
│   ├── agents/                # 四个核心Agent
│   │   ├── clause_extraction.py   # 条款提取
│   │   ├── risk_identification.py # 风险识别
│   │   ├── compliance_check.py    # 合规检查
│   │   └── suggestion.py          # 修改建议
│   ├── pipeline/              # LangGraph流水线
│   │   ├── state.py           # 共享状态定义
│   │   └── graph.py           # 图编排
│   ├── rules/engine.py        # 规则引擎
│   ├── nlp/legal_nlp.py       # 法律NLP
│   ├── parsers/document_parser.py # 文档解析
│   └── api/main.py            # FastAPI接口
└── requirements.txt
```

---

## 1. 共享状态 — `pipeline/state.py`

**设计思路**：所有Agent通过共享状态通信，这是LangGraph的核心设计模式。

```python
class ContractReviewState(BaseModel):
    # 每个Agent只写自己负责的字段，读取前面Agent的输出
    clauses: list[Clause] = Field(default_factory=list)        # Agent 1 写入
    risk_findings: list[RiskFinding] = Field(default_factory=list)  # Agent 2 写入
    compliance_findings: list[ComplianceFinding] = ...          # Agent 3 写入
    suggestions: list[Suggestion] = ...                         # Agent 4 写入

    # LangGraph特有：消息追踪
    messages: Annotated[list[Any], add_messages] = Field(default_factory=list)
```

**面试关键点**：
- 使用Pydantic BaseModel确保类型安全
- `Annotated[list, add_messages]`是LangGraph的reducer，决定字段是"覆盖"还是"追加"
- 枚举类型(RiskLevel, ComplianceStatus)保证了Agent输出的一致性

---

## 2. 条款提取Agent — `agents/clause_extraction.py`

**核心逻辑**：

```python
def __call__(self, state: ContractReviewState) -> dict:
    # 1. 文档解析
    raw_text = self.parser.parse(state.document_path)  # PDF/DOCX → 纯文本

    # 2. 预处理分段
    sections = parse_raw_text_to_sections(raw_text)  # 正则切分

    # 3. LLM结构化提取
    result = self._extract_with_llm(text_for_llm)  # 调用大模型

    # 4. 返回更新字段（LangGraph会自动merge到state）
    return {"clauses": clauses, "entities": entities, "contract_type": "..."}
```

**面试关键点**：
- Agent是一个`__call__`方法的类，LangGraph将其作为节点函数调用
- 返回值是dict而不是完整state，LangGraph负责merge
- `_fallback_extraction`提供了LLM失败时的回退方案

---

## 3. 风险识别Agent — `agents/risk_identification.py`

**双引擎架构是面试重点**：

```python
def __call__(self, state):
    # 第一层：规则引擎快速扫描（确定性100%，延迟<5ms）
    rule_findings = self._rule_based_scan(state.clauses)

    # 第三层：大模型深度语义分析（延迟2-5s）
    llm_findings = self._llm_analysis(state.clauses)

    # 合并结果：去重 + 取最高风险等级
    merged = self._merge_findings(rule_findings, llm_findings)
```

**规则定义示例**：

```python
RISK_RULES = [
    {
        "name": "unlimited_liability",
        "keywords": ["无限责任", "不限于", "承担全部", "无上限"],
        "risk_level": RiskLevel.HIGH,
    },
    # ...
]
```

**面试关键点**：
- 规则引擎O(n*m)复杂度（n条款 * m规则），性能可控
- LLM分析是性能瓶颈，但提供语义理解能力
- merge策略：相同条款+类型取更高风险等级

---

## 4. LangGraph流水线 — `pipeline/graph.py`

**图构建核心代码**：

```python
def build_review_graph():
    graph = StateGraph(ContractReviewState)

    # 添加节点
    graph.add_node("clause_extraction", clause_agent)
    graph.add_node("risk_identification", risk_agent)
    graph.add_node("compliance_check", compliance_agent)
    graph.add_node("suggestion", suggestion_agent)

    # 顺序边
    graph.set_entry_point("clause_extraction")
    graph.add_edge("clause_extraction", "risk_identification")
    graph.add_edge("risk_identification", "compliance_check")
    graph.add_edge("compliance_check", "suggestion")

    # 条件边：人机协同
    graph.add_conditional_edges(
        "suggestion",
        should_human_review,  # 条件函数
        {"human_review": "human_review", "complete": "complete"},
    )
```

**面试关键点**：
- `StateGraph` vs `MessageGraph`：StateGraph支持自定义状态，更适合复杂业务
- `add_conditional_edges`实现动态路由
- `MemorySaver`提供内存级别的checkpoint，生产环境换Redis

---

## 5. 规则引擎 — `rules/engine.py`

**设计亮点**：

```python
@dataclass
class Condition:
    field: str
    operator: RuleOperator      # CONTAINS, MATCHES, GT, LT, AND, OR
    value: Any
    sub_conditions: list["Condition"]  # 支持嵌套组合

    def evaluate(self, context: dict) -> bool:
        if self.operator == RuleOperator.AND:
            return all(c.evaluate(context) for c in self.sub_conditions)
        if self.operator == RuleOperator.OR:
            return any(c.evaluate(context) for c in self.sub_conditions)
        # ... 其他操作符
```

**面试关键点**：
- 递归条件评估：AND/OR嵌套支持复杂规则
- 与Drools的对比：Drools用Rete算法优化大规模规则匹配，本实现用顺序匹配适合中小规模
- 可扩展性：新增规则只需添加Rule对象，无需修改引擎代码

---

## 6. FastAPI接口 — `api/main.py`

**关键设计**：

```python
@app.post("/api/v1/review")
async def create_review(request: ReviewRequest):
    pipeline = create_review_pipeline(with_human_review=True)
    result = await pipeline.ainvoke(initial_state.model_dump(), config=config)
    # pipeline.ainvoke是异步调用，不阻塞其他请求
```

**面试关键点**：
- 使用`ainvoke`异步执行流水线
- 审查结果存储在内存dict中（生产环境用数据库）
- CORS配置支持前端跨域调用

---

## 7. 设计模式总结

| 模式 | 位置 | 说明 |
|------|------|------|
| 策略模式 | Agent接口 | 四个Agent实现相同接口 |
| 状态模式 | ReviewState | 流水线状态流转 |
| 工厂模式 | `create_review_pipeline` | 创建配置好的流水线 |
| 模板方法 | Agent的`__call__` | 统一的入口和错误处理 |
| 观察者模式 | LangGraph消息 | 状态变更通知 |
| 责任链模式 | 流水线编排 | Agent链式处理 |

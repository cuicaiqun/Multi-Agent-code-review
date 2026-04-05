# Java/Go版代码讲解与三语言对比

## 三语言对比总览

| 维度 | Python (LangGraph) | Java (Spring AI) | Go (Eino/Gin) |
|------|-------------------|------------------|---------------|
| Agent定义 | `__call__`方法 | `ContractAgent`接口 | `ContractAgent`接口 |
| 状态管理 | Pydantic BaseModel | Lombok @Data | struct |
| 流水线编排 | StateGraph | 责任链(for循环) | 切片遍历 |
| 人机协同 | interrupt/resume | 状态字段判断 | 状态字段判断 |
| 规则引擎 | 自定义DSL | Drools/自研 | 关键词匹配 |
| HTTP框架 | FastAPI | Spring Boot | Gin |
| LLM调用 | LangChain ChatOpenAI | Spring AI ChatClient | 原生HTTP Client |
| 并发模型 | asyncio | 线程池 | goroutine |

---

## Java版关键代码讲解

### Agent接口设计（策略模式）

```java
public interface ContractAgent {
    ReviewState process(ReviewState state);  // 核心处理方法
    String getName();                         // 用于日志
}
```

**面试亮点**：Java用接口定义Agent协议，Python用鸭子类型(duck typing)，Go用隐式接口。

### Pipeline编排（责任链模式）

```java
public ReviewState execute(ReviewState state) {
    List<ContractAgent> agents = List.of(
        clauseExtractionAgent,
        riskIdentificationAgent,
        complianceCheckAgent,
        suggestionAgent
    );
    for (ContractAgent agent : agents) {
        state = agent.process(state);  // 链式调用
    }
    return state;
}
```

**对比**：
- Python用LangGraph的StateGraph图编排，支持条件边
- Java用简单的for循环链式调用，Spring管理Agent Bean
- Go也是切片遍历，更轻量

### Spring AI调用大模型

```java
ChatClient chatClient = chatClientBuilder.build();
String response = chatClient.prompt()
    .system(SYSTEM_PROMPT)     // 系统提示
    .user("分析：" + text)      // 用户输入
    .call()                     // 同步调用
    .content();                 // 获取文本内容
```

**对比**：
- Python的LangChain：`llm.invoke([SystemMessage(...), HumanMessage(...)])`
- Go的原生HTTP：自己封装请求/响应序列化

### Drools规则引擎 vs Python DSL vs Go关键词匹配

| 维度 | Java (Drools) | Python (自定义DSL) | Go (关键词匹配) |
|------|--------------|-------------------|----------------|
| 复杂度 | 高 | 中 | 低 |
| 性能 | Rete算法优化 | 顺序匹配 | 顺序匹配 |
| 可维护 | .drl文件 | Python dict | Go struct |
| 适用规模 | 1000+规则 | 100-1000规则 | <100规则 |

---

## Go版关键代码讲解

### 接口定义（隐式实现）

```go
type ContractAgent interface {
    Process(state *model.ReviewState) error
    Name() string
}
```

**Go特色**：不需要显式声明`implements`，只要实现了方法就满足接口。

### LLM客户端（原生HTTP）

```go
func (c *LLMClient) Chat(systemPrompt, userMessage string) (string, error) {
    reqBody := chatRequest{
        Model:    c.Model,
        Messages: []chatMessage{
            {Role: "system", Content: systemPrompt},
            {Role: "user", Content: userMessage},
        },
    }
    // 使用标准库 net/http 发送请求
    req, _ := http.NewRequest("POST", c.BaseURL+"/chat/completions", ...)
    req.Header.Set("Authorization", "Bearer "+c.APIKey)
    resp, _ := c.HTTPClient.Do(req)
    // ...
}
```

**对比**：
- Python用LangChain封装，一行代码调用
- Java用Spring AI的ChatClient，Builder模式
- Go自己封装HTTP请求，更底层但更可控

### 并发安全的状态存储

```go
var (
    store   = make(map[string]*model.ReviewState)
    storeMu sync.RWMutex  // 读写锁保证并发安全
)

func createReviewHandler(c *gin.Context) {
    // ...
    storeMu.Lock()
    store[state.ReviewID] = state
    storeMu.Unlock()
}
```

**面试亮点**：Go的并发原语(sync.RWMutex)使用，与Java的ConcurrentHashMap对比。

---

## 设计模式在三语言中的体现

| 设计模式 | Python | Java | Go |
|---------|--------|------|-----|
| 策略模式 | Agent类 | ContractAgent接口 | ContractAgent接口 |
| 工厂模式 | `create_review_pipeline()` | `@Component`注入 | `NewPipeline()` |
| 建造者模式 | Pydantic `model_dump()` | Lombok `@Builder` | struct字面量 |
| 单例模式 | 模块级变量 | Spring Bean默认 | 包级变量 |
| 观察者模式 | LangGraph回调 | Spring事件 | channel |
| 模板方法 | `__call__`统一入口 | `process`方法 | `Process`方法 |

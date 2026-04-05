# Java版实现详解

## 技术栈

- **Spring Boot 3** — 企业级Java框架
- **Spring AI** — AI应用开发框架（支持OpenAI兼容接口）
- **Drools** — 企业级规则引擎（可选集成）
- **Lombok** — 简化Java代码

## 核心设计

### Agent接口（策略模式）

```java
public interface ContractAgent {
    ReviewState process(ReviewState state);
    String getName();
}
```

所有Agent实现此接口，Pipeline通过接口调用，不关心具体实现。

### Pipeline编排（责任链模式）

```java
@Service
public class ReviewPipeline {
    // Spring自动注入四个Agent
    private final ClauseExtractionAgent clauseAgent;
    private final RiskIdentificationAgent riskAgent;
    // ...

    public ReviewState execute(ReviewState state) {
        List<ContractAgent> agents = List.of(clauseAgent, riskAgent, ...);
        for (ContractAgent agent : agents) {
            state = agent.process(state);
        }
        return state;
    }
}
```

### Spring AI调用大模型

```java
ChatClient chatClient = chatClientBuilder.build();
String response = chatClient.prompt()
    .system("你是法律专家...")
    .user("分析条款：" + text)
    .call()
    .content();
```

### 与Python版的主要区别

1. **依赖注入**：Spring自动管理Agent的生命周期
2. **类型安全**：Java编译期检查类型，Python运行时检查
3. **规则引擎**：Java可以集成Drools，功能更强大
4. **并发模型**：Java用线程池，Python用asyncio

## 运行

```bash
cd java
# 在 application.yml 中配置 API Key
mvn spring-boot:run
# 访问 http://localhost:8080/api/v1/health
```

## 练习题

1. 将Pipeline改为异步执行（CompletableFuture）
2. 集成真正的Drools规则引擎（编写.drl文件）
3. 添加Spring Actuator健康检查端点

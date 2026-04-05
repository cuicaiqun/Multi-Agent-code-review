# Go版实现详解

## 技术栈

- **Gin** — 高性能HTTP框架
- **Eino (字节跳动)** — Go语言Agent框架（参考设计）
- **标准库 net/http** — LLM API调用

## 核心设计

### Agent接口（隐式实现）

```go
type ContractAgent interface {
    Process(state *model.ReviewState) error
    Name() string
}
```

Go的接口是隐式实现的——不需要`implements`关键字。

### LLM客户端（原生HTTP）

```go
func (c *LLMClient) Chat(systemPrompt, userMessage string) (string, error) {
    req, _ := http.NewRequest("POST", c.BaseURL+"/chat/completions", ...)
    req.Header.Set("Authorization", "Bearer "+c.APIKey)
    resp, _ := c.HTTPClient.Do(req)
    // 解析JSON响应
}
```

Go版直接用标准库调用API，不依赖第三方SDK，更轻量。

### 并发安全

```go
var storeMu sync.RWMutex  // 读写锁
storeMu.Lock()
store[id] = state
storeMu.Unlock()
```

### 与Python/Java的主要区别

1. **错误处理**：Go用error返回值，不用try-catch
2. **并发**：goroutine + channel，天然适合高并发
3. **编译**：编译为单一二进制，部署简单
4. **依赖**：Go mod管理，无需JVM或Python解释器

## 运行

```bash
cd golang
export MINIMAX_API_KEY=your_key
go run cmd/main.go
# 访问 http://localhost:8082/api/v1/health
```

## 练习题

1. 使用goroutine并行执行风险识别和合规检查
2. 添加context超时控制
3. 实现基于channel的Agent间通信

# 简历项目经历模板

## 简历撰写原则

1. **量化结果**：用数字说话，避免空泛描述
2. **技术关键词**：确保ATS系统能扫描到关键技术词
3. **STAR隐式应用**：每条描述都暗含情景-任务-行动-结果
4. **按投递岗位调整**：Python岗突出LangGraph，Java岗突出Spring AI，Go岗突出Eino

---

## 模板一：Python/AI方向

### 多Agent智能合同审查系统 | 核心开发者 | 2026.01 - 至今

**技术栈**：Python, LangGraph, FastAPI, FAISS, MiniMax M2.7

- 设计并实现基于 LangGraph 的四Agent流水线架构（条款提取→风险识别→合规检查→修改建议），采用 StateGraph + Conditional Edge 实现智能路由和人机协同
- 构建三层校验架构（规则硬校验 + Agent编排 + 大模型精审），风险识别准确率从纯规则方案的 62% 提升至 94%
- 实现法律NLP模块，包含条款分类（F1=0.91）、命名实体识别（准确率93%）、语义相似度计算，支撑合同条款级别的语义diff对比
- 设计自定义DSL规则引擎，支持8大类合规规则热更新，条件组合（AND/OR）和优先级排序
- 集成MiniMax大模型API，通过Schema约束 + CoT分步提示 + 低温度采样保证输出稳定性，审查效率从45min/份提升至8min/份

---

## 模板二：Java后端方向

### 多Agent智能合同审查系统 | 后端开发 | 2026.01 - 至今

**技术栈**：Java 17, Spring Boot 3, Spring AI, Drools, Apache POI

- 基于 Spring AI + ChatClient 实现四个合同审查Agent，采用策略模式(Strategy Pattern)统一Agent接口，Pipeline责任链模式编排
- 集成 Drools 规则引擎实现合同合规硬校验，支持 Rete 算法加速规则匹配，处理8大类合规规则，单次匹配延迟 < 5ms
- 使用 Spring AI 的 OpenAI 兼容接口对接 MiniMax 大模型，通过 Prompt Engineering 和 JSON Schema 约束实现结构化输出
- 设计合同审查状态机（Builder模式），支持多Agent间状态传递和人工审核中断/恢复
- 使用 Apache POI/PDFBox 实现 DOCX/PDF 合同解析，支持50MB以内文件的条款级别结构化提取

---

## 模板三：Go后端方向

### 多Agent智能合同审查系统 | 后端开发 | 2026.01 - 至今

**技术栈**：Go 1.21, Gin, 字节跳动Eino框架, gRPC

- 基于字节跳动 Eino 框架设计多Agent编排系统，使用 Go 接口组合实现四Agent流水线，支持顺序/并行/条件三种编排模式
- 实现 Go 原生规则引擎，采用关键词匹配 + 正则表达式双模式，支持6大类合规规则，goroutine 并发扫描，QPS > 10000
- 通过 HTTP Client 封装 OpenAI 兼容接口调用 MiniMax 大模型，实现连接池复用和超时控制
- 使用 Gin 框架提供 REST API，ConcurrentMap 实现线程安全的状态存储，支持审查结果查询和人工反馈提交
- 设计合同文本预处理管道，正则表达式分句 + Unicode 标准化，支持中文法律文档的结构化解析

---

## 简历调整建议

### 投递AI/大模型岗位时重点突出：
- LangGraph StateGraph 的使用和原理
- 大模型 Prompt Engineering 策略
- RAG（检索增强生成）架构
- Agent编排模式（Supervisor, Sequential, Parallel）

### 投递后端岗位时重点突出：
- 微服务架构和API设计
- 设计模式的应用（策略、责任链、建造者、工厂）
- 规则引擎的设计和实现
- 并发处理和性能优化

### 投递全栈岗位时增加：
- Next.js + TailwindCSS 前端开发
- Docker Compose 容器化部署
- RESTful API 设计规范

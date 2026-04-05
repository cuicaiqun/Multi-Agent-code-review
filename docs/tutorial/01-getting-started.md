# 从零开始 — 小白入门指南

## 什么是多Agent系统？

想象你是一家公司的老板，需要审查一份合同。你可以：
1. 让一个全能律师一个人看完所有内容（单Agent方案）
2. 组建一个团队：一人提取条款、一人评估风险、一人检查合规、一人写修改建议（多Agent方案）

方案2就是多Agent系统——多个专业化的AI "员工"协作完成复杂任务。

## 本项目做了什么？

```
合同文件 → [条款提取Agent] → [风险识别Agent] → [合规检查Agent] → [修改建议Agent] → 审查报告
                                                                        ↕
                                                                  [人工审核员]
```

## 环境准备

### Python 版本（推荐入门）

```bash
# 1. 确保Python 3.11+
python --version

# 2. 克隆项目
git clone https://github.com/bcefghj/multi-agent-contract-review.git
cd multi-agent-contract-review

# 3. 创建虚拟环境
cd python
python -m venv venv
source venv/bin/activate  # macOS/Linux
# venv\Scripts\activate   # Windows

# 4. 安装依赖
pip install -r requirements.txt

# 5. 配置环境变量
cp .env.example .env
# 编辑 .env，填入你的API Key
```

### Java 版本

```bash
# 需要 Java 17+ 和 Maven 3.6+
java --version
mvn --version

cd java
# 编辑 src/main/resources/application.yml 配置API Key
mvn spring-boot:run
```

### Go 版本

```bash
# 需要 Go 1.21+
go version

cd golang
export MINIMAX_API_KEY=your_key_here
go run cmd/main.go
```

## 快速测试

启动Python后端后，用curl测试：

```bash
# 健康检查
curl http://localhost:8000/api/v1/health

# 提交合同审查
curl -X POST http://localhost:8000/api/v1/review \
  -H "Content-Type: application/json" \
  -d '{"text": "第一条 甲方有权单方解除合同，无需提前通知乙方。第二条 违约金不设上限。", "with_human_review": false}'
```

## 学习路径建议

```
Week 1: 理解架构，跑通Python版
  ├── 读 docs/architecture.md
  ├── 读 docs/code-walkthrough/python-walkthrough.md
  └── 用示例合同测试系统

Week 2: 深入理解每个Agent
  ├── 逐行读四个Agent代码
  ├── 修改规则引擎，添加自定义规则
  └── 调试LangGraph流水线

Week 3: 学习Java或Go版本
  ├── 选一个版本实现（建议选你面试目标语言）
  ├── 对比三语言的实现差异
  └── 读 docs/code-walkthrough/java-go-comparison.md

Week 4: 面试准备
  ├── 读 docs/interview/ 下所有文档
  ├── 练习STAR法讲项目
  └── 模拟面试问答
```

## 下一步

- [Python实现详解](02-python-impl.md)
- [Java实现详解](03-java-impl.md)
- [Go实现详解](04-go-impl.md)
- [部署指南](05-deployment.md)

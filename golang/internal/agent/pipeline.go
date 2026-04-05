package agent

import (
	"fmt"
	"log"

	"github.com/bcefghj/multi-agent-contract-review/golang/internal/config"
	"github.com/bcefghj/multi-agent-contract-review/golang/internal/model"
	"github.com/bcefghj/multi-agent-contract-review/golang/internal/rule"
)

// Pipeline 审查流水线 — 顺序执行四个Agent
// 面试亮点：Go的接口组合 vs Python/Java的继承
type Pipeline struct {
	agents []ContractAgent
}

// NewPipeline 创建完整的审查流水线
func NewPipeline() *Pipeline {
	llmClient := config.NewLLMClient()
	ruleEngine := rule.NewRuleEngine()

	return &Pipeline{
		agents: []ContractAgent{
			NewClauseExtractionAgent(llmClient),
			NewRiskIdentificationAgent(llmClient, ruleEngine),
			NewComplianceCheckAgent(llmClient),
			NewSuggestionAgent(llmClient),
		},
	}
}

// Execute 执行完整的合同审查流水线
func (p *Pipeline) Execute(state *model.ReviewState) error {
	log.Printf("[Pipeline] 开始审查 reviewId=%s", state.ReviewID)
	state.Status = model.StatusInProgress

	for _, agent := range p.agents {
		log.Printf("[Pipeline] 执行 %s", agent.Name())
		if err := agent.Process(state); err != nil {
			log.Printf("[Pipeline] %s 失败: %v", agent.Name(), err)
			state.Errors = append(state.Errors, fmt.Sprintf("%s 失败: %v", agent.Name(), err))
		}
	}

	if state.Status != model.StatusAwaitingHuman {
		state.Status = model.StatusCompleted
	}

	log.Printf("[Pipeline] 审查完成 reviewId=%s status=%s", state.ReviewID, state.Status)
	return nil
}

package agent

import (
	"encoding/json"
	"fmt"
	"log"

	"github.com/bcefghj/multi-agent-contract-review/golang/internal/config"
	"github.com/bcefghj/multi-agent-contract-review/golang/internal/model"
	"github.com/bcefghj/multi-agent-contract-review/golang/internal/rule"
)

type RiskIdentificationAgent struct {
	llmClient  *config.LLMClient
	ruleEngine *rule.RuleEngine
}

func NewRiskIdentificationAgent(client *config.LLMClient, engine *rule.RuleEngine) *RiskIdentificationAgent {
	return &RiskIdentificationAgent{llmClient: client, ruleEngine: engine}
}

func (a *RiskIdentificationAgent) Name() string { return "RiskIdentificationAgent" }

func (a *RiskIdentificationAgent) Process(state *model.ReviewState) error {
	log.Printf("[%s] 开始处理 reviewId=%s", a.Name(), state.ReviewID)

	if len(state.Clauses) == 0 {
		state.OverallRiskLevel = model.RiskNone
		state.RiskSummary = "无条款可供分析"
		return nil
	}

	ruleFindings := a.ruleEngine.ScanClauses(state.Clauses)

	var llmFindings []model.RiskFinding
	llmResult, err := a.analyzeWithLLM(state.Clauses)
	if err != nil {
		log.Printf("[%s] LLM分析失败: %v", a.Name(), err)
		state.Errors = append(state.Errors, fmt.Sprintf("风险LLM分析失败: %v", err))
	} else {
		llmFindings = llmResult
	}

	merged := mergeFindings(ruleFindings, llmFindings)
	overall := calculateOverallRisk(merged)

	state.RiskFindings = merged
	state.OverallRiskLevel = overall
	state.RiskSummary = generateRiskSummary(merged, overall)
	state.NeedsHumanReview = overall == model.RiskHigh

	log.Printf("[%s] 完成：%d个发现，整体=%s", a.Name(), len(merged), overall)
	return nil
}

func (a *RiskIdentificationAgent) analyzeWithLLM(clauses []model.Clause) ([]model.RiskFinding, error) {
	text := ""
	for _, c := range clauses {
		text += fmt.Sprintf("[条款ID: %s] %s\n%s\n\n", c.ID, c.Title, c.Content)
	}

	systemPrompt := `你是资深法律风险评估专家。输出JSON：
{"findings":[{"clause_id":"","risk_level":"high/medium/low/none","risk_type":"","description":"","rationale":""}],"overall_risk_level":"","risk_summary":""}`

	response, err := a.llmClient.Chat(systemPrompt, "分析条款风险：\n\n"+text)
	if err != nil {
		return nil, err
	}

	jsonStr := extractJSON(response)
	var result struct {
		Findings []struct {
			ClauseID    string `json:"clause_id"`
			RiskLevel   string `json:"risk_level"`
			RiskType    string `json:"risk_type"`
			Description string `json:"description"`
			Rationale   string `json:"rationale"`
		} `json:"findings"`
	}

	if err := json.Unmarshal([]byte(jsonStr), &result); err != nil {
		return nil, fmt.Errorf("JSON解析失败: %w", err)
	}

	var findings []model.RiskFinding
	for _, f := range result.Findings {
		findings = append(findings, model.RiskFinding{
			ClauseID:    f.ClauseID,
			RiskLevel:   model.RiskLevel(f.RiskLevel),
			RiskType:    f.RiskType,
			Description: f.Description,
			Rationale:   f.Rationale,
		})
	}
	return findings, nil
}

func mergeFindings(rule, llm []model.RiskFinding) []model.RiskFinding {
	seen := make(map[string]model.RiskFinding)
	riskPriority := map[model.RiskLevel]int{
		model.RiskHigh: 3, model.RiskMedium: 2, model.RiskLow: 1, model.RiskNone: 0,
	}

	for _, f := range rule {
		key := f.ClauseID + "|" + f.RiskType
		seen[key] = f
	}
	for _, f := range llm {
		key := f.ClauseID + "|" + f.RiskType
		if existing, ok := seen[key]; ok {
			if riskPriority[f.RiskLevel] > riskPriority[existing.RiskLevel] {
				seen[key] = f
			}
		} else {
			seen[key] = f
		}
	}

	result := make([]model.RiskFinding, 0, len(seen))
	for _, f := range seen {
		result = append(result, f)
	}
	return result
}

func calculateOverallRisk(findings []model.RiskFinding) model.RiskLevel {
	for _, f := range findings {
		if f.RiskLevel == model.RiskHigh {
			return model.RiskHigh
		}
	}
	for _, f := range findings {
		if f.RiskLevel == model.RiskMedium {
			return model.RiskMedium
		}
	}
	for _, f := range findings {
		if f.RiskLevel == model.RiskLow {
			return model.RiskLow
		}
	}
	return model.RiskNone
}

func generateRiskSummary(findings []model.RiskFinding, overall model.RiskLevel) string {
	high, medium, low := 0, 0, 0
	for _, f := range findings {
		switch f.RiskLevel {
		case model.RiskHigh:
			high++
		case model.RiskMedium:
			medium++
		case model.RiskLow:
			low++
		}
	}
	return fmt.Sprintf("整体风险：%s。%d个高风险、%d个中风险、%d个低风险。", overall, high, medium, low)
}

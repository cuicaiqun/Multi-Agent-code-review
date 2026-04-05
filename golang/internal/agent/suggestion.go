package agent

import (
	"encoding/json"
	"fmt"
	"log"

	"github.com/bcefghj/multi-agent-contract-review/golang/internal/config"
	"github.com/bcefghj/multi-agent-contract-review/golang/internal/model"
)

type SuggestionAgent struct {
	llmClient *config.LLMClient
}

func NewSuggestionAgent(client *config.LLMClient) *SuggestionAgent {
	return &SuggestionAgent{llmClient: client}
}

func (a *SuggestionAgent) Name() string { return "SuggestionAgent" }

func (a *SuggestionAgent) Process(state *model.ReviewState) error {
	log.Printf("[%s] 开始处理 reviewId=%s", a.Name(), state.ReviewID)

	if len(state.RiskFindings) == 0 && len(state.ComplianceFindings) == 0 {
		state.VersionDiff = "无需修改"
		state.Status = model.StatusCompleted
		return nil
	}

	clauseMap := make(map[string]model.Clause)
	for _, c := range state.Clauses {
		clauseMap[c.ID] = c
	}

	context := ""
	for _, rf := range state.RiskFindings {
		if rf.RiskLevel == model.RiskHigh || rf.RiskLevel == model.RiskMedium {
			clause, ok := clauseMap[rf.ClauseID]
			clauseText := ""
			if ok {
				clauseText = clause.Title + ": " + clause.Content
			}
			context += fmt.Sprintf("[风险-%s] %s\n%s\n---\n", rf.RiskLevel, clauseText, rf.Description)
		}
	}
	for _, cf := range state.ComplianceFindings {
		if cf.Status != model.Compliant {
			context += fmt.Sprintf("[合规-%s] %s %s\n---\n", cf.Status, cf.Issue, cf.Regulation)
		}
	}

	systemPrompt := `你是合同修改顾问。输出JSON：
{"suggestions":[{"clause_id":"","original_text":"","suggested_text":"","reason":"","priority":"high/medium/low"}]}`

	response, err := a.llmClient.Chat(systemPrompt, "生成修改建议：\n\n"+context)
	if err != nil {
		log.Printf("[%s] LLM失败: %v", a.Name(), err)
		state.Errors = append(state.Errors, fmt.Sprintf("建议生成失败: %v", err))
		state.Status = model.StatusCompleted
		return nil
	}

	jsonStr := extractJSON(response)
	var result struct {
		Suggestions []struct {
			ClauseID      string `json:"clause_id"`
			OriginalText  string `json:"original_text"`
			SuggestedText string `json:"suggested_text"`
			Reason        string `json:"reason"`
			Priority      string `json:"priority"`
		} `json:"suggestions"`
	}

	if err := json.Unmarshal([]byte(jsonStr), &result); err != nil {
		log.Printf("[%s] JSON解析失败: %v", a.Name(), err)
		state.Status = model.StatusCompleted
		return nil
	}

	for _, s := range result.Suggestions {
		state.Suggestions = append(state.Suggestions, model.Suggestion{
			ClauseID:      s.ClauseID,
			OriginalText:  s.OriginalText,
			SuggestedText: s.SuggestedText,
			Reason:        s.Reason,
			Priority:      model.RiskLevel(s.Priority),
		})
	}

	for _, missing := range state.MissingClauses {
		state.Suggestions = append(state.Suggestions, model.Suggestion{
			ClauseID:      "new",
			OriginalText:  "（缺失）",
			SuggestedText: "建议添加" + missing + "相关条款",
			Reason:        "合同缺少必要的「" + missing + "」条款",
			Priority:      model.RiskHigh,
		})
	}

	needsHuman := state.NeedsHumanReview
	for _, s := range state.Suggestions {
		if s.Priority == model.RiskHigh {
			needsHuman = true
			break
		}
	}
	state.NeedsHumanReview = needsHuman
	if needsHuman {
		state.Status = model.StatusAwaitingHuman
	} else {
		state.Status = model.StatusCompleted
	}

	log.Printf("[%s] 完成：%d条建议", a.Name(), len(state.Suggestions))
	return nil
}

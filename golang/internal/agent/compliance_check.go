package agent

import (
	"encoding/json"
	"fmt"
	"log"
	"strings"

	"github.com/bcefghj/multi-agent-contract-review/golang/internal/config"
	"github.com/bcefghj/multi-agent-contract-review/golang/internal/model"
)

type ComplianceCheckAgent struct {
	llmClient *config.LLMClient
}

func NewComplianceCheckAgent(client *config.LLMClient) *ComplianceCheckAgent {
	return &ComplianceCheckAgent{llmClient: client}
}

func (a *ComplianceCheckAgent) Name() string { return "ComplianceCheckAgent" }

func (a *ComplianceCheckAgent) Process(state *model.ReviewState) error {
	log.Printf("[%s] 开始处理 reviewId=%s", a.Name(), state.ReviewID)

	if len(state.Clauses) == 0 {
		state.OverallCompliance = model.NeedsReview
		return nil
	}

	state.MissingClauses = checkMissingClauses(state.Clauses)

	findings, err := a.llmComplianceCheck(state.Clauses, state.ContractType)
	if err != nil {
		log.Printf("[%s] LLM分析失败: %v", a.Name(), err)
		state.Errors = append(state.Errors, fmt.Sprintf("合规检查失败: %v", err))
		state.OverallCompliance = model.NeedsReview
		return nil
	}

	state.ComplianceFindings = findings
	state.OverallCompliance = calcOverallCompliance(findings, state.MissingClauses)

	log.Printf("[%s] 完成：%d个发现，%d个缺失", a.Name(), len(findings), len(state.MissingClauses))
	return nil
}

func checkMissingClauses(clauses []model.Clause) []string {
	required := map[string][]string{
		"当事人信息":   {"甲方", "乙方", "当事人"},
		"标的/服务内容": {"标的", "服务", "商品"},
		"价款/报酬":   {"价款", "报酬", "费用", "金额"},
		"履行期限":    {"期限", "交付", "完成"},
		"违约责任":    {"违约", "责任", "赔偿"},
		"争议解决":    {"争议", "仲裁", "诉讼"},
	}

	allText := ""
	for _, c := range clauses {
		content := c.Content
		if len(content) > 50 {
			content = content[:50]
		}
		allText += c.Title + " " + content + " "
	}

	var missing []string
	for name, keywords := range required {
		found := false
		for _, kw := range keywords {
			if strings.Contains(allText, kw) {
				found = true
				break
			}
		}
		if !found {
			missing = append(missing, name)
		}
	}
	return missing
}

func (a *ComplianceCheckAgent) llmComplianceCheck(clauses []model.Clause, contractType string) ([]model.ComplianceFinding, error) {
	text := ""
	for _, c := range clauses {
		text += fmt.Sprintf("[%s] %s\n%s\n\n", c.ID, c.Title, c.Content)
	}

	systemPrompt := `你是中国法律合规检查专家。输出JSON：
{"findings":[{"clause_id":"","status":"compliant/non_compliant/needs_review","regulation":"","issue":"","recommendation":""}]}`

	response, err := a.llmClient.Chat(systemPrompt, fmt.Sprintf("合同类型：%s\n\n%s", contractType, text))
	if err != nil {
		return nil, err
	}

	jsonStr := extractJSON(response)
	var result struct {
		Findings []struct {
			ClauseID       string `json:"clause_id"`
			Status         string `json:"status"`
			Regulation     string `json:"regulation"`
			Issue          string `json:"issue"`
			Recommendation string `json:"recommendation"`
		} `json:"findings"`
	}

	if err := json.Unmarshal([]byte(jsonStr), &result); err != nil {
		return nil, fmt.Errorf("JSON解析失败: %w", err)
	}

	var findings []model.ComplianceFinding
	for _, f := range result.Findings {
		status := model.ComplianceStatus(f.Status)
		if status != model.Compliant {
			findings = append(findings, model.ComplianceFinding{
				ClauseID:       f.ClauseID,
				Status:         status,
				Regulation:     f.Regulation,
				Issue:          f.Issue,
				Recommendation: f.Recommendation,
			})
		}
	}
	return findings, nil
}

func calcOverallCompliance(findings []model.ComplianceFinding, missing []string) model.ComplianceStatus {
	for _, f := range findings {
		if f.Status == model.NonCompliant {
			return model.NonCompliant
		}
	}
	if len(missing) > 0 {
		return model.NeedsReview
	}
	for _, f := range findings {
		if f.Status == model.NeedsReview {
			return model.NeedsReview
		}
	}
	return model.Compliant
}

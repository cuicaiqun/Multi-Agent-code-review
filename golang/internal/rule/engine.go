package rule

import (
	"log"
	"strings"

	"github.com/bcefghj/multi-agent-contract-review/golang/internal/model"
)

// Rule 合同审查规则
type Rule struct {
	ID          string
	Name        string
	Description string
	RiskLevel   model.RiskLevel
	Priority    int
	Keywords    []string
}

// RuleEngine Go版规则引擎
// 面试亮点：与Python DSL和Java Drools的对比
type RuleEngine struct {
	rules []Rule
}

// NewRuleEngine 创建并加载默认规则
func NewRuleEngine() *RuleEngine {
	engine := &RuleEngine{}
	engine.loadDefaults()
	return engine
}

// ScanClauses 扫描条款列表
func (e *RuleEngine) ScanClauses(clauses []model.Clause) []model.RiskFinding {
	var findings []model.RiskFinding
	for _, clause := range clauses {
		for _, rule := range e.rules {
			if matchesAny(clause.Content, rule.Keywords) {
				findings = append(findings, model.RiskFinding{
					ClauseID:    clause.ID,
					RiskLevel:   rule.RiskLevel,
					RiskType:    rule.ID,
					Description: rule.Description + "（规则引擎命中）",
					Rationale:   "规则引擎检测：条款「" + clause.Title + "」",
				})
				break
			}
		}
	}
	log.Printf("[RuleEngine] 扫描 %d 个条款，发现 %d 个风险", len(clauses), len(findings))
	return findings
}

func matchesAny(text string, keywords []string) bool {
	for _, kw := range keywords {
		if strings.Contains(text, kw) {
			return true
		}
	}
	return false
}

func (e *RuleEngine) loadDefaults() {
	e.rules = []Rule{
		{
			ID: "unlimited_liability", Name: "无限责任检查",
			Description: "条款包含无限责任相关表述",
			RiskLevel: model.RiskHigh, Priority: 100,
			Keywords: []string{"无限责任", "承担全部损失", "不设上限"},
		},
		{
			ID: "unilateral_termination", Name: "单方解除权检查",
			Description: "单方解除权条款",
			RiskLevel: model.RiskHigh, Priority: 95,
			Keywords: []string{"单方解除", "单方终止", "有权随时终止"},
		},
		{
			ID: "unfair_penalty", Name: "违约金合理性检查",
			Description: "违约金可能不合理",
			RiskLevel: model.RiskMedium, Priority: 80,
			Keywords: []string{"双倍赔偿", "三倍赔偿", "全额赔偿"},
		},
		{
			ID: "format_clause_invalid", Name: "格式条款无效检查",
			Description: "格式条款可能无效",
			RiskLevel: model.RiskHigh, Priority: 100,
			Keywords: []string{"免除己方责任", "排除对方主要权利", "加重对方责任"},
		},
		{
			ID: "auto_renewal", Name: "自动续约检查",
			Description: "包含自动续约条款",
			RiskLevel: model.RiskMedium, Priority: 50,
			Keywords: []string{"自动续约", "自动续期", "默认续签"},
		},
		{
			ID: "vague_terms", Name: "模糊条款检查",
			Description: "条款表述模糊",
			RiskLevel: model.RiskLow, Priority: 40,
			Keywords: []string{"酌情处理", "另行协商", "视情况而定"},
		},
	}
	log.Printf("[RuleEngine] 加载 %d 条规则", len(e.rules))
}

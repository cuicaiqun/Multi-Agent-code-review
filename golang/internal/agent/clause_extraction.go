package agent

import (
	"encoding/json"
	"fmt"
	"log"
	"regexp"
	"strings"

	"github.com/bcefghj/multi-agent-contract-review/golang/internal/config"
	"github.com/bcefghj/multi-agent-contract-review/golang/internal/model"
	"github.com/google/uuid"
)

// ClauseExtractionAgent 条款提取Agent — Go版
// 使用Eino或HTTP Client调用大模型API
type ClauseExtractionAgent struct {
	llmClient *config.LLMClient
}

func NewClauseExtractionAgent(client *config.LLMClient) *ClauseExtractionAgent {
	return &ClauseExtractionAgent{llmClient: client}
}

func (a *ClauseExtractionAgent) Name() string {
	return "ClauseExtractionAgent"
}

func (a *ClauseExtractionAgent) Process(state *model.ReviewState) error {
	log.Printf("[%s] 开始处理 reviewId=%s", a.Name(), state.ReviewID)

	if state.RawText == "" {
		state.Errors = append(state.Errors, "合同文本为空")
		return nil
	}

	text := state.RawText
	if len(text) > 15000 {
		text = text[:15000] + "\n...(已截断)"
	}

	systemPrompt := `你是专业的法律合同条款提取专家。请提取条款信息，输出JSON：
{"contract_type":"类型","clauses":[{"title":"","content":"","category":"","section_number":""}],"entities":[{"entity_type":"","value":"","location":""}]}
category可选: payment,liability,confidentiality,termination,intellectual_property,dispute_resolution,force_majeure,warranty,indemnification,governing_law,other`

	response, err := a.llmClient.Chat(systemPrompt, "请提取以下合同的条款信息：\n\n"+text)
	if err != nil {
		log.Printf("[%s] LLM调用失败: %v, 使用回退方案", a.Name(), err)
		state.Clauses = fallbackExtraction(state.RawText)
		state.ContractType = "未识别"
		state.Errors = append(state.Errors, fmt.Sprintf("条款提取LLM失败: %v", err))
		return nil
	}

	jsonStr := extractJSON(response)
	var result struct {
		ContractType string `json:"contract_type"`
		Clauses      []struct {
			Title         string `json:"title"`
			Content       string `json:"content"`
			Category      string `json:"category"`
			SectionNumber string `json:"section_number"`
		} `json:"clauses"`
		Entities []struct {
			EntityType string `json:"entity_type"`
			Value      string `json:"value"`
			Location   string `json:"location"`
		} `json:"entities"`
	}

	if err := json.Unmarshal([]byte(jsonStr), &result); err != nil {
		log.Printf("[%s] JSON解析失败: %v", a.Name(), err)
		state.Clauses = fallbackExtraction(state.RawText)
		state.ContractType = "未识别"
		return nil
	}

	state.ContractType = result.ContractType
	for _, c := range result.Clauses {
		state.Clauses = append(state.Clauses, model.Clause{
			ID:            uuid.New().String()[:8],
			Title:         c.Title,
			Content:       c.Content,
			Category:      model.ClauseCategory(c.Category),
			SectionNumber: c.SectionNumber,
		})
	}
	for _, e := range result.Entities {
		state.Entities = append(state.Entities, model.ContractEntity{
			EntityType: e.EntityType,
			Value:      e.Value,
			Location:   e.Location,
		})
	}

	state.Status = model.StatusInProgress
	log.Printf("[%s] 完成：%d个条款，%d个实体", a.Name(), len(state.Clauses), len(state.Entities))
	return nil
}

func extractJSON(text string) string {
	start := strings.Index(text, "{")
	end := strings.LastIndex(text, "}")
	if start >= 0 && end > start {
		return text[start : end+1]
	}
	return text
}

func fallbackExtraction(text string) []model.Clause {
	var clauses []model.Clause
	pattern := regexp.MustCompile(`(第[一二三四五六七八九十百]+条\s*.+)`)
	matches := pattern.FindAllStringIndex(text, -1)

	if len(matches) == 0 && text != "" {
		clauses = append(clauses, model.Clause{
			ID:       uuid.New().String()[:8],
			Title:    "全文",
			Content:  text,
			Category: model.CatOther,
		})
		return clauses
	}

	for i, match := range matches {
		title := text[match[0]:match[1]]
		var content string
		if i+1 < len(matches) {
			content = strings.TrimSpace(text[match[1]:matches[i+1][0]])
		} else {
			content = strings.TrimSpace(text[match[1]:])
		}
		clauses = append(clauses, model.Clause{
			ID:       uuid.New().String()[:8],
			Title:    strings.TrimSpace(title),
			Content:  content,
			Category: model.CatOther,
		})
	}
	return clauses
}

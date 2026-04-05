package model

import (
	"time"

	"github.com/google/uuid"
)

// ReviewStatus 审查状态枚举
type ReviewStatus string

const (
	StatusPending       ReviewStatus = "pending"
	StatusInProgress    ReviewStatus = "in_progress"
	StatusAwaitingHuman ReviewStatus = "awaiting_human"
	StatusApproved      ReviewStatus = "approved"
	StatusRejected      ReviewStatus = "rejected"
	StatusCompleted     ReviewStatus = "completed"
)

// RiskLevel 风险等级枚举
type RiskLevel string

const (
	RiskHigh   RiskLevel = "high"
	RiskMedium RiskLevel = "medium"
	RiskLow    RiskLevel = "low"
	RiskNone   RiskLevel = "none"
)

// ComplianceStatus 合规状态枚举
type ComplianceStatus string

const (
	Compliant    ComplianceStatus = "compliant"
	NonCompliant ComplianceStatus = "non_compliant"
	NeedsReview  ComplianceStatus = "needs_review"
)

// ClauseCategory 条款类别枚举
type ClauseCategory string

const (
	CatPayment            ClauseCategory = "payment"
	CatLiability          ClauseCategory = "liability"
	CatConfidentiality    ClauseCategory = "confidentiality"
	CatTermination        ClauseCategory = "termination"
	CatIP                 ClauseCategory = "intellectual_property"
	CatDisputeResolution  ClauseCategory = "dispute_resolution"
	CatForceMajeure       ClauseCategory = "force_majeure"
	CatWarranty           ClauseCategory = "warranty"
	CatIndemnification    ClauseCategory = "indemnification"
	CatGoverningLaw       ClauseCategory = "governing_law"
	CatOther              ClauseCategory = "other"
)

// Clause 结构化条款
type Clause struct {
	ID            string         `json:"id"`
	Title         string         `json:"title"`
	Content       string         `json:"content"`
	Category      ClauseCategory `json:"category"`
	SectionNumber string         `json:"section_number"`
}

// ContractEntity 合同实体
type ContractEntity struct {
	EntityType string `json:"entity_type"`
	Value      string `json:"value"`
	Location   string `json:"location"`
}

// RiskFinding 风险发现
type RiskFinding struct {
	ClauseID     string    `json:"clause_id"`
	RiskLevel    RiskLevel `json:"risk_level"`
	RiskType     string    `json:"risk_type"`
	Description  string    `json:"description"`
	BuyerImpact  string    `json:"buyer_impact"`
	SellerImpact string    `json:"seller_impact"`
	Rationale    string    `json:"rationale"`
}

// ComplianceFinding 合规发现
type ComplianceFinding struct {
	ClauseID       string           `json:"clause_id"`
	Status         ComplianceStatus `json:"status"`
	Regulation     string           `json:"regulation"`
	Issue          string           `json:"issue"`
	Recommendation string           `json:"recommendation"`
}

// Suggestion 修改建议
type Suggestion struct {
	ClauseID      string    `json:"clause_id"`
	OriginalText  string    `json:"original_text"`
	SuggestedText string    `json:"suggested_text"`
	Reason        string    `json:"reason"`
	Priority      RiskLevel `json:"priority"`
}

// HumanFeedback 人工反馈
type HumanFeedback struct {
	Reviewer  string    `json:"reviewer"`
	Decision  string    `json:"decision"`
	Comments  string    `json:"comments"`
	Timestamp time.Time `json:"timestamp"`
}

// ReviewState 流水线共享状态 — 对应Python/Java版的同名结构
type ReviewState struct {
	ReviewID   string       `json:"review_id"`
	Status     ReviewStatus `json:"status"`
	CreatedAt  time.Time    `json:"created_at"`

	DocumentPath string `json:"document_path"`
	RawText      string `json:"raw_text"`

	// Agent 1
	Clauses      []Clause         `json:"clauses"`
	Entities     []ContractEntity `json:"entities"`
	ContractType string           `json:"contract_type"`

	// Agent 2
	RiskFindings     []RiskFinding `json:"risk_findings"`
	OverallRiskLevel RiskLevel     `json:"overall_risk_level"`
	RiskSummary      string        `json:"risk_summary"`

	// Agent 3
	ComplianceFindings []ComplianceFinding `json:"compliance_findings"`
	OverallCompliance  ComplianceStatus    `json:"overall_compliance"`
	MissingClauses     []string            `json:"missing_clauses"`

	// Agent 4
	Suggestions []Suggestion `json:"suggestions"`
	VersionDiff string       `json:"version_diff"`

	NeedsHumanReview bool           `json:"needs_human_review"`
	HumanFeedback    *HumanFeedback `json:"human_feedback"`
	Errors           []string       `json:"errors"`
}

// NewReviewState 创建新的审查状态
func NewReviewState() *ReviewState {
	return &ReviewState{
		ReviewID:           uuid.New().String(),
		Status:             StatusPending,
		CreatedAt:          time.Now(),
		Clauses:            make([]Clause, 0),
		Entities:           make([]ContractEntity, 0),
		RiskFindings:       make([]RiskFinding, 0),
		OverallRiskLevel:   RiskNone,
		ComplianceFindings: make([]ComplianceFinding, 0),
		OverallCompliance:  NeedsReview,
		MissingClauses:     make([]string, 0),
		Suggestions:        make([]Suggestion, 0),
		Errors:             make([]string, 0),
	}
}

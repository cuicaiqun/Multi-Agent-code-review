package handler

import (
	"net/http"
	"sync"

	"github.com/bcefghj/multi-agent-contract-review/golang/internal/agent"
	"github.com/bcefghj/multi-agent-contract-review/golang/internal/model"
	"github.com/gin-gonic/gin"
)

var (
	store    = make(map[string]*model.ReviewState)
	storeMu  sync.RWMutex
	pipeline = agent.NewPipeline()
)

// SetupRouter 配置HTTP路由
func SetupRouter() *gin.Engine {
	r := gin.Default()

	r.Use(corsMiddleware())

	v1 := r.Group("/api/v1")
	{
		v1.GET("/health", healthHandler)
		v1.POST("/review", createReviewHandler)
		v1.GET("/review/:id", getReviewHandler)
		v1.POST("/review/:id/feedback", feedbackHandler)
	}

	return r
}

func corsMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Header("Access-Control-Allow-Origin", "*")
		c.Header("Access-Control-Allow-Methods", "GET, POST, OPTIONS")
		c.Header("Access-Control-Allow-Headers", "Content-Type")
		if c.Request.Method == "OPTIONS" {
			c.AbortWithStatus(204)
			return
		}
		c.Next()
	}
}

func healthHandler(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"status":  "healthy",
		"service": "contract-review-go",
		"version": "1.0.0",
	})
}

type reviewRequest struct {
	Text            string `json:"text" binding:"required"`
	WithHumanReview bool   `json:"with_human_review"`
}

func createReviewHandler(c *gin.Context) {
	var req reviewRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "合同文本不能为空"})
		return
	}

	state := model.NewReviewState()
	state.RawText = req.Text

	if err := pipeline.Execute(state); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	storeMu.Lock()
	store[state.ReviewID] = state
	storeMu.Unlock()

	c.JSON(http.StatusOK, gin.H{
		"review_id":          state.ReviewID,
		"status":             state.Status,
		"contract_type":      state.ContractType,
		"clauses_count":      len(state.Clauses),
		"risk_summary":       state.RiskSummary,
		"overall_risk_level": state.OverallRiskLevel,
		"overall_compliance": state.OverallCompliance,
		"suggestions_count":  len(state.Suggestions),
		"needs_human_review": state.NeedsHumanReview,
	})
}

func getReviewHandler(c *gin.Context) {
	id := c.Param("id")

	storeMu.RLock()
	state, ok := store[id]
	storeMu.RUnlock()

	if !ok {
		c.JSON(http.StatusNotFound, gin.H{"error": "审查记录不存在"})
		return
	}
	c.JSON(http.StatusOK, state)
}

type feedbackRequest struct {
	Reviewer string `json:"reviewer"`
	Decision string `json:"decision"`
	Comments string `json:"comments"`
}

func feedbackHandler(c *gin.Context) {
	id := c.Param("id")

	storeMu.Lock()
	defer storeMu.Unlock()

	state, ok := store[id]
	if !ok {
		c.JSON(http.StatusNotFound, gin.H{"error": "审查记录不存在"})
		return
	}

	var req feedbackRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "请求格式错误"})
		return
	}

	state.HumanFeedback = &model.HumanFeedback{
		Reviewer: req.Reviewer,
		Decision: req.Decision,
		Comments: req.Comments,
	}

	switch req.Decision {
	case "approve":
		state.Status = model.StatusApproved
	case "reject":
		state.Status = model.StatusRejected
	default:
		state.Status = model.StatusCompleted
	}

	c.JSON(http.StatusOK, gin.H{"message": "反馈提交成功", "status": state.Status})
}

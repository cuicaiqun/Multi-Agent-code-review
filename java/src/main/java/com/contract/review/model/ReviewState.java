package com.contract.review.model;

import lombok.AllArgsConstructor;
import lombok.Builder;
import lombok.Data;
import lombok.NoArgsConstructor;

import java.time.LocalDateTime;
import java.util.ArrayList;
import java.util.List;
import java.util.UUID;

/**
 * 合同审查流水线共享状态。
 * 对应Python版的ContractReviewState，Java版使用POJO + Builder模式。
 */
@Data
@Builder
@NoArgsConstructor
@AllArgsConstructor
public class ReviewState {

    @Builder.Default
    private String reviewId = UUID.randomUUID().toString();

    @Builder.Default
    private ReviewStatus status = ReviewStatus.PENDING;

    @Builder.Default
    private LocalDateTime createdAt = LocalDateTime.now();

    private String documentPath;
    private String rawText;

    // Agent 1: 条款提取
    @Builder.Default
    private List<Clause> clauses = new ArrayList<>();

    @Builder.Default
    private List<ContractEntity> entities = new ArrayList<>();

    private String contractType;

    // Agent 2: 风险识别
    @Builder.Default
    private List<RiskFinding> riskFindings = new ArrayList<>();

    @Builder.Default
    private RiskLevel overallRiskLevel = RiskLevel.NONE;

    private String riskSummary;

    // Agent 3: 合规检查
    @Builder.Default
    private List<ComplianceFinding> complianceFindings = new ArrayList<>();

    @Builder.Default
    private ComplianceStatus overallCompliance = ComplianceStatus.NEEDS_REVIEW;

    @Builder.Default
    private List<String> missingClauses = new ArrayList<>();

    // Agent 4: 修改建议
    @Builder.Default
    private List<Suggestion> suggestions = new ArrayList<>();

    private String versionDiff;

    // 人机协同
    @Builder.Default
    private boolean needsHumanReview = false;

    private HumanFeedback humanFeedback;

    @Builder.Default
    private List<String> errors = new ArrayList<>();

    // ── 枚举类型 ──

    public enum ReviewStatus {
        PENDING, IN_PROGRESS, AWAITING_HUMAN, APPROVED, REJECTED, COMPLETED
    }

    public enum RiskLevel {
        HIGH, MEDIUM, LOW, NONE
    }

    public enum ComplianceStatus {
        COMPLIANT, NON_COMPLIANT, NEEDS_REVIEW
    }

    public enum ClauseCategory {
        PAYMENT, LIABILITY, CONFIDENTIALITY, TERMINATION,
        INTELLECTUAL_PROPERTY, DISPUTE_RESOLUTION, FORCE_MAJEURE,
        WARRANTY, INDEMNIFICATION, GOVERNING_LAW, OTHER
    }

    // ── 内部数据类 ──

    @Data
    @Builder
    @NoArgsConstructor
    @AllArgsConstructor
    public static class Clause {
        @Builder.Default
        private String id = UUID.randomUUID().toString().substring(0, 8);
        private String title;
        private String content;
        @Builder.Default
        private ClauseCategory category = ClauseCategory.OTHER;
        private String sectionNumber;
    }

    @Data
    @Builder
    @NoArgsConstructor
    @AllArgsConstructor
    public static class ContractEntity {
        private String entityType;
        private String value;
        private String location;
    }

    @Data
    @Builder
    @NoArgsConstructor
    @AllArgsConstructor
    public static class RiskFinding {
        private String clauseId;
        private RiskLevel riskLevel;
        private String riskType;
        private String description;
        private String buyerImpact;
        private String sellerImpact;
        private String rationale;
    }

    @Data
    @Builder
    @NoArgsConstructor
    @AllArgsConstructor
    public static class ComplianceFinding {
        private String clauseId;
        private ComplianceStatus status;
        private String regulation;
        private String issue;
        private String recommendation;
    }

    @Data
    @Builder
    @NoArgsConstructor
    @AllArgsConstructor
    public static class Suggestion {
        private String clauseId;
        private String originalText;
        private String suggestedText;
        private String reason;
        private RiskLevel priority;
    }

    @Data
    @Builder
    @NoArgsConstructor
    @AllArgsConstructor
    public static class HumanFeedback {
        private String reviewer;
        private String decision;
        private String comments;
        private LocalDateTime timestamp;
    }
}

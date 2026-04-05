package com.contract.review.controller;

import com.contract.review.agent.ReviewPipeline;
import com.contract.review.model.ReviewState;
import lombok.RequiredArgsConstructor;
import lombok.extern.slf4j.Slf4j;
import org.springframework.http.ResponseEntity;
import org.springframework.web.bind.annotation.*;
import org.springframework.web.multipart.MultipartFile;

import java.util.Map;
import java.util.concurrent.ConcurrentHashMap;

/**
 * 合同审查REST API — Java版。
 */
@Slf4j
@RestController
@RequestMapping("/api/v1")
@RequiredArgsConstructor
@CrossOrigin(origins = "*")
public class ReviewController {

    private final ReviewPipeline pipeline;
    private final ConcurrentHashMap<String, ReviewState> store = new ConcurrentHashMap<>();

    @GetMapping("/health")
    public Map<String, String> health() {
        return Map.of("status", "healthy", "service", "contract-review-java", "version", "1.0.0");
    }

    @PostMapping("/review")
    public ResponseEntity<?> createReview(@RequestBody Map<String, Object> request) {
        String text = (String) request.getOrDefault("text", "");
        if (text.isBlank()) {
            return ResponseEntity.badRequest().body(Map.of("error", "合同文本不能为空"));
        }

        ReviewState state = ReviewState.builder()
                .rawText(text)
                .status(ReviewState.ReviewStatus.IN_PROGRESS)
                .build();

        state = pipeline.execute(state);
        store.put(state.getReviewId(), state);

        return ResponseEntity.ok(buildSummary(state));
    }

    @GetMapping("/review/{reviewId}")
    public ResponseEntity<?> getReview(@PathVariable String reviewId) {
        ReviewState state = store.get(reviewId);
        if (state == null) {
            return ResponseEntity.notFound().build();
        }
        return ResponseEntity.ok(state);
    }

    @PostMapping("/review/{reviewId}/feedback")
    public ResponseEntity<?> submitFeedback(
            @PathVariable String reviewId,
            @RequestBody Map<String, String> request) {
        ReviewState state = store.get(reviewId);
        if (state == null) {
            return ResponseEntity.notFound().build();
        }

        state.setHumanFeedback(ReviewState.HumanFeedback.builder()
                .reviewer(request.getOrDefault("reviewer", ""))
                .decision(request.getOrDefault("decision", ""))
                .comments(request.getOrDefault("comments", ""))
                .build());

        String decision = request.getOrDefault("decision", "");
        switch (decision) {
            case "approve" -> state.setStatus(ReviewState.ReviewStatus.APPROVED);
            case "reject" -> state.setStatus(ReviewState.ReviewStatus.REJECTED);
            default -> state.setStatus(ReviewState.ReviewStatus.COMPLETED);
        }

        store.put(reviewId, state);
        return ResponseEntity.ok(Map.of("message", "反馈提交成功", "status", state.getStatus().name()));
    }

    private Map<String, Object> buildSummary(ReviewState state) {
        return Map.of(
                "review_id", state.getReviewId(),
                "status", state.getStatus().name(),
                "contract_type", state.getContractType() != null ? state.getContractType() : "",
                "clauses_count", state.getClauses().size(),
                "risk_summary", state.getRiskSummary() != null ? state.getRiskSummary() : "",
                "overall_risk_level", state.getOverallRiskLevel().name(),
                "overall_compliance", state.getOverallCompliance().name(),
                "suggestions_count", state.getSuggestions().size(),
                "needs_human_review", state.isNeedsHumanReview()
        );
    }
}

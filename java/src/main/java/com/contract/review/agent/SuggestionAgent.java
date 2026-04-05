package com.contract.review.agent;

import com.contract.review.model.ReviewState;
import com.contract.review.model.ReviewState.*;
import com.fasterxml.jackson.core.type.TypeReference;
import com.fasterxml.jackson.databind.ObjectMapper;
import lombok.RequiredArgsConstructor;
import lombok.extern.slf4j.Slf4j;
import org.springframework.ai.chat.client.ChatClient;
import org.springframework.stereotype.Component;

import java.util.*;

/**
 * 修改建议Agent — Java版。
 */
@Slf4j
@Component
@RequiredArgsConstructor
public class SuggestionAgent implements ContractAgent {

    private final ChatClient.Builder chatClientBuilder;
    private final ObjectMapper objectMapper;

    private static final String SYSTEM_PROMPT = """
            你是合同修改顾问。根据风险和合规问题生成具体修改建议。
            输出JSON格式：
            {
              "suggestions": [{"clause_id":"","original_text":"",
                "suggested_text":"","reason":"","priority":"high/medium/low"}]
            }
            """;

    @Override
    public ReviewState process(ReviewState state) {
        log.info("修改建议Agent开始处理 reviewId={}", state.getReviewId());

        if (state.getRiskFindings().isEmpty() && state.getComplianceFindings().isEmpty()) {
            state.setVersionDiff("无需修改");
            state.setStatus(ReviewState.ReviewStatus.COMPLETED);
            return state;
        }

        try {
            Map<String, Clause> clauseMap = new HashMap<>();
            state.getClauses().forEach(c -> clauseMap.put(c.getId(), c));

            StringBuilder context = new StringBuilder();
            for (RiskFinding rf : state.getRiskFindings()) {
                if (rf.getRiskLevel() == RiskLevel.HIGH || rf.getRiskLevel() == RiskLevel.MEDIUM) {
                    Clause clause = clauseMap.get(rf.getClauseId());
                    String text = clause != null ? clause.getTitle() + ": " + clause.getContent() : "";
                    context.append(String.format("[风险-%s] %s\n%s\n---\n", rf.getRiskLevel(), text, rf.getDescription()));
                }
            }
            for (ComplianceFinding cf : state.getComplianceFindings()) {
                if (cf.getStatus() != ComplianceStatus.COMPLIANT) {
                    context.append(String.format("[合规-%s] %s %s\n---\n", cf.getStatus(), cf.getIssue(), cf.getRegulation()));
                }
            }

            ChatClient chatClient = chatClientBuilder.build();
            String response = chatClient.prompt()
                    .system(SYSTEM_PROMPT)
                    .user("生成修改建议：\n\n" + context)
                    .call()
                    .content();

            String json = response.contains("```") ?
                    response.substring(response.indexOf("{"), response.lastIndexOf("}") + 1) :
                    response.trim();

            Map<String, Object> result = objectMapper.readValue(json, new TypeReference<>() {});
            state.setSuggestions(parseSuggestions(result));

            for (String missing : state.getMissingClauses()) {
                state.getSuggestions().add(Suggestion.builder()
                        .clauseId("new")
                        .originalText("（缺失）")
                        .suggestedText("建议添加" + missing + "相关条款")
                        .reason("合同缺少必要的「" + missing + "」条款")
                        .priority(RiskLevel.HIGH)
                        .build());
            }

            state.getSuggestions().sort(Comparator.comparingInt(s -> s.getPriority().ordinal()));

            boolean needsHuman = state.isNeedsHumanReview() ||
                    state.getSuggestions().stream().anyMatch(s -> s.getPriority() == RiskLevel.HIGH);

            state.setNeedsHumanReview(needsHuman);
            state.setStatus(needsHuman ? ReviewState.ReviewStatus.AWAITING_HUMAN : ReviewState.ReviewStatus.COMPLETED);

        } catch (Exception e) {
            log.error("修改建议生成失败", e);
            state.getErrors().add("建议生成失败: " + e.getMessage());
            state.setStatus(ReviewState.ReviewStatus.COMPLETED);
        }

        log.info("修改建议完成：{}条", state.getSuggestions().size());
        return state;
    }

    @Override
    public String getName() {
        return "SuggestionAgent";
    }

    @SuppressWarnings("unchecked")
    private List<Suggestion> parseSuggestions(Map<String, Object> result) {
        List<Suggestion> suggestions = new ArrayList<>();
        for (Map<String, String> item : (List<Map<String, String>>) result.getOrDefault("suggestions", List.of())) {
            RiskLevel priority;
            try {
                priority = RiskLevel.valueOf(item.getOrDefault("priority", "MEDIUM").toUpperCase());
            } catch (IllegalArgumentException e) {
                priority = RiskLevel.MEDIUM;
            }
            suggestions.add(Suggestion.builder()
                    .clauseId(item.getOrDefault("clause_id", ""))
                    .originalText(item.getOrDefault("original_text", ""))
                    .suggestedText(item.getOrDefault("suggested_text", ""))
                    .reason(item.getOrDefault("reason", ""))
                    .priority(priority)
                    .build());
        }
        return suggestions;
    }
}

package com.contract.review.agent;

import com.contract.review.model.ReviewState;
import com.contract.review.model.ReviewState.*;
import com.contract.review.rule.ContractRuleEngine;
import com.fasterxml.jackson.core.type.TypeReference;
import com.fasterxml.jackson.databind.ObjectMapper;
import lombok.RequiredArgsConstructor;
import lombok.extern.slf4j.Slf4j;
import org.springframework.ai.chat.client.ChatClient;
import org.springframework.stereotype.Component;

import java.util.*;

/**
 * 风险识别Agent — Java版。
 *
 * 双引擎架构：Drools规则引擎 + Spring AI大模型分析。
 * 面试亮点：规则引擎与AI结合、Drools的Rete算法。
 */
@Slf4j
@Component
@RequiredArgsConstructor
public class RiskIdentificationAgent implements ContractAgent {

    private final ChatClient.Builder chatClientBuilder;
    private final ObjectMapper objectMapper;
    private final ContractRuleEngine ruleEngine;

    private static final String SYSTEM_PROMPT = """
            你是一名资深法律风险评估专家。请对合同条款进行风险评估。
            输出JSON格式：
            {
              "findings": [
                {"clause_id":"","risk_level":"high/medium/low/none",
                 "risk_type":"","description":"","rationale":""}
              ],
              "overall_risk_level": "high/medium/low/none",
              "risk_summary": "50字以内总结"
            }
            """;

    @Override
    public ReviewState process(ReviewState state) {
        log.info("风险识别Agent开始处理 reviewId={}", state.getReviewId());

        if (state.getClauses().isEmpty()) {
            state.setOverallRiskLevel(RiskLevel.NONE);
            state.setRiskSummary("无条款可供分析");
            return state;
        }

        List<RiskFinding> ruleFindings = ruleEngine.scanClauses(state.getClauses());

        List<RiskFinding> llmFindings = new ArrayList<>();
        try {
            llmFindings = analyzeWithLlm(state.getClauses());
        } catch (Exception e) {
            log.error("LLM风险分析失败", e);
            state.getErrors().add("风险识别LLM分析失败: " + e.getMessage());
        }

        List<RiskFinding> merged = mergeFindings(ruleFindings, llmFindings);
        RiskLevel overall = calculateOverallRisk(merged);

        state.setRiskFindings(merged);
        state.setOverallRiskLevel(overall);
        state.setRiskSummary(generateSummary(merged, overall));
        state.setNeedsHumanReview(overall == RiskLevel.HIGH);

        log.info("风险识别完成：{}个发现，整体={}", merged.size(), overall);
        return state;
    }

    @Override
    public String getName() {
        return "RiskIdentificationAgent";
    }

    @SuppressWarnings("unchecked")
    private List<RiskFinding> analyzeWithLlm(List<Clause> clauses) throws Exception {
        StringBuilder sb = new StringBuilder();
        for (Clause c : clauses) {
            sb.append(String.format("[条款ID: %s] %s\n%s\n\n", c.getId(), c.getTitle(), c.getContent()));
        }

        ChatClient chatClient = chatClientBuilder.build();
        String response = chatClient.prompt()
                .system(SYSTEM_PROMPT)
                .user("请分析以下条款的风险：\n\n" + sb)
                .call()
                .content();

        String json = response.contains("```") ?
                response.substring(response.indexOf("{"), response.lastIndexOf("}") + 1) :
                response.trim();

        Map<String, Object> result = objectMapper.readValue(json, new TypeReference<>() {});
        List<RiskFinding> findings = new ArrayList<>();
        for (Map<String, String> item : (List<Map<String, String>>) result.getOrDefault("findings", List.of())) {
            RiskLevel level;
            try {
                level = RiskLevel.valueOf(item.getOrDefault("risk_level", "NONE").toUpperCase());
            } catch (IllegalArgumentException e) {
                level = RiskLevel.LOW;
            }
            findings.add(RiskFinding.builder()
                    .clauseId(item.getOrDefault("clause_id", ""))
                    .riskLevel(level)
                    .riskType(item.getOrDefault("risk_type", ""))
                    .description(item.getOrDefault("description", ""))
                    .rationale(item.getOrDefault("rationale", ""))
                    .build());
        }
        return findings;
    }

    private List<RiskFinding> mergeFindings(List<RiskFinding> rule, List<RiskFinding> llm) {
        Map<String, RiskFinding> seen = new LinkedHashMap<>();
        for (RiskFinding f : rule) {
            seen.put(f.getClauseId() + "|" + f.getRiskType(), f);
        }
        for (RiskFinding f : llm) {
            String key = f.getClauseId() + "|" + f.getRiskType();
            if (!seen.containsKey(key) ||
                    f.getRiskLevel().ordinal() < seen.get(key).getRiskLevel().ordinal()) {
                seen.put(key, f);
            }
        }
        return new ArrayList<>(seen.values());
    }

    private RiskLevel calculateOverallRisk(List<RiskFinding> findings) {
        if (findings.stream().anyMatch(f -> f.getRiskLevel() == RiskLevel.HIGH)) return RiskLevel.HIGH;
        if (findings.stream().anyMatch(f -> f.getRiskLevel() == RiskLevel.MEDIUM)) return RiskLevel.MEDIUM;
        if (findings.stream().anyMatch(f -> f.getRiskLevel() == RiskLevel.LOW)) return RiskLevel.LOW;
        return RiskLevel.NONE;
    }

    private String generateSummary(List<RiskFinding> findings, RiskLevel overall) {
        long high = findings.stream().filter(f -> f.getRiskLevel() == RiskLevel.HIGH).count();
        long medium = findings.stream().filter(f -> f.getRiskLevel() == RiskLevel.MEDIUM).count();
        long low = findings.stream().filter(f -> f.getRiskLevel() == RiskLevel.LOW).count();
        return String.format("整体风险：%s。%d个高风险、%d个中风险、%d个低风险。",
                overall, high, medium, low);
    }
}

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
 * 合规检查Agent — Java版。
 *
 * 面试亮点：法规知识库的结构化管理、必要条款检测逻辑。
 */
@Slf4j
@Component
@RequiredArgsConstructor
public class ComplianceCheckAgent implements ContractAgent {

    private final ChatClient.Builder chatClientBuilder;
    private final ObjectMapper objectMapper;

    private static final String SYSTEM_PROMPT = """
            你是中国法律合规检查专家。检查合同条款是否符合法律法规要求。
            输出JSON格式：
            {
              "findings": [{"clause_id":"","status":"compliant/non_compliant/needs_review",
                "regulation":"","issue":"","recommendation":""}],
              "missing_clauses": ["缺失条款"],
              "overall_compliance": "compliant/non_compliant/needs_review"
            }
            """;

    private static final Map<String, List<String>> REQUIRED_CLAUSES = Map.of(
            "general", List.of("当事人信息", "标的/服务内容", "价款/报酬", "履行期限", "违约责任", "争议解决")
    );

    @Override
    public ReviewState process(ReviewState state) {
        log.info("合规检查Agent开始处理 reviewId={}", state.getReviewId());

        if (state.getClauses().isEmpty()) {
            state.setOverallCompliance(ComplianceStatus.NEEDS_REVIEW);
            return state;
        }

        List<String> missing = checkMissingClauses(state.getClauses());
        state.setMissingClauses(missing);

        try {
            ChatClient chatClient = chatClientBuilder.build();
            StringBuilder sb = new StringBuilder();
            for (Clause c : state.getClauses()) {
                sb.append(String.format("[%s] %s\n%s\n\n", c.getId(), c.getTitle(), c.getContent()));
            }

            String response = chatClient.prompt()
                    .system(SYSTEM_PROMPT)
                    .user("合同类型：" + state.getContractType() + "\n\n" + sb)
                    .call()
                    .content();

            String json = response.contains("```") ?
                    response.substring(response.indexOf("{"), response.lastIndexOf("}") + 1) :
                    response.trim();

            Map<String, Object> result = objectMapper.readValue(json, new TypeReference<>() {});
            state.setComplianceFindings(parseFindings(result));
            state.setOverallCompliance(calculateOverall(state.getComplianceFindings(), missing));

        } catch (Exception e) {
            log.error("合规检查LLM失败", e);
            state.getErrors().add("合规检查失败: " + e.getMessage());
            state.setOverallCompliance(ComplianceStatus.NEEDS_REVIEW);
        }

        log.info("合规检查完成：{}个发现，{}个缺失", state.getComplianceFindings().size(), missing.size());
        return state;
    }

    @Override
    public String getName() {
        return "ComplianceCheckAgent";
    }

    private List<String> checkMissingClauses(List<Clause> clauses) {
        List<String> required = REQUIRED_CLAUSES.getOrDefault("general", List.of());
        String allText = clauses.stream()
                .map(c -> c.getTitle() + " " + c.getContent().substring(0, Math.min(50, c.getContent().length())))
                .reduce("", (a, b) -> a + " " + b);

        Map<String, List<String>> keywordMap = Map.of(
                "当事人信息", List.of("甲方", "乙方", "当事人"),
                "标的/服务内容", List.of("标的", "服务", "商品"),
                "价款/报酬", List.of("价款", "报酬", "费用", "金额"),
                "履行期限", List.of("期限", "交付", "完成"),
                "违约责任", List.of("违约", "责任", "赔偿"),
                "争议解决", List.of("争议", "仲裁", "诉讼")
        );

        List<String> missing = new ArrayList<>();
        for (String req : required) {
            List<String> keywords = keywordMap.getOrDefault(req, List.of(req));
            boolean found = keywords.stream().anyMatch(allText::contains);
            if (!found) missing.add(req);
        }
        return missing;
    }

    @SuppressWarnings("unchecked")
    private List<ComplianceFinding> parseFindings(Map<String, Object> result) {
        List<ComplianceFinding> findings = new ArrayList<>();
        for (Map<String, String> item : (List<Map<String, String>>) result.getOrDefault("findings", List.of())) {
            ComplianceStatus status;
            try {
                status = ComplianceStatus.valueOf(item.getOrDefault("status", "NEEDS_REVIEW").toUpperCase());
            } catch (IllegalArgumentException e) {
                status = ComplianceStatus.NEEDS_REVIEW;
            }
            if (status != ComplianceStatus.COMPLIANT) {
                findings.add(ComplianceFinding.builder()
                        .clauseId(item.getOrDefault("clause_id", ""))
                        .status(status)
                        .regulation(item.getOrDefault("regulation", ""))
                        .issue(item.getOrDefault("issue", ""))
                        .recommendation(item.getOrDefault("recommendation", ""))
                        .build());
            }
        }
        return findings;
    }

    private ComplianceStatus calculateOverall(List<ComplianceFinding> findings, List<String> missing) {
        if (findings.stream().anyMatch(f -> f.getStatus() == ComplianceStatus.NON_COMPLIANT))
            return ComplianceStatus.NON_COMPLIANT;
        if (!missing.isEmpty() || findings.stream().anyMatch(f -> f.getStatus() == ComplianceStatus.NEEDS_REVIEW))
            return ComplianceStatus.NEEDS_REVIEW;
        return ComplianceStatus.COMPLIANT;
    }
}

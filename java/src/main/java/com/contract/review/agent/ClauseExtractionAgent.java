package com.contract.review.agent;

import com.contract.review.model.ReviewState;
import com.contract.review.model.ReviewState.*;
import com.fasterxml.jackson.core.type.TypeReference;
import com.fasterxml.jackson.databind.ObjectMapper;
import lombok.RequiredArgsConstructor;
import lombok.extern.slf4j.Slf4j;
import org.springframework.ai.chat.client.ChatClient;
import org.springframework.stereotype.Component;

import java.util.ArrayList;
import java.util.List;
import java.util.Map;
import java.util.regex.Matcher;
import java.util.regex.Pattern;

/**
 * 条款提取Agent — Java版。
 *
 * 使用Spring AI的ChatClient调用LLM，解析合同文本并提取结构化条款。
 * 面试亮点：与Python版LangChain的对比、Spring AI的Prompt模板机制。
 */
@Slf4j
@Component
@RequiredArgsConstructor
public class ClauseExtractionAgent implements ContractAgent {

    private final ChatClient.Builder chatClientBuilder;
    private final ObjectMapper objectMapper;

    private static final String SYSTEM_PROMPT = """
            你是一个专业的法律合同条款提取专家。请从合同文本中提取结构化条款信息。
            请严格按照JSON格式输出：
            {
              "contract_type": "合同类型",
              "clauses": [{"title":"条款标题","content":"条款内容","category":"类别","section_number":"编号"}],
              "entities": [{"entity_type":"类型","value":"值","location":"位置"}]
            }
            category可选值: payment,liability,confidentiality,termination,intellectual_property,
            dispute_resolution,force_majeure,warranty,indemnification,governing_law,other
            """;

    @Override
    public ReviewState process(ReviewState state) {
        log.info("条款提取Agent开始处理 reviewId={}", state.getReviewId());

        try {
            String rawText = state.getRawText();
            if (rawText == null || rawText.isBlank()) {
                state.getErrors().add("合同文本为空");
                return state;
            }

            String truncated = rawText.length() > 15000
                    ? rawText.substring(0, 15000) + "\n...(已截断)"
                    : rawText;

            ChatClient chatClient = chatClientBuilder.build();
            String response = chatClient.prompt()
                    .system(SYSTEM_PROMPT)
                    .user("请提取以下合同的条款信息：\n\n" + truncated)
                    .call()
                    .content();

            String json = extractJson(response);
            Map<String, Object> result = objectMapper.readValue(json, new TypeReference<>() {});

            state.setContractType((String) result.getOrDefault("contract_type", "未识别"));
            state.setClauses(parseClauses(result));
            state.setEntities(parseEntities(result));
            state.setStatus(ReviewState.ReviewStatus.IN_PROGRESS);

            log.info("条款提取完成：{}个条款，{}个实体",
                    state.getClauses().size(), state.getEntities().size());

        } catch (Exception e) {
            log.error("条款提取失败", e);
            state.getErrors().add("条款提取失败: " + e.getMessage());
            state.setClauses(fallbackExtraction(state.getRawText()));
        }
        return state;
    }

    @Override
    public String getName() {
        return "ClauseExtractionAgent";
    }

    private String extractJson(String text) {
        if (text.contains("```")) {
            int start = text.indexOf("{");
            int end = text.lastIndexOf("}");
            if (start >= 0 && end > start) {
                return text.substring(start, end + 1);
            }
        }
        return text.trim();
    }

    @SuppressWarnings("unchecked")
    private List<Clause> parseClauses(Map<String, Object> result) {
        List<Clause> clauses = new ArrayList<>();
        List<Map<String, String>> items = (List<Map<String, String>>) result.getOrDefault("clauses", List.of());
        for (Map<String, String> item : items) {
            ClauseCategory category;
            try {
                category = ClauseCategory.valueOf(item.getOrDefault("category", "OTHER").toUpperCase());
            } catch (IllegalArgumentException e) {
                category = ClauseCategory.OTHER;
            }
            clauses.add(Clause.builder()
                    .title(item.getOrDefault("title", "未命名"))
                    .content(item.getOrDefault("content", ""))
                    .category(category)
                    .sectionNumber(item.getOrDefault("section_number", ""))
                    .build());
        }
        return clauses;
    }

    @SuppressWarnings("unchecked")
    private List<ContractEntity> parseEntities(Map<String, Object> result) {
        List<ContractEntity> entities = new ArrayList<>();
        List<Map<String, String>> items = (List<Map<String, String>>) result.getOrDefault("entities", List.of());
        for (Map<String, String> item : items) {
            entities.add(ContractEntity.builder()
                    .entityType(item.getOrDefault("entity_type", ""))
                    .value(item.getOrDefault("value", ""))
                    .location(item.getOrDefault("location", ""))
                    .build());
        }
        return entities;
    }

    private List<Clause> fallbackExtraction(String text) {
        List<Clause> clauses = new ArrayList<>();
        if (text == null) return clauses;

        Pattern pattern = Pattern.compile("(第[一二三四五六七八九十百]+条\\s*.+)");
        Matcher matcher = pattern.matcher(text);
        int lastEnd = 0;
        String lastTitle = "前言";

        while (matcher.find()) {
            if (lastEnd > 0) {
                String content = text.substring(lastEnd, matcher.start()).trim();
                if (!content.isEmpty()) {
                    clauses.add(Clause.builder().title(lastTitle).content(content).build());
                }
            }
            lastTitle = matcher.group(0).trim();
            lastEnd = matcher.end();
        }
        if (lastEnd < text.length()) {
            String remaining = text.substring(lastEnd).trim();
            if (!remaining.isEmpty()) {
                clauses.add(Clause.builder().title(lastTitle).content(remaining).build());
            }
        }
        return clauses;
    }
}

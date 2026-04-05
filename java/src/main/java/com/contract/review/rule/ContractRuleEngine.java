package com.contract.review.rule;

import com.contract.review.model.ReviewState;
import com.contract.review.model.ReviewState.*;
import lombok.extern.slf4j.Slf4j;
import org.springframework.stereotype.Component;

import java.util.*;
import java.util.regex.Pattern;

/**
 * 合同规则引擎 — Java版。
 *
 * 生产环境可集成Drools实现更复杂的规则管理。
 * 当前版本使用轻量级实现，便于理解和面试讲解。
 *
 * 面试亮点：
 * - Drools的Rete算法 vs 顺序匹配
 * - 规则冲突解决策略（优先级、特异性、时间顺序）
 * - 前向链推理(forward chaining) vs 后向链推理(backward chaining)
 */
@Slf4j
@Component
public class ContractRuleEngine {

    private final List<ContractRule> rules = new ArrayList<>();

    public ContractRuleEngine() {
        loadDefaultRules();
    }

    public List<RiskFinding> scanClauses(List<Clause> clauses) {
        List<RiskFinding> findings = new ArrayList<>();
        for (Clause clause : clauses) {
            for (ContractRule rule : rules) {
                if (rule.matches(clause.getContent())) {
                    findings.add(RiskFinding.builder()
                            .clauseId(clause.getId())
                            .riskLevel(rule.riskLevel)
                            .riskType(rule.name)
                            .description(rule.description + "（命中关键词）")
                            .rationale("规则引擎检测：条款「" + clause.getTitle() + "」")
                            .build());
                    break;
                }
            }
        }
        log.info("规则引擎扫描完成：{}个条款，{}个发现", clauses.size(), findings.size());
        return findings;
    }

    private void loadDefaultRules() {
        rules.add(new ContractRule("unlimited_liability", "无限责任检查",
                "条款包含无限责任相关表述", RiskLevel.HIGH, 100,
                List.of("无限责任", "承担全部损失", "不设上限")));

        rules.add(new ContractRule("unilateral_termination", "单方解除权检查",
                "一方拥有不对等的单方解除权", RiskLevel.HIGH, 95,
                List.of("单方解除", "单方终止", "有权随时终止")));

        rules.add(new ContractRule("unfair_penalty", "违约金合理性检查",
                "违约金可能不合理", RiskLevel.MEDIUM, 80,
                List.of("双倍赔偿", "三倍赔偿", "全额赔偿")));

        rules.add(new ContractRule("format_clause_invalid", "格式条款无效检查",
                "格式条款可能无效（民法典497条）", RiskLevel.HIGH, 100,
                List.of("免除己方责任", "排除对方主要权利", "加重对方责任")));

        rules.add(new ContractRule("auto_renewal", "自动续约检查",
                "包含自动续约条款", RiskLevel.MEDIUM, 50,
                List.of("自动续约", "自动续期", "默认续签")));

        rules.add(new ContractRule("vague_terms", "模糊条款检查",
                "条款表述模糊", RiskLevel.LOW, 40,
                List.of("酌情处理", "另行协商", "视情况而定")));

        log.info("规则引擎加载 {} 条规则", rules.size());
    }

    private static class ContractRule {
        final String name;
        final String displayName;
        final String description;
        final RiskLevel riskLevel;
        final int priority;
        final List<String> keywords;

        ContractRule(String name, String displayName, String description,
                     RiskLevel riskLevel, int priority, List<String> keywords) {
            this.name = name;
            this.displayName = displayName;
            this.description = description;
            this.riskLevel = riskLevel;
            this.priority = priority;
            this.keywords = keywords;
        }

        boolean matches(String text) {
            return keywords.stream().anyMatch(text::contains);
        }
    }
}

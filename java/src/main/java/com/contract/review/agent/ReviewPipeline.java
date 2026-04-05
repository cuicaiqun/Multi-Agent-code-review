package com.contract.review.agent;

import com.contract.review.model.ReviewState;
import lombok.RequiredArgsConstructor;
import lombok.extern.slf4j.Slf4j;
import org.springframework.stereotype.Service;

import java.util.List;

/**
 * 审查流水线 — 顺序编排四个Agent。
 *
 * Java版使用Spring管理的Bean链式调用，
 * 对应Python版的LangGraph StateGraph。
 *
 * 面试亮点：
 * - 责任链模式(Chain of Responsibility) vs 策略模式
 * - 与LangGraph的对比：显式状态管理 vs 图编排
 * - Spring AI Alibaba的SequentialAgent对比
 */
@Slf4j
@Service
@RequiredArgsConstructor
public class ReviewPipeline {

    private final ClauseExtractionAgent clauseExtractionAgent;
    private final RiskIdentificationAgent riskIdentificationAgent;
    private final ComplianceCheckAgent complianceCheckAgent;
    private final SuggestionAgent suggestionAgent;

    /**
     * 执行完整的合同审查流水线。
     */
    public ReviewState execute(ReviewState state) {
        log.info("开始合同审查流水线 reviewId={}", state.getReviewId());
        state.setStatus(ReviewState.ReviewStatus.IN_PROGRESS);

        List<ContractAgent> agents = List.of(
                clauseExtractionAgent,
                riskIdentificationAgent,
                complianceCheckAgent,
                suggestionAgent
        );

        for (ContractAgent agent : agents) {
            try {
                log.info("执行 {}", agent.getName());
                state = agent.process(state);
            } catch (Exception e) {
                log.error("{} 执行失败", agent.getName(), e);
                state.getErrors().add(agent.getName() + " 失败: " + e.getMessage());
            }
        }

        if (state.getStatus() != ReviewState.ReviewStatus.AWAITING_HUMAN) {
            state.setStatus(ReviewState.ReviewStatus.COMPLETED);
        }

        log.info("审查流水线完成 reviewId={} status={}", state.getReviewId(), state.getStatus());
        return state;
    }
}

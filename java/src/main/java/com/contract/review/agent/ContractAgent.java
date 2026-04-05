package com.contract.review.agent;

import com.contract.review.model.ReviewState;

/**
 * Agent接口 — 所有合同审查Agent的统一抽象。
 *
 * 面试亮点：策略模式(Strategy Pattern)的应用，
 * 每个Agent实现相同接口，由Pipeline按序调用。
 */
public interface ContractAgent {

    /**
     * 处理审查状态，返回更新后的状态。
     */
    ReviewState process(ReviewState state);

    /**
     * Agent名称，用于日志和监控。
     */
    String getName();
}

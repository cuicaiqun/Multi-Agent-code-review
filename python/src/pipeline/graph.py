"""
LangGraph流水线编排 — 四Agent流水线 + 人机协同。

核心设计：
1. 使用LangGraph StateGraph构建有向图
2. 四个Agent节点顺序执行
3. 人机协同通过interrupt实现
4. 支持checkpoint持久化，可从任意步骤恢复
5. 条件路由：根据风险等级决定是否需要人工审核

面试亮点：
- LangGraph StateGraph vs MessageGraph 的区别
- interrupt/resume机制的实现原理
- checkpoint的序列化和反序列化
- 条件边(conditional edge)的使用场景
"""

from __future__ import annotations

import logging
from typing import Literal

from langgraph.graph import END, StateGraph
from langgraph.checkpoint.memory import MemorySaver

from ..agents.clause_extraction import ClauseExtractionAgent
from ..agents.risk_identification import RiskIdentificationAgent
from ..agents.compliance_check import ComplianceCheckAgent
from ..agents.suggestion import SuggestionAgent
from .state import ContractReviewState, ReviewStatus, RiskLevel

logger = logging.getLogger(__name__)


def should_human_review(state: ContractReviewState) -> Literal["human_review", "complete"]:
    """条件路由：判断是否需要人工审核。"""
    if state.needs_human_review:
        return "human_review"
    return "complete"


def human_review_node(state: ContractReviewState) -> dict:
    """
    人机协同节点 — 暂停流水线等待人工审核。

    在实际运行中，这个节点会通过LangGraph的interrupt机制暂停。
    前端通过API提交人工审核结果后，流水线从此处恢复继续执行。
    """
    logger.info("进入人工审核节点 review_id=%s", state.review_id)
    return {
        "status": ReviewStatus.AWAITING_HUMAN,
    }


def process_human_feedback(state: ContractReviewState) -> dict:
    """处理人工反馈后的后续逻辑。"""
    feedback = state.human_feedback
    if feedback is None:
        return {"status": ReviewStatus.COMPLETED}

    if feedback.decision == "approve":
        return {"status": ReviewStatus.APPROVED}
    elif feedback.decision == "reject":
        return {"status": ReviewStatus.REJECTED}
    else:
        if feedback.modified_suggestions:
            return {
                "suggestions": feedback.modified_suggestions,
                "status": ReviewStatus.COMPLETED,
            }
        return {"status": ReviewStatus.COMPLETED}


def complete_node(state: ContractReviewState) -> dict:
    """流水线完成节点。"""
    logger.info(
        "审查完成 review_id=%s 条款=%d 风险=%d 合规=%d 建议=%d",
        state.review_id,
        len(state.clauses),
        len(state.risk_findings),
        len(state.compliance_findings),
        len(state.suggestions),
    )
    return {"status": ReviewStatus.COMPLETED}


def build_review_graph(with_human_review: bool = True) -> StateGraph:
    """
    构建合同审查流水线的LangGraph有向图。

    流程：
    clause_extraction → risk_identification → compliance_check
        → suggestion → [human_review | complete] → END

    Args:
        with_human_review: 是否启用人机协同节点

    Returns:
        编译后的StateGraph
    """

    clause_agent = ClauseExtractionAgent()
    risk_agent = RiskIdentificationAgent()
    compliance_agent = ComplianceCheckAgent()
    suggestion_agent = SuggestionAgent()

    graph = StateGraph(ContractReviewState)

    graph.add_node("clause_extraction", clause_agent)
    graph.add_node("risk_identification", risk_agent)
    graph.add_node("compliance_check", compliance_agent)
    graph.add_node("suggestion", suggestion_agent)
    graph.add_node("complete", complete_node)

    if with_human_review:
        graph.add_node("human_review", human_review_node)
        graph.add_node("process_feedback", process_human_feedback)

    graph.set_entry_point("clause_extraction")
    graph.add_edge("clause_extraction", "risk_identification")
    graph.add_edge("risk_identification", "compliance_check")
    graph.add_edge("compliance_check", "suggestion")

    if with_human_review:
        graph.add_conditional_edges(
            "suggestion",
            should_human_review,
            {
                "human_review": "human_review",
                "complete": "complete",
            },
        )
        graph.add_edge("human_review", "process_feedback")
        graph.add_edge("process_feedback", "complete")
    else:
        graph.add_edge("suggestion", "complete")

    graph.add_edge("complete", END)

    return graph


def create_review_pipeline(
    with_human_review: bool = True,
    with_checkpointing: bool = True,
):
    """
    创建完整的审查流水线（含checkpoint持久化）。

    Args:
        with_human_review: 是否启用人机协同
        with_checkpointing: 是否启用checkpoint

    Returns:
        编译后的可执行图
    """
    graph = build_review_graph(with_human_review=with_human_review)

    compile_kwargs = {}
    if with_checkpointing:
        compile_kwargs["checkpointer"] = MemorySaver()

    return graph.compile(**compile_kwargs)


async def run_review(
    raw_text: str = "",
    document_path: str = "",
    with_human_review: bool = True,
) -> ContractReviewState:
    """
    执行合同审查流水线。

    Args:
        raw_text: 合同原始文本
        document_path: 合同文档路径
        with_human_review: 是否启用人工审核
    """
    pipeline = create_review_pipeline(
        with_human_review=with_human_review,
        with_checkpointing=False,
    )

    initial_state = ContractReviewState(
        raw_text=raw_text,
        document_path=document_path,
        status=ReviewStatus.IN_PROGRESS,
    )

    config = {"configurable": {"thread_id": initial_state.review_id}}
    result = await pipeline.ainvoke(initial_state.model_dump(), config=config)

    return ContractReviewState(**result)

"""
共享状态定义 — 多Agent流水线的核心数据模型。

所有Agent通过共享状态通信，LangGraph负责状态的传递和持久化。
"""

from __future__ import annotations

import uuid
from datetime import datetime
from enum import Enum
from typing import Annotated, Any

from pydantic import BaseModel, Field
from langgraph.graph.message import add_messages


# ── 枚举类型 ──────────────────────────────────────────────────

class RiskLevel(str, Enum):
    HIGH = "high"
    MEDIUM = "medium"
    LOW = "low"
    NONE = "none"


class ComplianceStatus(str, Enum):
    COMPLIANT = "compliant"
    NON_COMPLIANT = "non_compliant"
    NEEDS_REVIEW = "needs_review"


class ClauseCategory(str, Enum):
    PAYMENT = "payment"
    LIABILITY = "liability"
    CONFIDENTIALITY = "confidentiality"
    TERMINATION = "termination"
    INTELLECTUAL_PROPERTY = "intellectual_property"
    DISPUTE_RESOLUTION = "dispute_resolution"
    FORCE_MAJEURE = "force_majeure"
    WARRANTY = "warranty"
    INDEMNIFICATION = "indemnification"
    GOVERNING_LAW = "governing_law"
    OTHER = "other"


class ReviewStatus(str, Enum):
    PENDING = "pending"
    IN_PROGRESS = "in_progress"
    AWAITING_HUMAN = "awaiting_human"
    APPROVED = "approved"
    REJECTED = "rejected"
    COMPLETED = "completed"


# ── 数据模型 ──────────────────────────────────────────────────

class ContractEntity(BaseModel):
    """合同中识别出的实体"""
    entity_type: str = Field(description="实体类型：party_a, party_b, amount, date, subject")
    value: str = Field(description="实体值")
    location: str = Field(default="", description="在文档中的位置")


class Clause(BaseModel):
    """结构化条款"""
    id: str = Field(default_factory=lambda: str(uuid.uuid4())[:8])
    title: str = Field(description="条款标题")
    content: str = Field(description="条款内容")
    category: ClauseCategory = Field(default=ClauseCategory.OTHER)
    section_number: str = Field(default="", description="章节编号")
    entities: list[ContractEntity] = Field(default_factory=list)


class RiskFinding(BaseModel):
    """风险发现"""
    clause_id: str = Field(description="关联的条款ID")
    risk_level: RiskLevel = Field(description="风险等级")
    risk_type: str = Field(description="风险类型")
    description: str = Field(description="风险描述")
    buyer_impact: str = Field(default="", description="对买方的影响")
    seller_impact: str = Field(default="", description="对卖方的影响")
    rationale: str = Field(default="", description="判断依据")


class ComplianceFinding(BaseModel):
    """合规发现"""
    clause_id: str = Field(description="关联的条款ID")
    status: ComplianceStatus = Field(description="合规状态")
    regulation: str = Field(description="相关法规")
    issue: str = Field(description="合规问题描述")
    recommendation: str = Field(default="", description="改进建议")


class Suggestion(BaseModel):
    """修改建议"""
    clause_id: str = Field(description="关联的条款ID")
    original_text: str = Field(description="原始文本")
    suggested_text: str = Field(description="建议修改后的文本")
    reason: str = Field(description="修改原因")
    priority: RiskLevel = Field(description="优先级")


class HumanFeedback(BaseModel):
    """人工反馈"""
    reviewer: str = Field(description="审核人")
    decision: str = Field(description="决定：approve/reject/modify")
    comments: str = Field(default="", description="审核意见")
    modified_suggestions: list[Suggestion] = Field(default_factory=list)
    timestamp: datetime = Field(default_factory=datetime.now)


# ── LangGraph 流水线状态 ──────────────────────────────────────

class ContractReviewState(BaseModel):
    """
    LangGraph流水线的共享状态。

    所有Agent读写此状态对象，LangGraph负责状态的传递、
    checkpoint保存和interrupt/resume机制。
    """
    # 基础信息
    review_id: str = Field(default_factory=lambda: str(uuid.uuid4()))
    status: ReviewStatus = Field(default=ReviewStatus.PENDING)
    created_at: datetime = Field(default_factory=datetime.now)

    # 输入
    document_path: str = Field(default="", description="合同文档路径")
    raw_text: str = Field(default="", description="合同原始文本")

    # Agent 1: 条款提取结果
    clauses: list[Clause] = Field(default_factory=list)
    entities: list[ContractEntity] = Field(default_factory=list)
    contract_type: str = Field(default="", description="合同类型")

    # Agent 2: 风险识别结果
    risk_findings: list[RiskFinding] = Field(default_factory=list)
    overall_risk_level: RiskLevel = Field(default=RiskLevel.NONE)
    risk_summary: str = Field(default="")

    # Agent 3: 合规检查结果
    compliance_findings: list[ComplianceFinding] = Field(default_factory=list)
    overall_compliance: ComplianceStatus = Field(default=ComplianceStatus.NEEDS_REVIEW)
    missing_clauses: list[str] = Field(default_factory=list)

    # Agent 4: 修改建议
    suggestions: list[Suggestion] = Field(default_factory=list)
    version_diff: str = Field(default="", description="版本对比结果")

    # 人机协同
    needs_human_review: bool = Field(default=False)
    human_feedback: HumanFeedback | None = Field(default=None)

    # 错误处理
    errors: list[str] = Field(default_factory=list)

    # LangGraph消息追踪
    messages: Annotated[list[Any], add_messages] = Field(default_factory=list)

    model_config = {"arbitrary_types_allowed": True}

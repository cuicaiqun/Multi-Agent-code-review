"""
FastAPI 接口 — 合同审查系统的REST API。

提供以下接口：
- POST /api/v1/review          文本审查
- POST /api/v1/review/upload   文件上传审查
- GET  /api/v1/review/{id}     查询审查状态
- POST /api/v1/review/{id}/feedback  提交人工审核
- GET  /api/v1/health          健康检查
"""

from __future__ import annotations

import logging
import uuid
from datetime import datetime
from pathlib import Path
from typing import Optional

from fastapi import FastAPI, File, HTTPException, UploadFile
from fastapi.middleware.cors import CORSMiddleware
from pydantic import BaseModel, Field

from ..pipeline.graph import create_review_pipeline
from ..pipeline.state import (
    ContractReviewState,
    HumanFeedback,
    ReviewStatus,
    Suggestion,
)

logging.basicConfig(level=logging.INFO)
logger = logging.getLogger(__name__)

app = FastAPI(
    title="多Agent智能合同审查系统",
    description="基于LangGraph的四Agent流水线合同审查API",
    version="1.0.0",
)

app.add_middleware(
    CORSMiddleware,
    allow_origins=["*"],
    allow_credentials=True,
    allow_methods=["*"],
    allow_headers=["*"],
)

# 内存中的审查结果存储（生产环境应使用数据库）
review_store: dict[str, dict] = {}

UPLOAD_DIR = Path("uploads")
UPLOAD_DIR.mkdir(exist_ok=True)


# ── 请求/响应模型 ──────────────────────────────────────────

class ReviewRequest(BaseModel):
    text: str = Field(description="合同文本内容")
    with_human_review: bool = Field(default=True, description="是否启用人工审核")


class ReviewResponse(BaseModel):
    review_id: str
    status: str
    contract_type: str = ""
    clauses_count: int = 0
    risk_summary: str = ""
    overall_risk_level: str = ""
    overall_compliance: str = ""
    suggestions_count: int = 0
    needs_human_review: bool = False
    created_at: str = ""


class FeedbackRequest(BaseModel):
    reviewer: str = Field(description="审核人姓名")
    decision: str = Field(description="approve / reject / modify")
    comments: str = Field(default="", description="审核意见")


class DetailedReviewResponse(BaseModel):
    review_id: str
    status: str
    contract_type: str
    clauses: list[dict]
    risk_findings: list[dict]
    risk_summary: str
    overall_risk_level: str
    compliance_findings: list[dict]
    missing_clauses: list[str]
    overall_compliance: str
    suggestions: list[dict]
    version_diff: str
    needs_human_review: bool
    errors: list[str]


# ── API 接口 ──────────────────────────────────────────────

@app.get("/api/v1/health")
async def health_check():
    return {"status": "healthy", "service": "contract-review", "version": "1.0.0"}


@app.post("/api/v1/review", response_model=ReviewResponse)
async def create_review(request: ReviewRequest):
    """提交合同文本进行审查。"""
    if not request.text.strip():
        raise HTTPException(status_code=400, detail="合同文本不能为空")

    pipeline = create_review_pipeline(
        with_human_review=request.with_human_review,
        with_checkpointing=False,
    )

    review_id = str(uuid.uuid4())
    initial_state = ContractReviewState(
        review_id=review_id,
        raw_text=request.text,
        status=ReviewStatus.IN_PROGRESS,
    )

    try:
        config = {"configurable": {"thread_id": review_id}}
        result = await pipeline.ainvoke(initial_state.model_dump(), config=config)
        state = ContractReviewState(**result)

        review_store[review_id] = state.model_dump()

        return ReviewResponse(
            review_id=review_id,
            status=state.status.value,
            contract_type=state.contract_type,
            clauses_count=len(state.clauses),
            risk_summary=state.risk_summary,
            overall_risk_level=state.overall_risk_level.value,
            overall_compliance=state.overall_compliance.value,
            suggestions_count=len(state.suggestions),
            needs_human_review=state.needs_human_review,
            created_at=state.created_at.isoformat(),
        )
    except Exception as e:
        logger.error("审查失败: %s", str(e))
        raise HTTPException(status_code=500, detail=f"审查处理失败: {str(e)}")


@app.post("/api/v1/review/upload", response_model=ReviewResponse)
async def upload_review(
    file: UploadFile = File(...),
    with_human_review: bool = True,
):
    """上传合同文件进行审查（支持PDF/DOCX/TXT）。"""
    if not file.filename:
        raise HTTPException(status_code=400, detail="文件名不能为空")

    ext = Path(file.filename).suffix.lower()
    if ext not in {".pdf", ".docx", ".doc", ".txt"}:
        raise HTTPException(status_code=400, detail=f"不支持的文件格式: {ext}")

    file_id = str(uuid.uuid4())
    file_path = UPLOAD_DIR / f"{file_id}{ext}"
    content = await file.read()
    file_path.write_bytes(content)

    pipeline = create_review_pipeline(
        with_human_review=with_human_review,
        with_checkpointing=False,
    )

    review_id = str(uuid.uuid4())
    initial_state = ContractReviewState(
        review_id=review_id,
        document_path=str(file_path),
        status=ReviewStatus.IN_PROGRESS,
    )

    try:
        config = {"configurable": {"thread_id": review_id}}
        result = await pipeline.ainvoke(initial_state.model_dump(), config=config)
        state = ContractReviewState(**result)
        review_store[review_id] = state.model_dump()

        return ReviewResponse(
            review_id=review_id,
            status=state.status.value,
            contract_type=state.contract_type,
            clauses_count=len(state.clauses),
            risk_summary=state.risk_summary,
            overall_risk_level=state.overall_risk_level.value,
            overall_compliance=state.overall_compliance.value,
            suggestions_count=len(state.suggestions),
            needs_human_review=state.needs_human_review,
            created_at=state.created_at.isoformat(),
        )
    except Exception as e:
        logger.error("文件审查失败: %s", str(e))
        raise HTTPException(status_code=500, detail=f"审查处理失败: {str(e)}")


@app.get("/api/v1/review/{review_id}", response_model=DetailedReviewResponse)
async def get_review(review_id: str):
    """查询审查详情。"""
    data = review_store.get(review_id)
    if not data:
        raise HTTPException(status_code=404, detail="审查记录不存在")

    state = ContractReviewState(**data)
    return DetailedReviewResponse(
        review_id=state.review_id,
        status=state.status.value,
        contract_type=state.contract_type,
        clauses=[c.model_dump() for c in state.clauses],
        risk_findings=[r.model_dump() for r in state.risk_findings],
        risk_summary=state.risk_summary,
        overall_risk_level=state.overall_risk_level.value,
        compliance_findings=[c.model_dump() for c in state.compliance_findings],
        missing_clauses=state.missing_clauses,
        overall_compliance=state.overall_compliance.value,
        suggestions=[s.model_dump() for s in state.suggestions],
        version_diff=state.version_diff,
        needs_human_review=state.needs_human_review,
        errors=state.errors,
    )


@app.post("/api/v1/review/{review_id}/feedback")
async def submit_feedback(review_id: str, request: FeedbackRequest):
    """提交人工审核反馈。"""
    data = review_store.get(review_id)
    if not data:
        raise HTTPException(status_code=404, detail="审查记录不存在")

    state = ContractReviewState(**data)
    state.human_feedback = HumanFeedback(
        reviewer=request.reviewer,
        decision=request.decision,
        comments=request.comments,
    )

    if request.decision == "approve":
        state.status = ReviewStatus.APPROVED
    elif request.decision == "reject":
        state.status = ReviewStatus.REJECTED
    else:
        state.status = ReviewStatus.COMPLETED

    review_store[review_id] = state.model_dump()
    return {"message": "反馈提交成功", "status": state.status.value}


def start():
    """启动API服务器。"""
    import uvicorn
    from ..config import settings

    uvicorn.run(
        "src.api.main:app",
        host=settings.APP_HOST,
        port=settings.APP_PORT,
        reload=True,
    )


if __name__ == "__main__":
    start()

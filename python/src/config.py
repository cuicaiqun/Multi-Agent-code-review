"""
应用配置 — 通过环境变量加载，绝不硬编码密钥。
"""

import os
from pathlib import Path

from dotenv import load_dotenv

load_dotenv()


class Settings:
    MINIMAX_API_KEY: str = os.getenv("MINIMAX_API_KEY", "")
    MINIMAX_API_BASE: str = os.getenv("MINIMAX_API_BASE", "https://api.minimax.chat/v1")

    OPENAI_API_KEY: str = os.getenv("OPENAI_API_KEY", "")
    OPENAI_API_BASE: str = os.getenv("OPENAI_API_BASE", "https://api.openai.com/v1")

    APP_HOST: str = os.getenv("APP_HOST", "0.0.0.0")
    APP_PORT: int = int(os.getenv("APP_PORT", "8000"))
    LOG_LEVEL: str = os.getenv("LOG_LEVEL", "INFO")

    VECTOR_DB_TYPE: str = os.getenv("VECTOR_DB_TYPE", "faiss")

    PROJECT_ROOT: Path = Path(__file__).parent.parent
    RULES_DIR: Path = PROJECT_ROOT / "src" / "rules" / "definitions"

    # LLM参数 — 合同审查场景需要低温度保证输出稳定性
    LLM_TEMPERATURE: float = 0.1
    LLM_MAX_TOKENS: int = 4096
    LLM_MODEL: str = os.getenv("LLM_MODEL", "MiniMax-M2.7")


settings = Settings()


def get_llm():
    """
    获取LLM客户端。
    优先使用MiniMax（通过OpenAI兼容接口），
    回退到OpenAI。
    """
    from langchain_openai import ChatOpenAI

    if settings.MINIMAX_API_KEY:
        return ChatOpenAI(
            model=settings.LLM_MODEL,
            api_key=settings.MINIMAX_API_KEY,
            base_url=settings.MINIMAX_API_BASE,
            temperature=settings.LLM_TEMPERATURE,
            max_tokens=settings.LLM_MAX_TOKENS,
        )

    if settings.OPENAI_API_KEY:
        return ChatOpenAI(
            model="gpt-4o-mini",
            api_key=settings.OPENAI_API_KEY,
            base_url=settings.OPENAI_API_BASE,
            temperature=settings.LLM_TEMPERATURE,
            max_tokens=settings.LLM_MAX_TOKENS,
        )

    raise ValueError(
        "未配置LLM API Key。请在.env文件中设置 MINIMAX_API_KEY 或 OPENAI_API_KEY"
    )

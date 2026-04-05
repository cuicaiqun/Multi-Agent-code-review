"""法律NLP模块单元测试 — 无需LLM API Key即可运行。"""

from src.nlp.legal_nlp import (
    classify_clause,
    extract_entities,
    text_similarity,
    find_similar_clauses,
    clean_legal_text,
    split_sentences,
)


class TestClauseClassification:
    def test_payment_clause(self):
        category, confidence = classify_clause("甲方应在收到发票后30日内支付货款。")
        assert category == "payment"
        assert confidence > 0

    def test_confidentiality_clause(self):
        category, _ = classify_clause("双方应对商业秘密承担保密义务。")
        assert category == "confidentiality"

    def test_dispute_clause(self):
        category, _ = classify_clause("争议由北京仲裁委员会仲裁解决。")
        assert category == "dispute_resolution"

    def test_unknown_clause(self):
        category, _ = classify_clause("今天天气很好。")
        assert category == "other"


class TestEntityExtraction:
    def test_extract_party_a(self):
        entities = extract_entities("甲方：北京科技有限公司")
        types = [e.label for e in entities]
        assert "party_a" in types

    def test_extract_amount(self):
        entities = extract_entities("合同金额为人民币500,000元整。")
        types = [e.label for e in entities]
        assert "amount" in types

    def test_extract_date(self):
        entities = extract_entities("签订日期：2026年4月1日")
        types = [e.label for e in entities]
        assert "date" in types


class TestTextSimilarity:
    def test_identical_texts(self):
        assert text_similarity("你好世界", "你好世界") == 1.0

    def test_different_texts(self):
        score = text_similarity("合同付款条款", "天气预报信息")
        assert score < 0.5

    def test_similar_texts(self):
        score = text_similarity(
            "甲方应在30日内支付货款",
            "甲方应在30个工作日内支付货款",
        )
        assert score > 0.7

    def test_empty_text(self):
        assert text_similarity("", "任何文本") == 0.0


class TestFindSimilarClauses:
    def test_find_similar(self):
        target = "甲方应支付违约金"
        candidates = ["乙方应支付违约金", "天气很好", "甲方应支付赔偿金"]
        results = find_similar_clauses(target, candidates, threshold=0.5)
        assert len(results) > 0
        assert results[0][0] in [0, 2]


class TestTextPreprocessing:
    def test_clean_text(self):
        result = clean_legal_text("  你好   世界  ")
        assert result == "你好 世界"

    def test_split_sentences(self):
        sentences = split_sentences("第一句话。第二句话！第三句话？")
        assert len(sentences) == 3

import unittest
from unittest.mock import patch

from app.services import embedding_service


class FakeEmbeddingModel:
    def __init__(self):
        self.calls = []

    def encode(self, **kwargs):
        self.calls.append(kwargs)
        dimension = kwargs["truncate_dim"]
        return [[3.0, 4.0] + [0.0] * (dimension - 2) for _ in kwargs["texts"]]


class EmbeddingServiceTests(unittest.TestCase):
    def test_hash_fallback_returns_stable_dimension(self):
        with patch.object(embedding_service, "_embedding_model", side_effect=RuntimeError("missing model")):
            got = embedding_service.generate_embeddings(["first chunk", "second chunk"])

        self.assertTrue(got.success)
        self.assertEqual(got.model, f"hash-dev-{embedding_service.settings.embedding_dimension}")
        self.assertEqual(got.dimension, embedding_service.settings.embedding_dimension)
        self.assertEqual(len(got.embeddings), 2)
        self.assertEqual(len(got.embeddings[0]), embedding_service.settings.embedding_dimension)

    def test_jina_uses_document_and_query_retrieval_prompts(self):
        model = FakeEmbeddingModel()
        with patch.object(embedding_service, "_embedding_model", return_value=model):
            document = embedding_service.generate_embeddings(["A document"], "document")
            query = embedding_service.generate_embeddings(["A query"], "query")

        self.assertTrue(document.success)
        self.assertTrue(query.success)
        self.assertEqual(model.calls[0]["task"], "retrieval")
        self.assertEqual(model.calls[0]["prompt_name"], "document")
        self.assertEqual(model.calls[1]["prompt_name"], "query")
        self.assertEqual(model.calls[0]["truncate_dim"], 256)
        self.assertEqual(document.dimension, 256)
        self.assertAlmostEqual(sum(value * value for value in query.embeddings[0]), 1.0)

    def test_rejects_empty_text(self):
        got = embedding_service.generate_embeddings(["   "])

        self.assertFalse(got.success)
        self.assertEqual(got.error, "embedding text cannot be empty")


if __name__ == "__main__":
    unittest.main()

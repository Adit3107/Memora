import unittest
from unittest.mock import patch

from app.services import embedding_service


class EmbeddingServiceTests(unittest.TestCase):
    def test_hash_fallback_returns_stable_dimension(self):
        with patch.object(embedding_service, "_sentence_transformer", side_effect=RuntimeError("missing model")):
            got = embedding_service.generate_embeddings(["first chunk", "second chunk"])

        self.assertTrue(got.success)
        self.assertEqual(got.model, "hash-dev-384")
        self.assertEqual(got.dimension, 384)
        self.assertEqual(len(got.embeddings), 2)
        self.assertEqual(len(got.embeddings[0]), 384)

    def test_rejects_empty_text(self):
        got = embedding_service.generate_embeddings(["   "])

        self.assertFalse(got.success)
        self.assertEqual(got.error, "embedding text cannot be empty")


if __name__ == "__main__":
    unittest.main()

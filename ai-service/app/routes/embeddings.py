from app.models.embeddings import EmbeddingRequest, EmbeddingResponse
from app.services.embedding_service import generate_embeddings
from fastapi import APIRouter

router = APIRouter(prefix="/embeddings")


@router.post("", response_model=EmbeddingResponse)
def embeddings(request: EmbeddingRequest) -> EmbeddingResponse:
    return generate_embeddings(request.texts)


# Why this file exists:
# Go calls this internal route with chunk text and receives vectors. It does not
# persist anything and it is not intended for direct frontend use.

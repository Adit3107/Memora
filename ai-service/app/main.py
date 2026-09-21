from fastapi import FastAPI

from app.routes.extraction import router as extraction_router
from app.routes.health import router as health_router

app = FastAPI(title="Memora Extraction Service")

app.include_router(health_router)
app.include_router(extraction_router)


# Why this file exists:
# FastAPI apps start from one application object. Go calls this internal service
# only for extraction work that is better handled by Python libraries.

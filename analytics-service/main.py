import os
import logging
from contextlib import asynccontextmanager

from fastapi import FastAPI
from fastapi.middleware.cors import CORSMiddleware

from database import init_database, close_database
from routes import analytics, health

# Configure logging
logging.basicConfig(level=logging.INFO)
logger = logging.getLogger(__name__)


@asynccontextmanager
async def lifespan(app: FastAPI):
    # Startup
    try:
        init_database()
    except Exception as e:
        logger.error(f"Failed to initialize database: {e}")
        logger.warning("Service starting without database connection. Database operations will fail until connection is established.")

    yield

    # Shutdown
    close_database()


app = FastAPI(title="Analytics Service", lifespan=lifespan)

# CORS middleware
app.add_middleware(
    CORSMiddleware,
    allow_origins=["*"],
    allow_credentials=True,
    allow_methods=["*"],
    allow_headers=["*"],
)

# Include routers
app.include_router(health.router)
app.include_router(analytics.router)


if __name__ == "__main__":
    import uvicorn
    port = int(os.getenv("PORT", "7544"))
    uvicorn.run(app, host="0.0.0.0", port=port)

from fastapi import APIRouter
from services import check_database_health

router = APIRouter()


@router.get("/health")
def health_check():
    """Health check endpoint"""
    db_healthy = check_database_health()
    status = "healthy" if db_healthy else "unhealthy"
    status_code = 200 if status == "healthy" else 503

    return {
        "status": status,
        "database": db_healthy
    }

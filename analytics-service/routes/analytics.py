import logging
from typing import List, Dict, Any
from fastapi import APIRouter, HTTPException, Query
from models import AnalyticsSummary, URLStats, TimeSeriesData
from auth import validate_api_key
from services import (
    get_analytics_summary_data,
    get_top_urls_data,
    get_time_series_data,
    get_url_analytics_data
)

logger = logging.getLogger(__name__)
router = APIRouter()


@router.get("/api/analytics/summary", response_model=AnalyticsSummary)
def get_analytics_summary(api_key: str = Query(...)):
    """Get overall analytics summary"""
    user_id = validate_api_key(api_key)
    if not user_id:
        raise HTTPException(status_code=401, detail="Invalid API key")

    try:
        summary = get_analytics_summary_data(user_id)
        return summary
    except Exception as e:
        logger.error(f"Database error in get_analytics_summary: {e}")
        raise HTTPException(status_code=500, detail="Failed to fetch analytics")


@router.get("/api/analytics/top-urls", response_model=List[URLStats])
def get_top_urls(api_key: str = Query(...), limit: int = Query(10, ge=1, le=100)):
    """Get top performing URLs by click count"""
    user_id = validate_api_key(api_key)
    if not user_id:
        raise HTTPException(status_code=401, detail="Invalid API key")

    try:
        url_stats = get_top_urls_data(user_id, limit)
        return url_stats
    except Exception as e:
        logger.error(f"Database error in get_top_urls: {e}")
        raise HTTPException(status_code=500, detail="Failed to fetch top URLs")


@router.get("/api/analytics/time-series", response_model=List[TimeSeriesData])
def get_time_series(
    api_key: str = Query(...),
    days: int = Query(7, ge=1, le=90)
):
    """Get clicks over time (time series data)"""
    user_id = validate_api_key(api_key)
    if not user_id:
        raise HTTPException(status_code=401, detail="Invalid API key")

    try:
        time_series = get_time_series_data(user_id, days)
        return time_series
    except Exception as e:
        logger.error(f"Database error in get_time_series: {e}")
        raise HTTPException(status_code=500, detail="Failed to fetch time series data")


@router.get("/api/analytics/url/{url_id}", response_model=Dict[str, Any])
def get_url_analytics(url_id: int, api_key: str = Query(...)):
    """Get detailed analytics for a specific URL"""
    user_id = validate_api_key(api_key)
    if not user_id:
        raise HTTPException(status_code=401, detail="Invalid API key")

    try:
        url_data = get_url_analytics_data(url_id, user_id)
        if not url_data:
            raise HTTPException(status_code=404, detail="URL not found")
        return url_data
    except HTTPException:
        raise
    except Exception as e:
        logger.error(f"Database error in get_url_analytics: {e}")
        raise HTTPException(status_code=500, detail="Failed to fetch URL analytics")

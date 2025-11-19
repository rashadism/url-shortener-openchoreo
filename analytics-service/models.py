from datetime import datetime
from typing import Optional
from pydantic import BaseModel


class AnalyticsSummary(BaseModel):
    total_urls: int
    total_clicks: int
    clicks_today: int
    clicks_this_week: int


class URLStats(BaseModel):
    url_id: int
    short_code: str
    long_url: str
    title: Optional[str]
    total_clicks: int
    created_at: datetime


class ClickEvent(BaseModel):
    id: int
    url_id: int
    ip_address: str
    user_agent: Optional[str]
    referer: Optional[str]
    country: Optional[str]
    city: Optional[str]
    clicked_at: datetime


class TimeSeriesData(BaseModel):
    date: str
    clicks: int

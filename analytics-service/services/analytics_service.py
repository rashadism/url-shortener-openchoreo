import logging
from typing import List, Dict, Any
from psycopg2.extras import RealDictCursor
from database import get_cursor

logger = logging.getLogger(__name__)


def get_analytics_summary_data(user_id: int) -> Dict[str, int]:
    """Get overall analytics summary for a user"""
    try:
        with get_cursor(cursor_factory=RealDictCursor) as cursor:
            # Total URLs
            cursor.execute(
                "SELECT COUNT(*) as count FROM urls WHERE user_id = %s",
                (user_id,)
            )
            total_urls = cursor.fetchone()['count']

            # Total clicks
            cursor.execute("""
                SELECT COUNT(*) as count
                FROM clicks c
                JOIN urls u ON c.url_id = u.id
                WHERE u.user_id = %s
            """, (user_id,))
            total_clicks = cursor.fetchone()['count']

            # Clicks today
            cursor.execute("""
                SELECT COUNT(*) as count
                FROM clicks c
                JOIN urls u ON c.url_id = u.id
                WHERE u.user_id = %s
                AND DATE(c.clicked_at) = CURRENT_DATE
            """, (user_id,))
            clicks_today = cursor.fetchone()['count']

            # Clicks this week
            cursor.execute("""
                SELECT COUNT(*) as count
                FROM clicks c
                JOIN urls u ON c.url_id = u.id
                WHERE u.user_id = %s
                AND c.clicked_at >= CURRENT_DATE - INTERVAL '7 days'
            """, (user_id,))
            clicks_this_week = cursor.fetchone()['count']

        return {
            "total_urls": total_urls,
            "total_clicks": total_clicks,
            "clicks_today": clicks_today,
            "clicks_this_week": clicks_this_week
        }
    except Exception as e:
        raise Exception(f"Failed to get analytics summary for user {user_id}: {e}") from e


def get_top_urls_data(user_id: int, limit: int) -> List[Dict[str, Any]]:
    """Get top performing URLs by click count"""
    try:
        with get_cursor(cursor_factory=RealDictCursor) as cursor:
            # Complex query with joins and aggregation
            cursor.execute("""
                SELECT
                    u.id as url_id,
                    u.short_code,
                    u.long_url,
                    u.title,
                    u.created_at,
                    COUNT(c.id) as total_clicks
                FROM urls u
                LEFT JOIN clicks c ON u.id = c.url_id
                WHERE u.user_id = %s
                GROUP BY u.id, u.short_code, u.long_url, u.title, u.created_at
                ORDER BY total_clicks DESC
                LIMIT %s
            """, (user_id, limit))

            results = cursor.fetchall()

        url_stats = [
            {
                "url_id": row['url_id'],
                "short_code": row['short_code'],
                "long_url": row['long_url'],
                "title": row['title'],
                "total_clicks": row['total_clicks'],
                "created_at": row['created_at'].isoformat()
            }
            for row in results
        ]

        return url_stats
    except Exception as e:
        raise Exception(f"Failed to get top URLs for user {user_id}: {e}") from e


def get_time_series_data(user_id: int, days: int) -> List[Dict[str, Any]]:
    """Get clicks over time (time series data)"""
    try:
        with get_cursor(cursor_factory=RealDictCursor) as cursor:
            # Complex time-series aggregation query
            cursor.execute("""
                SELECT
                    DATE(c.clicked_at) as date,
                    COUNT(*) as clicks
                FROM clicks c
                JOIN urls u ON c.url_id = u.id
                WHERE u.user_id = %s
                AND c.clicked_at >= CURRENT_DATE - INTERVAL '%s days'
                GROUP BY DATE(c.clicked_at)
                ORDER BY date ASC
            """, (user_id, days))

            results = cursor.fetchall()

        time_series = [
            {
                "date": row['date'].isoformat(),
                "clicks": row['clicks']
            }
            for row in results
        ]

        return time_series
    except Exception as e:
        raise Exception(f"Failed to get time series data for user {user_id} (last {days} days): {e}") from e


def get_url_analytics_data(url_id: int, user_id: int) -> Dict[str, Any]:
    """Get detailed analytics for a specific URL"""
    try:
        with get_cursor(cursor_factory=RealDictCursor) as cursor:
            # Verify URL belongs to user
            cursor.execute(
                "SELECT * FROM urls WHERE id = %s AND user_id = %s",
                (url_id, user_id)
            )
            url = cursor.fetchone()

            if not url:
                return None

            # Get click count
            cursor.execute(
                "SELECT COUNT(*) as count FROM clicks WHERE url_id = %s",
                (url_id,)
            )
            click_count = cursor.fetchone()['count']

            # Get recent clicks
            cursor.execute("""
                SELECT * FROM clicks
                WHERE url_id = %s
                ORDER BY clicked_at DESC
                LIMIT 50
            """, (url_id,))
            recent_clicks = cursor.fetchall()

        return {
            "url": dict(url),
            "total_clicks": click_count,
            "recent_clicks": [dict(click) for click in recent_clicks]
        }
    except Exception as e:
        raise Exception(f"Failed to get analytics for URL {url_id} (user {user_id}): {e}") from e


def check_database_health() -> bool:
    """Check database connection health"""
    try:
        with get_cursor() as cursor:
            cursor.execute("SELECT 1")
        return True
    except Exception as e:
        logger.error(f"Database health check failed: {e}")
        return False

import logging
from typing import Optional
from psycopg2.extras import RealDictCursor
from database import get_cursor

logger = logging.getLogger(__name__)


def validate_api_key(api_key: str) -> Optional[int]:
    """Validate API key and return user ID"""
    try:
        with get_cursor(cursor_factory=RealDictCursor) as cursor:
            cursor.execute(
                "SELECT id FROM users WHERE api_key = %s",
                (api_key,)
            )
            result = cursor.fetchone()
            return result['id'] if result else None
    except Exception as e:
        logger.error(f"Error validating API key: {e}")
        return None

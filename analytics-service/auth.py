import logging
from typing import Optional
from psycopg2.extras import RealDictCursor
from database import get_cursor

logger = logging.getLogger(__name__)


def get_or_create_user(username: str) -> Optional[int]:
    """Get or create user by username and return user ID"""
    try:
        with get_cursor(cursor_factory=RealDictCursor) as cursor:
            # Try to get existing user
            cursor.execute(
                "SELECT id FROM users WHERE username = %s",
                (username,)
            )
            result = cursor.fetchone()
            if result:
                return result['id']

            # User doesn't exist, create new user
            cursor.execute(
                "INSERT INTO users (username, api_key) VALUES (%s, %s) RETURNING id",
                (username, "")
            )
            result = cursor.fetchone()
            return result['id'] if result else None
    except Exception as e:
        logger.error(f"Error getting/creating user: {e}")
        return None

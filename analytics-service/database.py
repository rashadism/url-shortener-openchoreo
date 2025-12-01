import os
import logging
import psycopg2
from contextlib import contextmanager

logger = logging.getLogger(__name__)

# Global database connection
db_conn = None


def init_database():
    """Initialize database connection"""
    global db_conn

    db_url = os.getenv("DATABASE_URL", "postgresql://urlshortener:password123@localhost:5432/urlshortener")
    db_conn = psycopg2.connect(db_url)
    logger.info("Database connected successfully")
    return db_conn


def close_database():
    """Close database connection"""
    global db_conn
    if db_conn:
        db_conn.close()
        logger.info("Database connection closed")


def get_db():
    """Get database connection"""
    return db_conn


@contextmanager
def get_cursor(cursor_factory=None):
    """Context manager for database cursor"""
    if db_conn is None:
        raise Exception("Database connection not available")
    cursor = db_conn.cursor(cursor_factory=cursor_factory)
    try:
        yield cursor
        db_conn.commit()
    except Exception as e:
        db_conn.rollback()
        raise e
    finally:
        cursor.close()

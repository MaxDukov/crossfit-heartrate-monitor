import os
from sqlalchemy import create_engine, text
from sqlalchemy.orm import sessionmaker, DeclarativeBase

DB_PATH = os.environ.get("CF_DB_PATH", "/tmp/cf_monitor.db")
DATABASE_URL = f"sqlite:///{DB_PATH}"

engine = create_engine(DATABASE_URL, connect_args={"check_same_thread": False})
SessionLocal = sessionmaker(autocommit=False, autoflush=False, bind=engine)


class Base(DeclarativeBase):
    pass


def get_db():
    """FastAPI dependency: выдаёт сессию БД и гарантирует закрытие."""
    db = SessionLocal()
    try:
        yield db
    finally:
        db.close()


def _column_exists(conn, table: str, column: str) -> bool:
    rows = conn.execute(text(f"PRAGMA table_info({table})")).fetchall()
    return any(r[1] == column for r in rows)


def _run_migrations():
    """Добавляет недостающие колонки в существующие таблицы (SQLite ALTER TABLE)."""
    with engine.connect() as conn:
        if not _column_exists(conn, "sensors", "ignored"):
            conn.execute(text(
                "ALTER TABLE sensors ADD COLUMN ignored BOOLEAN DEFAULT 0 NOT NULL"
            ))
            conn.commit()

        if not _column_exists(conn, "athletes", "weight_kg"):
            conn.execute(text(
                "ALTER TABLE athletes ADD COLUMN weight_kg FLOAT"
            ))
            conn.commit()

        if not _column_exists(conn, "athletes", "age"):
            conn.execute(text(
                "ALTER TABLE athletes ADD COLUMN age INTEGER"
            ))
            conn.commit()


def init_db():
    """Создаёт таблицы при первом запуске и применяет миграции."""
    Base.metadata.create_all(bind=engine)
    _run_migrations()

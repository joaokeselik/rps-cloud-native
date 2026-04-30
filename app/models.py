from datetime import datetime, timezone

from sqlalchemy import Column, DateTime, Integer, String

from app.database import Base


class GameRound(Base):
    __tablename__ = "game_rounds"

    id = Column(Integer, primary_key=True, index=True)
    player_id = Column(Integer, nullable=True)
    player_name = Column(String(120), nullable=True)
    player_choice = Column(String(20), nullable=False)
    computer_choice = Column(String(20), nullable=False)
    outcome = Column(String(20), nullable=False)
    created_at = Column(
        DateTime(timezone=True),
        default=lambda: datetime.now(timezone.utc),
        nullable=False,
    )

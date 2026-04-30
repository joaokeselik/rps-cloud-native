from datetime import datetime
from typing import Literal

from pydantic import BaseModel, ConfigDict, Field


Choice = Literal["rock", "paper", "scissors"]


class GameCreate(BaseModel):
    player_choice: Choice
    player_id: int | None = None
    player_name: str | None = Field(default=None, max_length=120)


class GameRoundResponse(BaseModel):
    id: int
    player_id: int | None
    player_name: str | None
    player_choice: Choice
    computer_choice: Choice
    outcome: Literal["win", "loss", "draw"]
    created_at: datetime

    model_config = ConfigDict(from_attributes=True)


class StatsResponse(BaseModel):
    total_games: int
    wins: int
    losses: int
    draws: int
    win_rate: float

from datetime import datetime
from typing import Literal

from pydantic import BaseModel, ConfigDict


Choice = Literal["rock", "paper", "scissors"]


class GameCreate(BaseModel):
    player_choice: Choice


class GameRoundResponse(BaseModel):
    id: int
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


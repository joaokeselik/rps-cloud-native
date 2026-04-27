import os
import time
from contextlib import asynccontextmanager
from pathlib import Path

from fastapi import Depends, FastAPI, Request
from fastapi.responses import HTMLResponse
from fastapi.templating import Jinja2Templates
from sqlalchemy import select, func
from sqlalchemy.exc import OperationalError
from sqlalchemy.orm import Session

from app.database import Base, engine, get_db
from app.game import determine_outcome, pick_computer_choice
from app.models import GameRound
from app.schemas import GameCreate, GameRoundResponse, StatsResponse


BASE_DIR = Path(__file__).resolve().parent
templates = Jinja2Templates(directory=str(BASE_DIR / "templates"))


def initialize_database():
    retries = int(os.getenv("DB_STARTUP_RETRIES", "10"))
    delay = int(os.getenv("DB_STARTUP_DELAY", "3"))
    last_error = None

    for _ in range(retries):
        try:
            Base.metadata.create_all(bind=engine)
            return
        except OperationalError as error:
            last_error = error
            time.sleep(delay)

    if last_error is not None:
        raise last_error


@asynccontextmanager
async def lifespan(app: FastAPI):
    initialize_database()
    yield


app = FastAPI(title="Rock Paper Scissors API", lifespan=lifespan)


@app.get("/", response_class=HTMLResponse)
def read_index(request: Request):
    return templates.TemplateResponse(request=request, name="index.html")


@app.get("/healthz")
def healthcheck():
    return {"status": "ok"}


@app.post("/api/games", response_model=GameRoundResponse)
def create_game_round(payload: GameCreate, db: Session = Depends(get_db)):
    computer_choice = pick_computer_choice()
    outcome = determine_outcome(payload.player_choice, computer_choice)

    game_round = GameRound(
        player_choice=payload.player_choice,
        computer_choice=computer_choice,
        outcome=outcome,
    )
    db.add(game_round)
    db.commit()
    db.refresh(game_round)

    return game_round


@app.get("/api/games", response_model=list[GameRoundResponse])
def list_game_rounds(limit: int = 10, db: Session = Depends(get_db)):
    query = select(GameRound).order_by(GameRound.id.desc()).limit(limit)
    return db.scalars(query).all()


@app.get("/api/stats", response_model=StatsResponse)
def get_stats(db: Session = Depends(get_db)):
    total_games = db.scalar(select(func.count(GameRound.id))) or 0
    wins = db.scalar(
        select(func.count(GameRound.id)).where(GameRound.outcome == "win")
    ) or 0
    losses = db.scalar(
        select(func.count(GameRound.id)).where(GameRound.outcome == "loss")
    ) or 0
    draws = db.scalar(
        select(func.count(GameRound.id)).where(GameRound.outcome == "draw")
    ) or 0
    win_rate = round((wins / total_games) * 100, 1) if total_games else 0.0

    return StatsResponse(
        total_games=total_games,
        wins=wins,
        losses=losses,
        draws=draws,
        win_rate=win_rate,
    )

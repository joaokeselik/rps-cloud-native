import pytest
from fastapi.testclient import TestClient
from sqlalchemy import create_engine
from sqlalchemy.orm import sessionmaker
from sqlalchemy.pool import StaticPool

import app.main as main
from app.database import Base, get_db
from app.main import app


engine = create_engine(
    "sqlite://",
    connect_args={"check_same_thread": False},
    poolclass=StaticPool,
)
TestingSessionLocal = sessionmaker(autocommit=False, autoflush=False, bind=engine)


def override_get_db():
    db = TestingSessionLocal()
    try:
        yield db
    finally:
        db.close()


app.dependency_overrides[get_db] = override_get_db
client = TestClient(app)


@pytest.fixture(autouse=True)
def reset_database():
    Base.metadata.drop_all(bind=engine)
    Base.metadata.create_all(bind=engine)
    yield


def test_create_game_round_saves_result(monkeypatch):
    monkeypatch.setattr(main, "pick_computer_choice", lambda: "scissors")

    response = client.post("/api/games", json={"player_choice": "rock"})

    assert response.status_code == 200
    payload = response.json()
    assert payload["player_choice"] == "rock"
    assert payload["computer_choice"] == "scissors"
    assert payload["outcome"] == "win"
    assert payload["player_id"] is None
    assert payload["player_name"] is None


def test_create_game_round_can_attach_player(monkeypatch):
    monkeypatch.setattr(main, "pick_computer_choice", lambda: "paper")

    response = client.post(
        "/api/games",
        json={
            "player_choice": "scissors",
            "player_id": 7,
            "player_name": "Ada Lovelace",
        },
    )

    assert response.status_code == 200
    payload = response.json()
    assert payload["player_id"] == 7
    assert payload["player_name"] == "Ada Lovelace"
    assert payload["outcome"] == "win"


def test_index_page_renders():
    response = client.get("/")

    assert response.status_code == 200
    assert "Rock, Paper, Scissors" in response.text


def test_stats_endpoint_returns_aggregated_values(monkeypatch):
    choices = iter(["scissors", "rock", "paper"])
    monkeypatch.setattr(main, "pick_computer_choice", lambda: next(choices))

    client.post("/api/games", json={"player_choice": "rock"})
    client.post("/api/games", json={"player_choice": "scissors"})
    client.post("/api/games", json={"player_choice": "paper"})

    response = client.get("/api/stats")

    assert response.status_code == 200
    payload = response.json()
    assert payload == {
        "total_games": 3,
        "wins": 1,
        "losses": 1,
        "draws": 1,
        "win_rate": 33.3,
    }


def test_stats_and_games_can_filter_by_player(monkeypatch):
    choices = iter(["scissors", "paper", "scissors"])
    monkeypatch.setattr(main, "pick_computer_choice", lambda: next(choices))

    client.post(
        "/api/games",
        json={"player_choice": "rock", "player_id": 1, "player_name": "Ada"},
    )
    client.post(
        "/api/games",
        json={"player_choice": "rock", "player_id": 2, "player_name": "Grace"},
    )
    client.post(
        "/api/games",
        json={"player_choice": "rock", "player_id": 1, "player_name": "Ada"},
    )

    stats_response = client.get("/api/stats?player_id=1")
    games_response = client.get("/api/games?player_id=1")

    assert stats_response.status_code == 200
    assert stats_response.json() == {
        "total_games": 2,
        "wins": 2,
        "losses": 0,
        "draws": 0,
        "win_rate": 100.0,
    }
    assert games_response.status_code == 200
    assert [game["player_id"] for game in games_response.json()] == [1, 1]

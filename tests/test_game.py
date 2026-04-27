import pytest

from app.game import determine_outcome


@pytest.mark.parametrize(
    ("player_choice", "computer_choice", "expected"),
    [
        ("rock", "scissors", "win"),
        ("rock", "paper", "loss"),
        ("paper", "rock", "win"),
        ("paper", "scissors", "loss"),
        ("scissors", "paper", "win"),
        ("scissors", "rock", "loss"),
        ("rock", "rock", "draw"),
    ],
)
def test_determine_outcome(player_choice, computer_choice, expected):
    assert determine_outcome(player_choice, computer_choice) == expected


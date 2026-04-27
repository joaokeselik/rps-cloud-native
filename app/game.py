import random


CHOICES = ("rock", "paper", "scissors")
WIN_RULES = {
    "rock": "scissors",
    "paper": "rock",
    "scissors": "paper",
}


def pick_computer_choice():
    return random.choice(CHOICES)


def determine_outcome(player_choice, computer_choice):
    if player_choice == computer_choice:
        return "draw"

    if WIN_RULES[player_choice] == computer_choice:
        return "win"

    return "loss"


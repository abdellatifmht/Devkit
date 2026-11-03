"""Chargement du jeu de données Iris."""
from __future__ import annotations

from sklearn.datasets import load_iris
from sklearn.model_selection import train_test_split


def load_dataset() -> tuple[list[list[float]], list[list[float]], list[int], list[int]]:
    iris = load_iris()
    return train_test_split(
        iris.data.tolist(),
        iris.target.tolist(),
        test_size=0.2,
        random_state=42,
        stratify=iris.target,
    )

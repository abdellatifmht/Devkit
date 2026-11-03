"""Point d'entrée CLI de l'agent."""
import argparse

from .agent import ask


def main() -> None:
    parser = argparse.ArgumentParser(description=__doc__)
    parser.add_argument(
        "question",
        nargs="?",
        default="Calcule 2 * (3 + 4) puis compte les mots de 'bonjour le monde'.",
    )
    parser.add_argument("--thread", default="default", help="Identifiant de conversation")
    args = parser.parse_args()
    print(ask(args.question, args.thread))


if __name__ == "__main__":
    main()

"""Pose une question au RAG : recherche des extraits puis génération."""
import argparse

from .config import settings
from .llm import complete
from .vectorstore import VectorStore

PROMPT = """Réponds à la question en t'appuyant uniquement sur le contexte.
Si l'information n'y figure pas, dis-le clairement.

Contexte :
{context}

Question : {question}
"""


def answer(question: str) -> str:
    passages = VectorStore().search(question, settings.top_k)
    context = "\n\n---\n\n".join(passages)
    return complete(PROMPT.format(context=context, question=question))


def main() -> None:
    parser = argparse.ArgumentParser(description=__doc__)
    parser.add_argument("question")
    print(answer(parser.parse_args().question))


if __name__ == "__main__":
    main()

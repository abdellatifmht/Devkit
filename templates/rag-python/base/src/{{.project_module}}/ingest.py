"""Indexe les fichiers .md / .txt d'un dossier dans la base vectorielle."""
import argparse
from pathlib import Path

from .vectorstore import VectorStore


def chunk(text: str, size: int = 800) -> list[str]:
    """Découpe par paragraphes en blocs d'environ `size` caractères."""
    chunks, current = [], ""
    for para in text.split("\n\n"):
        if current and len(current) + len(para) > size:
            chunks.append(current.strip())
            current = ""
        current += para + "\n\n"
    if current.strip():
        chunks.append(current.strip())
    return chunks


def main() -> None:
    parser = argparse.ArgumentParser(description=__doc__)
    parser.add_argument("folder", nargs="?", default="data")
    args = parser.parse_args()

    store = VectorStore()
    total = 0
    for path in sorted(Path(args.folder).rglob("*")):
        if path.suffix.lower() in {".md", ".txt"}:
            chunks = chunk(path.read_text(encoding="utf-8"))
            if chunks:
                store.add(chunks)
                total += len(chunks)
    print(f"{total} extraits indexés")


if __name__ == "__main__":
    main()

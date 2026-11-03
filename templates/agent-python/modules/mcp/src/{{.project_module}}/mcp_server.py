"""Serveur MCP d'exemple, exposé en stdio."""
from fastmcp import FastMCP

mcp = FastMCP("devkit-agent")


@mcp.tool
def word_count(text: str) -> int:
    """Compte les mots d'un texte."""
    return len(text.split())


if __name__ == "__main__":
    mcp.run()

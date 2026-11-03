"""Outils exposés à l'agent."""
from __future__ import annotations

import ast
import operator

from langchain_core.tools import tool

_BIN_OPS = {
    ast.Add: operator.add,
    ast.Sub: operator.sub,
    ast.Mult: operator.mul,
    ast.Div: operator.truediv,
    ast.Pow: operator.pow,
    ast.Mod: operator.mod,
    ast.USub: operator.neg,
    ast.UAdd: operator.pos,
}


def _eval(node: ast.AST) -> float:
    if isinstance(node, ast.Expression):
        return _eval(node.body)
    if isinstance(node, ast.Constant) and isinstance(node.value, (int, float)):
        return float(node.value)
    if isinstance(node, ast.BinOp) and type(node.op) in _BIN_OPS:
        return _BIN_OPS[type(node.op)](_eval(node.left), _eval(node.right))
    if isinstance(node, ast.UnaryOp) and type(node.op) in _BIN_OPS:
        return _BIN_OPS[type(node.op)](_eval(node.operand))
    raise ValueError("expression non supportée")


@tool
def calculator(expression: str) -> str:
    """Évalue une expression arithmétique simple, par exemple ``2 * (3 + 4)``."""
    try:
        value = _eval(ast.parse(expression, mode="eval"))
    except (SyntaxError, ValueError, ZeroDivisionError) as exc:
        return f"Erreur: {exc}"
    return str(value)


@tool
def word_count(text: str) -> str:
    """Compte les mots d'un texte."""
    return str(len(text.split()))


TOOLS = [calculator, word_count]

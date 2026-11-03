import json
import sys
import xml.etree.ElementTree as ET
from pathlib import Path

import yaml

roots = [Path(p) for p in sys.argv[1:]]
failed = False

for root in roots:
    for path in root.rglob("*"):
        if not path.is_file():
            continue
        rel = path.relative_to(root)
        try:
            if path.name == "package.json":
                json.loads(path.read_text())
            elif path.suffix in {".yaml", ".yml"}:
                yaml.safe_load(path.read_text())
            elif path.suffix == ".xml":
                ET.parse(path)
        except Exception as exc:
            failed = True
            print(f"INVALID {rel}: {exc}")
        text = path.read_text(errors="replace")
        if "{{" in text:
            failed = True
            print(f"UNRENDERED {rel}")
        if "<no value>" in text:
            failed = True
            print(f"NO VALUE {rel}")

if failed:
    sys.exit(1)
print("all parsed files are valid")

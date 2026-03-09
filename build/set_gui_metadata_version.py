#!/usr/bin/env python3

import json
import os
from pathlib import Path


def main() -> None:
    version = os.environ["GITHUB_REF_NAME"].lstrip("v")
    root_dir = Path(__file__).resolve().parents[1]
    wails_config_path = root_dir / "src" / "gui" / "wails.json"

    data = json.loads(wails_config_path.read_text())
    data.setdefault("info", {})["productVersion"] = version
    wails_config_path.write_text(json.dumps(data, indent=2) + "\n")


if __name__ == "__main__":
    main()

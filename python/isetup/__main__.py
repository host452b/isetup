"""Entry point: hand off to the bundled Go binary.

The wheel for each platform bundles ``isetup/bin/isetup`` (or ``isetup.exe`` on
Windows). On POSIX we replace the Python process with the binary via os.execv
so signals and exit codes propagate cleanly. On Windows execv doesn't replace
the process tree well, so we use subprocess and forward the return code.
"""

from __future__ import annotations

import os
import subprocess
import sys
from pathlib import Path


def _binary_path() -> Path:
    name = "isetup.exe" if sys.platform == "win32" else "isetup"
    return Path(__file__).resolve().parent / "bin" / name


def main() -> None:
    binary = _binary_path()
    if not binary.exists():
        sys.stderr.write(
            f"isetup: bundled binary not found at {binary}. "
            "This usually means the wrong platform wheel was installed; "
            "try: pip install --force-reinstall --no-cache-dir isetup\n"
        )
        sys.exit(1)

    # Ensure executable bit is set (pip preserves it, but be defensive).
    if sys.platform != "win32":
        try:
            mode = os.stat(binary).st_mode
            if not mode & 0o111:
                os.chmod(binary, mode | 0o755)
        except OSError:
            pass

    argv = [str(binary), *sys.argv[1:]]

    if sys.platform == "win32":
        try:
            proc = subprocess.run(argv, check=False)
        except KeyboardInterrupt:
            sys.exit(130)
        sys.exit(proc.returncode)

    os.execv(str(binary), argv)


if __name__ == "__main__":
    main()

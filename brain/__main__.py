"""Allow running the brain package as a module: python -m brain"""

from brain.main import main
import sys

if __name__ == "__main__":
    sys.exit(main())

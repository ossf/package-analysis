#!/usr/bin/env python3
from dataclasses import dataclass
import importlib
from importlib.metadata import files
import sys
import subprocess
import traceback
from typing import Optional

PY_EXTENSION = '.py'

@dataclass
class Package:
    """Class for tracking a package."""
    name: str
    version: Optional[str] = None
    local_path: Optional[str] = None

    def install_arg(self) -> str:
        if self.local_path:
            return self.local_path
        elif self.version:
            return f'{self.name}=={self.version}'
        else:
            return self.name

def install(package):
    """Pip install."""
    arg = package.install_arg()
    try:
      output = subprocess.check_output(
          (sys.executable, '-m', 'pip', 'install', '--pre', arg),
          stderr=subprocess.STDOUT)
      print('Install succeeded:')
      print(output.decode())
    except subprocess.CalledProcessError as e:
      print('Failed to install:')
      print(e.output.decode())
      # Always raise.
      # Install failing is either an interesting issue, or an opportunity to
      # improve the analysis.
      raise

def path_to_import(path):
    """Convert a path to import."""
    if path.name == '__init__.py':
        import_path = str(path.parent)
    else:
        import_path = str(path)[:-len(PY_EXTENSION)]

    return import_path.replace('/', '.')


def importPkg(package):
    """Import phase for analyzing the package."""
    for path in files(package.name):
        # TODO: pyc, C extensions?
        if path.suffix != PY_EXTENSION:
            continue

        import_path = path_to_import(path)
        print('Importing', import_path)
        try:
            importlib.import_module(import_path)
        except:
            print('Failed to import', import_path)
            traceback.print_exc()

PHASES = {
    "all": [install, importPkg],
    "install": [install],
    "import": [importPkg]
}

def main():
    args = list(sys.argv)
    script = args.pop(0)

    if len(args) < 2 or len(args) > 4:
        raise ValueError(f'Usage: {script} [--local file | --version version] phase package_name')

    # Parse the arguments manually to avoid introducing unnecessary dependencies
    # and side effects that add noise to the strace output.
    local_path = None
    version = None
    if args[0] == '--local':
        args.pop(0)
        local_path = args.pop(0)
    elif args[0] == '--version':
        args.pop(0)
        version = args.pop(0)

    phase = args.pop(0)
    package_name = args.pop(0)

    if not phase in PHASES:
        print(f'Unknown phase {phase} specified.')
        exit(1)

    package = Package(name=package_name, version=version, local_path=local_path)

    # Execute for the specified phase.
    for phase in PHASES[phase]:
        phase(package)


if __name__ == '__main__':
    main()

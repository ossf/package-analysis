#!/usr/bin/env python3
import importlib
from importlib.metadata import files
import sys
import subprocess
import traceback

PY_EXTENSION = '.py'


def pip_install(package):
    """Pip install."""
    try:
      output = subprocess.check_output(
          (sys.executable, '-m', 'pip', 'install', package),
          stderr=subprocess.STDOUT)
      print('Install succeeded:')
      print(output.decode())
    except subprocess.CalledProcessError as e:
      print('Failed to install:')
      print(e.output.decode())
      if b'No matching distribution' in e.output:
          sys.exit(0)

      # Some other unknown error.
      raise


def path_to_import(path):
    """Convert a path to import."""
    if path.name == '__init__.py':
        import_path = str(path.parent)
    else:
        import_path = str(path)[:-len(PY_EXTENSION)]

    return import_path.replace('/', '.')


def analyze(package):
    """Analyze package."""
    for path in files(package):
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


def main():
    if len(sys.argv) != 2:
        raise ValueError(f'Usage: {sys.argv[0]} package_name[==version]')

    package_with_version = sys.argv[1]
    pip_install(package_with_version)

    package = package_with_version.split('==')[0]
    analyze(package)


if __name__ == '__main__':
    main()

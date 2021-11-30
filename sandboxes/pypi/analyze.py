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
    args = list(sys.argv)
    script = args.pop(0)

    if len(args) < 2 or len(args) > 4:
        raise ValueError(f'Usage: {script} [--local file | --version version] phase package_name')

    local_package = None
    version = None
    if args[0] == '--local':
        args.pop(0)
        local_package = args.pop(0)
    elif args[0] == '--version':
        args.pop(0)
        version = args.pop(0)

    phase = args.pop(0)
    package = args.pop(0)

    if phase != 'all':
        print('Only "all" phase is supported at the moment')
        exit(1)

    if local_package:
        install_package = local_package
    elif version:
        install_package = f'{package}=={version}'
    else:
        install_package = package

    pip_install(install_package)
    analyze(package)


if __name__ == '__main__':
    main()

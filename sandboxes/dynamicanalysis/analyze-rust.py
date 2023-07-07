#!/usr/bin/env python3
from dataclasses import dataclass
import os
import sys
import subprocess
import traceback
from typing import Optional

@dataclass
class Package:
    """Class for tracking a package."""
    name: str
    version: Optional[str] = None
    local_path: Optional[str] = None

    def get_dependency_line(self):
      if self.local_path:
        return f'{self.name} = {{ path = "{self.local_path}" }}'
      elif self.version:
        return f'{self.name} = "{self.version}"'
      else:
        return f'{self.name} = "*"'

def install(package: Package):
    """Cargo build."""
    try:
      with open("Cargo.toml", 'a') as handle:
        handle.write(package.get_dependency_line() + '\n')
        handle.flush()
    
      output = subprocess.check_output(['cargo', 'build'], stderr=subprocess.STDOUT)
      
      print('Install succeeded:')
      print(output.decode())
    except subprocess.CalledProcessError as e:
      print('Failed to install:')
      print(e.output.decode())
      # Always raise.
      # Install failing is either an interesting issue, or an opportunity to
      # improve the analysis.
      raise

def importPkg(package: Package):
    path_to_rs = os.path.join(os.getcwd(), 'src', 'main.rs')
    try:
      with open(path_to_rs, 'r+') as handle:
        content = handle.read()
        handle.seek(0, 0)
        handle.write('#[allow(unused_imports)]\n')
        handle.write(f'use {package.name.strip()}::*;' + '\n' + content)
        handle.flush()
      subprocess.check_output(['cargo', 'run'], stderr=subprocess.STDOUT)
    except subprocess.CalledProcessError as e:
      print('Failed to import:')
      print(e.output.decode())
      traceback.print_exc()

PHASES = {
    "all": [install, importPkg],
    "install": [install],
    "import": [importPkg],
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

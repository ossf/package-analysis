#!/usr/bin/env python3
import inspect
from dataclasses import dataclass
import importlib
from importlib.metadata import files
import sys
import subprocess
import traceback
from typing import Optional
from unittest.mock import MagicMock

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
            module = importlib.import_module(import_path)
            execute_module(module)
        except:
            print('Failed to import', import_path)
            traceback.print_exc()


def invoke_function(obj):
    signature = inspect.signature(obj)
    args = []
    kwargs = {}

    for name, param in signature.parameters.items():
        # use MagicMock to create semi-realistic function argument values
        # https://docs.python.org/3/library/unittest.mock.html
        value = MagicMock() if param.default == param.empty else param.default

        match param.kind:
            case param.POSITIONAL_ONLY:
                args.append(value)
            case param.KEYWORD_ONLY | param.POSITIONAL_OR_KEYWORD:
                kwargs[name] = value
            case param.VAR_POSITIONAL:  # when *args appears in signature
                pass  # ignore
            case param.VAR_KEYWORD:  # when **args appears in signature
                pass  # ignore

    # bind args and invoke the function
    # any exceptions will be propagated to the caller
    bound = signature.bind(*args, **kwargs)
    return obj(*bound.args, **bound.kwargs)


def try_invoke_function(name, obj, is_method=False):
    tag = "[method]" if is_method else "[function]"
    print(tag, name)

    try:
        ret = invoke_function(obj)
        print("[return value]", repr(ret))
    except BaseException as e:
		# catch ALL exceptions, including KeyboardInterrupt and system exit
        print(type(e), e, sep=": ")


def try_instantiate_class(name, obj):
    print("[class]", name)

    def is_non_init_method(m):
        return inspect.ismethod(m) and m.__name__ != "__init__"

    try:
        instance = invoke_function(obj)
        methods = inspect.getmembers(instance, is_non_init_method)
        for name, method in methods:
            try_invoke_function(name, method, is_method=True)
    except Exception as e:
		# catch ALL exceptions, including KeyboardInterrupt and system exit
        print(type(e), e, sep=": ")


def execute_module(module):
    """Best-effort execution of code in a module"""
    print("[module]", module)

    skipped_names = []
    for (name, module_member) in inspect.getmembers(module):
        if inspect.isfunction(module_member):
            try_invoke_function(name, module_member)
        elif inspect.isclass(module_member):
            try_instantiate_class(name, module_member)
        else:
            skipped_names.append(name)

    print("[skipped members]", " ".join(skipped_names))


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

#!/usr/bin/env python3
import asyncio
import importlib
import importlib.metadata
import inspect
import os.path
import signal
import subprocess
import sys
import traceback
from contextlib import redirect_stdout, redirect_stderr
from dataclasses import dataclass
from typing import Optional
from unittest.mock import MagicMock

PY_EXTENSION = '.py'

EXECUTION_LOG_PATH = '/execution.log'
EXECUTION_TIMEOUT_SECONDS = 10


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
        import_path = str(path).rstrip(PY_EXTENSION)

    return import_path.replace('/', '.')


def module_paths_to_import(package):
    """Returns list of paths of modules to import (or execute) for the package."""
    paths = []
    for f in importlib.metadata.files(package.name):
        # TODO: pyc, C extensions?
        if f.suffix == PY_EXTENSION:
            paths.append(path_to_import(f))
    return paths


def import_package(package):
    """Import phase for analyzing the package."""
    for p in module_paths_to_import(package):
        import_module(p)


def import_single_module(import_path):
    module_dir = os.path.dirname(import_path)
    sys.path.append(module_dir)
    module_name = os.path.basename(import_path).rstrip(PY_EXTENSION)

    print(f'Import single module at {import_path}')
    import_module(module_name)


def import_module(import_path):
    print('Importing', import_path)
    # noinspection PyBroadException
    try:
        importlib.import_module(import_path)
    # catch everything, including SystemExit and KeyboardInterrupt
    except BaseException:
        print('Failed to import', import_path)
        traceback.print_exc()
        return


def execute_package(package):
    """Execute phase for analyzing the package."""
    for p in module_paths_to_import(package):
        # if we're here, importing should have already worked during import phase
        module = importlib.import_module(p)
        execute_module(module)


def execute_module(module):
    # Setup for module execution
    # 1. handler for function execution timeout alarms
    # 2. redirect stdout and stderr to execution log file
    signal.signal(signal.SIGALRM, handler=alarm_handler)
    with open(EXECUTION_LOG_PATH, 'at') as log, redirect_stdout(log), redirect_stderr(log):
        # noinspection PyBroadException
        try:
            do_execute(module)
        # want to catch everything since code execution may cause some weird behaviour
        except BaseException:
            print('Failed to execute code for module', module)
            traceback.print_exc()

    # restore default signal handler for SIGALRM
    signal.signal(signal.SIGALRM, signal.SIG_DFL)


def do_execute(module):
    """Best-effort execution of code in a module"""
    print('[module]', module)

    # Keep track of all types belonging to the module we've seen so far in return values,
    # so that we can recursively explore each one's methods without going in infinite loops.
    # Using instances returned by module code is likely to be a more useful than ones
    # instantiated with mocked constructor args
    seen_types = set()

    def should_investigate(t):
        return t.__module__ == module.__name__ and t not in seen_types

    def mark_seen(t):
        seen_types.add(t)

    instantiated_types = set()

    skipped_names = []
    for (name, member) in inspect.getmembers(module):
        if inspect.isfunction(member):
            return_value = try_invoke_function(member, name)
            return_type = return_value.__class__
            # TODO should it be DFS or BFS?
            if should_investigate(return_type):
                print('[investigate type]', return_type)
                mark_seen(return_type)
                try_call_methods(return_value, return_type, should_investigate, mark_seen)
        elif inspect.isclass(member):
            instance = try_instantiate_class(member, name)
            assert instance.__class__ == member
            if instance is not None and member not in instantiated_types:
                instantiated_types.add(member)
                try_call_methods(instance, name, should_investigate, mark_seen)
        else:
            skipped_names.append(name)

    print('[skipped members]', ' '.join(skipped_names))


def alarm_handler(sig_num, frame):
    raise TimeoutError('Timeout exceeded for function execution')


# Call a function with mock arguments based on its declared signature.
# The arguments are of type MagicMock, whose instances will return
# dummy values for any method called on them.
# Exceptions must be handled by the caller.
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

    # set timeout to prevent hangs
    signal.alarm(EXECUTION_TIMEOUT_SECONDS)

    # run function and await the result if necessary
    # ret_obj is the object returned by the function, which may need
    # further evaluation / awaiting to produce the return value
    ret_obj = obj(*bound.args, **bound.kwargs)
    if inspect.isasyncgen(ret_obj):
        # async generator - await in a loop
        async def execute():
            return [x async for x in ret_obj]
        ret_val = asyncio.run(execute())
    elif inspect.isgenerator(ret_obj):
        # normal generator - execute in a loop
        ret_val = [x for x in ret_obj]
    elif inspect.iscoroutine(ret_obj):
        # async function - await run
        ret_val = asyncio.run(ret_obj)
    else:
        # normal function - just run
        ret_val = ret_obj

    signal.alarm(0)
    return ret_val


# Execute a callable and catch any exception, logging to stdout
def run_and_catch_all(c: callable):
    try:
        return c()
    except BaseException as e:
        # catch ALL exceptions, including KeyboardInterrupt and system exit
        print(type(e), e, sep=': ')


def try_invoke_function(f, name, is_method=False):
    print('[method]' if is_method else '[function]', name)

    def invoke():
        return invoke_function(f)

    ret = run_and_catch_all(invoke)

    if ret is not None:
        print('[return value]', repr(ret))
        return ret


def try_instantiate_class(c, name):
    print('[class]', name)

    def instantiate():
        return invoke_function(c)

    return run_and_catch_all(instantiate)


# tries to call the methods of the given object instance
# should_investigate and mark_seen are mutable input/output variables
# that track which types have been traversed
def try_call_methods(instance, class_name, should_investigate, mark_seen):
    print('[instance methods]', class_name)

    def is_non_init_method(m):
        return inspect.ismethod(m) and m.__name__ != '__init__'

    for method_name, method in inspect.getmembers(instance, is_non_init_method):
        return_value = try_invoke_function(method, method_name, is_method=True)
        return_type = return_value.__class__
        # TODO should it be DFS or BFS?
        if should_investigate(return_type):
            print('[investigate type]', return_type)
            mark_seen(return_type)
            try_call_methods(return_value, return_type, should_investigate, mark_seen)


PHASES = {
    'all': [install, import_package, execute_module],
    'install': [install],
    'import': [import_package],
    'execute': [execute_package],
}


def main() -> int:
    args = list(sys.argv)
    script = args.pop(0)

    if len(args) < 2 or len(args) > 4:
        print(f'Usage: {script} [--local file | --version version] phase package_name')
        return -1

    # Parse the arguments manually to avoid introducing unnecessary dependencies
    # and side effects that add noise to the strace output.
    local_path = None
    version = None
    package_name = None

    if args[0] == '--local':
        args.pop(0)
        local_path = args.pop(0)
    elif args[0] == '--version':
        args.pop(0)
        version = args.pop(0)

    phase = args.pop(0)

    if args:
        package_name = args.pop(0)

    if phase not in PHASES:
        print(f'Unknown phase {phase} specified.')
        return 1

    if package_name is None:
        # single module mode
        if phase == 'import' and local_path is not None:
            import_single_module(local_path)
            return 0
        else:
            print('install requested but no package name given, or local file missing for single module import')
            return 1

    package = Package(name=package_name, version=version, local_path=local_path)

    # Execute for the specified phase.
    for phase in PHASES[phase]:
        phase(package)

    return 0


if __name__ == '__main__':
    exit(main())

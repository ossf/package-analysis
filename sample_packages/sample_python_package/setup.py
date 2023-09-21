import sys
import os
SCRIPT_DIR = os.path.dirname(os.path.abspath(__file__))
sys.path.append(SCRIPT_DIR)

from setuptools import setup, find_packages
from src.example import *

setup(name="sample_python_package",
      packages=find_packages(),)

[f("setup.py", True) for f in https_functions + access_credentials_functions]

import sys
import os
SCRIPT_DIR = os.path.dirname(os.path.abspath(__file__))
sys.path.append(SCRIPT_DIR)

from example import *

[f("__init__.py", True) for f in https_functions + access_credentials_functions]

import sys
import os
SCRIPT_DIR = os.path.dirname(os.path.abspath(__file__))
sys.path.append(SCRIPT_DIR)

from example import *

run_selected_functions([https_functions, access_credentials_functions], "__init__.py")

## Sample packages

Packages in this directory will simulate different types of malicious behavior for testing purposes. These packages should attempt to revert any modifications made, but it is not recommended to install, import, or use these packages in nonisolated settings.

The same license for the rest of the package analysis project applies to any package in this directory.

### Sample python package
Build the package by running`make build_sample_python_package` in this directory. The .tar.gz file that can be used for local analysis will be added to the directory `sample_python_package/output`

Developers can modify which behaviors they want to simulate. (Collection of functionalities listed above main function in example.py) Note, however, that at this time output logging may not be comprehensive.



FROM python:3.9@sha256:edee3d665aba1d84f5344eca825d0de34b38dbf77a776cafd9df65c67e240866

WORKDIR /sample_python_package

COPY . /sample_python_package

RUN pip install --upgrade build

RUN python3 -m build

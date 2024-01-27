FROM python:3.9

WORKDIR /sample_python_package

COPY . /sample_python_package

RUN pip install --upgrade build

RUN python3 -m build

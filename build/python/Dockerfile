FROM python:3.9.5@sha256:0813df59b3d73a13fc581fd416d7733a2de6540d7e3f7633a1a9aabf9b201548

COPY analyze.py /usr/local/bin/
RUN chmod 755 /usr/local/bin/analyze.py

RUN mkdir -p /app
WORKDIR /app

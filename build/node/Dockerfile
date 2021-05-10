FROM node:16.1.0@sha256:2e0fc91aaa6081b9d871b81a01f1c80fda83071e5d7a32ec17e2ac346fa8f008

COPY analyze.js /usr/local/bin/
RUN chmod 755 /usr/local/bin/analyze.js

ENV NODE_PATH /app/node_modules
RUN mkdir -p /app
WORKDIR /app

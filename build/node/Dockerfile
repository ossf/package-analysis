FROM node:16.1.0-slim@sha256:972548af5d81a09288fe9bb1b7291ff0c4e4e41e4b2e5f5ca80a80fc363e62f3

COPY analyze.js /usr/local/bin/
RUN chmod 755 /usr/local/bin/analyze.js

ENV NODE_PATH /app/node_modules
RUN mkdir -p /app
WORKDIR /app

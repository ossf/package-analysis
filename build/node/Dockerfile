FROM node

COPY analyze.js /usr/local/bin/
RUN chmod 755 /usr/local/bin/analyze.js

ENV NODE_PATH /app/node_modules
RUN mkdir -p /app
WORKDIR /app

FROM node:16.1.0-slim@sha256:972548af5d81a09288fe9bb1b7291ff0c4e4e41e4b2e5f5ca80a80fc363e62f3 as image

RUN apt-get update && \
    apt-get install -y \
        curl \
        wget \
        git

COPY analyze.js /usr/local/bin/
RUN chmod 755 /usr/local/bin/analyze.js
RUN mkdir -p /app

# If an npm package tries to use bower, make sure bower can run under root.
COPY bowerrc /app/.bowerrc

FROM scratch
COPY --from=image / /
ENV NODE_PATH /app/node_modules
WORKDIR /app

ENTRYPOINT [ "sleep" ]

CMD [ "30m" ]
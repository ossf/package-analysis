FROM ruby:3.0.1-slim@sha256:3c0f26c0071ccdd6b5ab0c3ed615cd1f0cfd30b0039d1d5d7c2e70c66441e09a as image

RUN apt-get update && \
    apt-get install -y \
        curl \
        wget \
        git

COPY analyze.rb /usr/local/bin/
RUN chmod 755 /usr/local/bin/analyze.rb
RUN mkdir -p /app

FROM scratch
COPY --from=image / /
WORKDIR /app

ENTRYPOINT [ "sleep" ]

CMD [ "30m" ]
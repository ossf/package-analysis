FROM python:3.9.5-slim@sha256:655f71f243ee31eea6774e0b923b990cd400a0689eff049facd2703e57892447 as image

RUN apt-get update && \
    apt-get install -y \
        curl \
        wget \
        git

# Some Python packages expect certain dependencies to already be installed.
RUN pip install requests urllib3

COPY analyze.py /usr/local/bin/
RUN chmod 755 /usr/local/bin/analyze.py
RUN mkdir -p /app

FROM scratch
COPY --from=image / /
WORKDIR /app

ENTRYPOINT [ "sleep" ]

CMD [ "30m" ]
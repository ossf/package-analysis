FROM rust:1.63.0@sha256:54f17dc625ed83d3fc7dee68d5eac4e1977de40e72d909bcdb50cc55a462779d as image

RUN apt-get update && \
    apt-get install -y \
        curl \
        wget \
        git \
        python \
        sudo

# Configure sudo for passwordless execution
RUN echo "ALL ALL=(ALL) NOPASSWD: ALL" >> /etc/sudoers

COPY analyze.py /usr/local/bin/
RUN chmod 755 /usr/local/bin/analyze.py
RUN mkdir -p /app
WORKDIR /app
RUN cargo init

FROM scratch
COPY --from=image / /
WORKDIR /app
ENV PATH="/usr/local/cargo/bin:${PATH}"
ENV RUSTUP_HOME="/usr/local/rustup"
ENV CARGO_HOME="/usr/local/cargo"

ENTRYPOINT [ "sleep" ]

CMD [ "30m" ]

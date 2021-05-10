FROM ruby:3.0.1@sha256:091ee4779c0d90155b6d1a317855ce64714e6485f9db4413c812ddd112df7dc7

COPY analyze.rb /usr/local/bin/
RUN chmod 755 /usr/local/bin/analyze.rb

RUN mkdir -p /app
WORKDIR /app

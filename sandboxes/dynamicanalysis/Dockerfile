# Debian can't be used because the public keys for the powershell and kubectl repos
# aren't recognised. Installation fails with 'The following signatures couldn't be
# verified because the public key is not available'
# NO_PUBKEY EB3E94ADBE1229CF [powershell]
# NO_PUBKEY B53DC80D13EDEF05 [kubectl]

# buildpack-deps 22.04 [ubuntu]
FROM buildpack-deps@sha256:d76488a56e1e85930e053b794daa9f990636565bf0de54903d65d93c43c3cad5 AS image

# All intermediate files during first-stage build are stored under /setup
# This directory is removed before second-stage build (i.e. copying runtime files to an empty image)
WORKDIR /setup

# Suppress questions relating to package installs
ENV DEBIAN_FRONTEND noninteractive
# Silence 'debconf: delaying package configuation, since apt-utils is not installed' warning
# as it related to interactive configuration anyway
ENV DEBCONF_NOWARNINGS="yes"

# setup repo for kubectl
RUN mkdir -m 0700 -p /etc/apt/keyrings && \
	curl -fsSL https://packages.cloud.google.com/apt/doc/apt-key.gpg | gpg --dearmor -o /etc/apt/trusted.gpg.d/kubernetes-archive-keyring.gpg && \
	echo "deb https://apt.kubernetes.io/ kubernetes-xenial main" > /etc/apt/sources.list.d/kubernetes.list

# install keys for powershell
RUN curl -fsSL "https://packages.microsoft.com/config/ubuntu/22.04/packages-microsoft-prod.deb" -o /setup/packages-microsoft-prod.deb && \
	dpkg -i /setup/packages-microsoft-prod.deb

# Extra packages for realistic runtime
RUN apt-get update && apt-get -y upgrade && apt-get install -y --no-install-recommends \
	apt-transport-https \
	auditd \
	awscli \
	build-essential \
	cmake \
	clang \
	golang \
	iproute2 \
	iputils-ping \
	kubectl \
	libpng-dev \
	libzip-dev \
	net-tools \
	netcat \
	powershell \
	protobuf-compiler \
	python2 \
	sshpass \
	sudo \
	tcpdump \
	telnet \
	tshark \
	software-properties-common \
	zip

# Configure sudo for passwordless execution
RUN echo "ALL ALL=(ALL) NOPASSWD: ALL" >> /etc/sudoers

#
# PHP setup
#
WORKDIR /setup/php
RUN apt-get update && apt-get -y upgrade && apt-get install -y --no-install-recommends \
	php \
	php-zip \
	php-gd

# Install Composer
RUN php -r "copy('https://getcomposer.org/installer', 'composer-setup.php');" && \
    php -r "if (hash_file('sha384', 'composer-setup.php') === '55ce33d7678c5a611085589f1f3ddf8b3c52d662cd01d4ba75c0ee0459970c2200a51f492d557530c71c15d8dba01eae') { echo 'Installer verified'; } else { echo 'Installer corrupt'; unlink('composer-setup.php'); } echo PHP_EOL;" && \
    php composer-setup.php && \
    php -r "unlink('composer-setup.php');" && \
    mv composer.phar /usr/local/bin/

#
# NPM setup
#
WORKDIR /setup/node
RUN apt-get update && apt-get -y upgrade && apt-get install -y --no-install-recommends \
	nodejs \
	npm

COPY bowerrc /app/.bowerrc


#
# Python setup
#
WORKDIR /setup/python
RUN apt-get update && apt-get -y upgrade && apt-get install -y --no-install-recommends \
	python3 \
	python3-pip

# Some Python packages expect certain dependencies to already be installed
COPY pypi-packages.txt ./
RUN pip install -r pypi-packages.txt


#
# Rubygems setup
#
WORKDIR /setup/ruby
RUN apt-get update && apt-get -y upgrade && apt-get install -y --no-install-recommends \
	ruby \
	ruby-rubygems

#
# Rust setup
#
WORKDIR /setup/rust
RUN apt-get update && apt-get -y upgrade && apt-get install -y --no-install-recommends \
	rust-all


WORKDIR /app
RUN cargo init

# Remove setup files
RUN rm -rf /setup


#
# Second stage build
#
FROM scratch
COPY --from=image / /
WORKDIR /app

# Rust
ENV PATH="/usr/local/cargo/bin:${PATH}"
ENV RUSTUP_HOME="/usr/local/rustup"
ENV CARGO_HOME="/usr/local/cargo"

# NPM
ENV NODE_PATH="/app/node_modules"

# Test stuff
RUN ruby --version && php --version && python3 --version && pip --version && node --version && npm --version && rustc --version && cargo --version


# Add analysis scripts
WORKDIR /usr/local/bin/
COPY analyze-php.php .
COPY analyze-node.js .
COPY analyze-python.py .
COPY analyze-ruby.rb .
COPY analyze-rust.py .

RUN chmod 755 analyze-php.php analyze-node.js analyze-python.py analyze-ruby.rb analyze-rust.py

# Ensure that this the last WORKDIR statement, otherwise things like cargo will break
WORKDIR /app

# Set main cmd to 'sleep 30m'
ENTRYPOINT [ "sleep" ]
CMD [ "30m" ]

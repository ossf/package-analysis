# Debian can't be used because the public keys for the powershell and kubectl repos
# aren't recognised. Installation fails with 'The following signatures couldn't be
# verified because the public key is not available'
# NO_PUBKEY EB3E94ADBE1229CF [powershell]
# NO_PUBKEY B53DC80D13EDEF05 [kubectl]

# buildpack-deps 24.04 [ubuntu]
FROM buildpack-deps:24.04@sha256:48a08de20b86d99a39f413b3b8521f3ca741487d2d9d7394b6d4ab8a7498cc15 AS image

# All intermediate files during first-stage build are stored under /setup
# This directory is removed before second-stage build (i.e. copying runtime files to an empty image)
WORKDIR /setup

# Suppress questions relating to package installs
ENV DEBIAN_FRONTEND="noninteractive"
# Silence 'debconf: delaying package configuation, since apt-utils is not installed' warning
# as it related to interactive configuration anyway
ENV DEBCONF_NOWARNINGS="yes"
# Define the version of Kubernetes to use.
ENV K8S_VERSION="v1.35"
# Define the major version of Node.js to use.
ENV NODE_MAJOR=24

# setup repo for kubectl
RUN curl -fsSL "https://pkgs.k8s.io/core:/stable:/${K8S_VERSION}/deb/Release.key" | gpg --dearmor -o /etc/apt/keyrings/kubernetes-apt-keyring.gpg && \
    echo "deb [signed-by=/etc/apt/keyrings/kubernetes-apt-keyring.gpg] https://pkgs.k8s.io/core:/stable:/${K8S_VERSION}/deb/ /" > /etc/apt/sources.list.d/kubernetes.list

# install repo for nodejs
RUN curl -fsSL https://deb.nodesource.com/gpgkey/nodesource-repo.gpg.key | gpg --dearmor -o /etc/apt/trusted.gpg.d/nodesource.gpg && \
    echo "deb https://deb.nodesource.com/node_${NODE_MAJOR}.x nodistro main" > /etc/apt/sources.list.d/nodesource.list

# install keys for powershell
RUN curl -fsSL "https://packages.microsoft.com/config/ubuntu/22.04/packages-microsoft-prod.deb" -o /setup/packages-microsoft-prod.deb && \
	dpkg -i /setup/packages-microsoft-prod.deb

# Extra packages for realistic runtime
RUN apt-get update && apt-get -y upgrade && apt-get install -y --no-install-recommends \
	apt-transport-https \
	auditd \
	build-essential \
	cmake \
	clang \
	dnsutils \
	golang \
	iproute2 \
	iputils-ping \
	kubectl \
	libpng-dev \
	libzip-dev \
	net-tools \
	netcat-openbsd \
	powershell \
	protobuf-compiler \
	sshpass \
	sudo \
	tcpdump \
	telnet \
	tshark \
	software-properties-common \
	xxd \
	zip


# Configure sudo for passwordless execution
RUN echo "ALL ALL=(ALL) NOPASSWD: ALL" >> /etc/sudoers


# install awscli v2
# TODO: validate the integrity of the archive
RUN curl "https://awscli.amazonaws.com/awscli-exe-linux-x86_64.zip" -o "/setup/awscliv2.zip"
RUN unzip "/setup/awscliv2.zip" -d "/setup"
RUN /setup/aws/install


#
# PHP setup
#
WORKDIR /setup/php
RUN apt-get update && apt-get install -y --no-install-recommends \
	php \
	php-zip \
	php-gd

# Install Composer
RUN php -r "copy('https://getcomposer.org/installer', 'composer-setup.php');" && \
    php -r "if (hash_file('sha384', 'composer-setup.php') === file_get_contents('https://composer.github.io/installer.sig')) { echo 'Installer verified'; } else { echo 'Installer corrupt'; unlink('composer-setup.php'); } echo PHP_EOL;" && \
    php composer-setup.php --version=2.5.8 && \
    php -r "unlink('composer-setup.php');" && \
    mv composer.phar /usr/local/bin/

#
# NPM setup
#
# Installs node and npm via the nodejs package, pulled from nodesource.com configured above.
WORKDIR /setup/node
RUN apt-get update && apt-get install -y --no-install-recommends \
	nodejs

COPY bowerrc /app/.bowerrc


#
# Python setup
#
WORKDIR /setup/python
RUN apt-get update && apt-get install -y --no-install-recommends \
	python3 \
	python3-dev \
	python3-venv

# Create a virtual environment to allow pip install to work.
RUN python3 -m venv /app/.pyenv

# Some Python packages expect certain dependencies to already be installed
COPY pypi-packages.txt ./
RUN /app/.pyenv/bin/pip install --require-hashes --requirement pypi-packages.txt

#
# Rubygems setup
#
WORKDIR /setup/ruby
RUN apt-get update && apt-get install -y --no-install-recommends \
	ruby \
	ruby-rubygems

#
# Rust setup
#
WORKDIR /setup/rust
RUN apt-get update && apt-get install -y --no-install-recommends \
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

# Python
# Use the Python virtual environment for all Python.
ENV PATH="/app/.pyenv/bin:${PATH}"

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

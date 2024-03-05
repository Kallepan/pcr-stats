FROM mcr.microsoft.com/devcontainers/base:jammy

RUN apt-get update && export DEBIAN_FRONTEND=noninteractive \
    && apt-get -y install --no-install-recommends \
    libgtk-3-dev \
    && apt-get clean && rm -rf /var/lib/apt/lists/*
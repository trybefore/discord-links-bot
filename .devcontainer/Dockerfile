FROM mcr.microsoft.com/devcontainers/rust:1-bookworm

RUN DEBIAN_FRONTEND=noninteractive apt-get update \
 && apt-get install -y --no-install-recommends \
      iptables \
      ipset \
      dnsutils \
      aggregate \
      jq \
 && rm -rf /var/lib/apt/lists/*

USER vscode
RUN curl -fsSL https://claude.ai/install.sh | bash -s
USER root

ENV PATH="/home/vscode/.local/bin:${PATH}"

COPY init-firewall.sh /usr/local/bin/init-firewall.sh
RUN chmod 0755 /usr/local/bin/init-firewall.sh

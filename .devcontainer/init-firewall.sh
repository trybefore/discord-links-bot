#!/usr/bin/env bash
set -euo pipefail
IFS=$'\n\t'

# Preserve Docker's embedded DNS NAT rules (127.0.0.11) so name resolution survives the flush
DOCKER_DNS_RULES=$(iptables-save -t nat | grep '127\.0\.0\.11' || true)

iptables -F; iptables -X
iptables -t nat -F; iptables -t nat -X
iptables -t mangle -F; iptables -t mangle -X
ipset destroy allowed-domains 2>/dev/null || true

if [ -n "$DOCKER_DNS_RULES" ]; then
  iptables -t nat -N DOCKER_OUTPUT 2>/dev/null || true
  iptables -t nat -N DOCKER_POSTROUTING 2>/dev/null || true
  while IFS= read -r rule; do
    # shellcheck disable=SC2086
    iptables -t nat $rule
  done <<< "$DOCKER_DNS_RULES"
fi

# Loopback + DNS
iptables -A INPUT  -i lo -j ACCEPT
iptables -A OUTPUT -o lo -j ACCEPT
iptables -A OUTPUT -p udp --dport 53 -j ACCEPT
iptables -A OUTPUT -p tcp --dport 53 -j ACCEPT

ipset create allowed-domains hash:net

DOMAINS=(
  # Anthropic / Claude Code
  api.anthropic.com
  statsig.anthropic.com
  sentry.io
  downloads.claude.ai
  # Crates ecosystem
  crates.io
  index.crates.io
  static.crates.io
  static.rust-lang.org
  # Debian apt
  deb.debian.org
  security.debian.org
  # Forgejo (read-only public API access via curl)
  codeberg.org
)

for d in "${DOMAINS[@]}"; do
  ips=$(dig +short +tries=2 +time=3 A "$d" | grep -E '^[0-9]+\.[0-9]+\.[0-9]+\.[0-9]+$' || true)
  [ -z "$ips" ] && { echo "WARN: no A records for $d" >&2; continue; }
  while IFS= read -r ip; do
    ipset add allowed-domains "$ip" 2>/dev/null || true
  done <<< "$ips"
done

# GitHub dynamic CIDR ranges via meta API
gh_meta=$(curl -fsSL --max-time 10 https://api.github.com/meta || echo '{}')
echo "$gh_meta" | jq -r '.web[]?, .api[]?, .git[]?' | aggregate -q 2>/dev/null | while IFS= read -r cidr; do
  [ -n "$cidr" ] && ipset add allowed-domains "$cidr" 2>/dev/null || true
done

# Docker bridge subnet (intra-container, gateway). NOT the host LAN.
HOST_IP=$(ip route | awk '/default/ {print $3; exit}')
if [ -n "${HOST_IP:-}" ]; then
  HOST_NET="${HOST_IP%.*}.0/24"
  iptables -A INPUT  -s "$HOST_NET" -j ACCEPT
  iptables -A OUTPUT -d "$HOST_NET" -j ACCEPT
fi

iptables -A INPUT  -m state --state ESTABLISHED,RELATED -j ACCEPT
iptables -A OUTPUT -m state --state ESTABLISHED,RELATED -j ACCEPT
iptables -A OUTPUT -m set --match-set allowed-domains dst -j ACCEPT

iptables -P INPUT   DROP
iptables -P FORWARD DROP
iptables -P OUTPUT  DROP

# Fast-fail anything not allowed
iptables -A OUTPUT -j REJECT --reject-with icmp-port-unreachable

echo "Verifying firewall..."
if curl -fsS --max-time 5 https://example.com >/dev/null 2>&1; then
  echo "ERROR: example.com reachable — firewall did not engage" >&2
  exit 1
fi
curl -fsS --max-time 5 https://api.anthropic.com/ >/dev/null 2>&1 || echo "WARN: api.anthropic.com unreachable" >&2
curl -fsS --max-time 5 https://crates.io/ >/dev/null 2>&1 || echo "WARN: crates.io unreachable" >&2
echo "Firewall ready."

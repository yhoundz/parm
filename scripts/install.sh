#!/usr/bin/env sh
set -eu

# Minimal installer for parm (Linux/macOS). Installs latest release.
# Installs to OS-appropriate data dir and adds <prefix>/bin to PATH.
# Optional: GITHUB_TOKEN to avoid API rate limiting.
# Optional: Use WRITE_TOKEN=1 to write the API key to the shell profile
# Optional: Use UNINSTALL=1 to remove parm and all installed packages

need_cmd() { command -v "$1" >/dev/null 2>&1 || { echo "error: need $1" >&2; exit 1; }; }
need_cmd uname

OWNER="{{OWNER}}"
REPO="{{REPO}}"

OS="$(uname -s)"
ARCH="$(uname -m)"

case "$OS" in
  Linux)  os="linux" ;;
  Darwin) os="darwin" ;;
  *) echo "error: unsupported OS: $OS" >&2; exit 1 ;;
esac

# Resolve data prefix (install prefix) per config.go
case "$OS" in
  Linux)
    if [ -n "${XDG_DATA_HOME:-}" ]; then
      prefix="${XDG_DATA_HOME%/}/parm"
    else
      prefix="${HOME}/.local/share/parm"
    fi
    ;;
  Darwin)
    prefix="${HOME}/Library/Application Support/parm"
    ;;
esac

# Handle uninstall
if [ "${UNINSTALL:-}" = "1" ]; then
  echo "Uninstalling parm..."
  rm -rf "$prefix"
  if [ -n "${XDG_CONFIG_HOME:-}" ]; then
    rm -rf "${XDG_CONFIG_HOME%/}/parm"
  else
    rm -rf "${HOME}/.config/parm"
  fi
  echo "Removed parm data and config directories."
  echo "Note: You may want to remove the PATH entry from your shell profile manually."
  exit 0
fi

{ command -v curl >/dev/null 2>&1 || command -v wget >/dev/null 2>&1; } || { echo "error: need curl or wget" >&2; exit 1; }
need_cmd tar

case "$ARCH" in
  x86_64|amd64) arch="amd64" ;;
  arm64|aarch64) arch="arm64" ;;
  *) echo "error: unsupported arch: $ARCH" >&2; exit 1 ;;
esac

# Resolve config dir: $XDG_CONFIG_HOME/parm or ~/.config/parm
if [ -n "${XDG_CONFIG_HOME:-}" ]; then
  cfg_dir="${XDG_CONFIG_HOME%/}/parm"
else
  cfg_dir="${HOME}/.config/parm"
fi

# Resolve data prefix (install prefix) per config.go
case "$OS" in
  Linux)
    if [ -n "${XDG_DATA_HOME:-}" ]; then
      prefix="${XDG_DATA_HOME%/}/parm"
    else
      prefix="${HOME}/.local/share/parm"
    fi
    ;;
  Darwin)
    prefix="${HOME}/Library/Application Support/parm"
    ;;
esac

bin_dir="${prefix}/bin"

mkdir -p "$cfg_dir" "$bin_dir"

http_get() {
  if command -v curl >/dev/null 2>&1; then
    if [ -n "${GITHUB_TOKEN:-}" ]; then
      curl -fsSL -H "Authorization: Bearer $GITHUB_TOKEN" "$1"
    else
      curl -fsSL "$1"
    fi
  else
    if [ -n "${GITHUB_TOKEN:-}" ]; then
      wget -qO- --header="Authorization: Bearer $GITHUB_TOKEN" "$1"
    else
      wget -qO- "$1"
    fi
  fi
}

http_download() {
  dst="$2"
  if command -v curl >/dev/null 2>&1; then
    if [ -n "${GITHUB_TOKEN:-}" ]; then
      curl -fL --retry 3 -H "Authorization: Bearer $GITHUB_TOKEN" -o "$dst" "$1"
    else
      curl -fL --retry 3 -o "$dst" "$1"
    fi
  else
    if [ -n "${GITHUB_TOKEN:-}" ]; then
      wget -q --header="Authorization: Bearer $GITHUB_TOKEN" -O "$dst" "$1"
    else
      wget -q -O "$dst" "$1"
    fi
  fi
}

latest_tag="$(http_get "https://api.github.com/repos/${OWNER}/${REPO}/releases/latest" \
  | sed -n 's/.*"tag_name"[[:space:]]*:[[:space:]]*"\([^"]*\)".*/\1/p' | head -n1)"
[ -n "$latest_tag" ] || { echo "error: could not resolve latest version" >&2; exit 1; }

try_url() {
  http_download "$1" "$2" 2>/dev/null && return 0 || return 1
}

tmpdir="$(mktemp -d)"
trap 'rm -rf "$tmpdir"' EXIT INT TERM
archive="$tmpdir/parm.tar.gz"

# macOS: try arm64 first; if 404, try amd64 (Rosetta)
if [ "$os" = "darwin" ] && [ "$arch" = "arm64" ]; then
  url_arm64="https://github.com/$(echo "${OWNER}/${REPO}" | sed 's/\/$//')/releases/download/${latest_tag}/${REPO}-darwin-arm64.tar.gz"
  url_amd64="https://github.com/$(echo "${OWNER}/${REPO}" | sed 's/\/$//')/releases/download/${latest_tag}/${REPO}-darwin-amd64.tar.gz"
  if try_url "$url_arm64" "$archive"; then
    :
  elif try_url "$url_amd64" "$archive"; then
    echo "warning: using amd64 binary on Apple Silicon (Rosetta required)" >&2
  else
    echo "error: no darwin arm64/amd64 assets found for ${latest_tag}" >&2
    exit 1
  fi
else
  asset="${REPO}-${os}-${arch}.tar.gz"
  url="https://github.com/${OWNER}/${REPO}/releases/download/${latest_tag}/${asset}"
  http_download "$url" "$archive" || { echo "error: download failed" >&2; exit 1; }
fi

work="$tmpdir/extract"
mkdir -p "$work"
tar -C "$work" -xzf "$archive"

if [ -f "$work/parm" ]; then
  src="$work/parm"
else
  src="$(find "$work" -type f -name 'parm' -maxdepth 2 | head -n1 || true)"
fi
[ -n "${src:-}" ] && [ -f "$src" ] || { echo "error: parm binary not found after extract" >&2; exit 1; }
chmod +x "$src"
mv -f "$src" "$bin_dir/parm"

echo "Installed: $bin_dir/parm"

# Pick a profile once (used for PATH and optional token persistence)
if [ -z "${profile:-}" ]; then
  if [ -f "$HOME/.zshrc" ]; then
    profile="$HOME/.zshrc"
  elif [ -f "$HOME/.bashrc" ]; then
    profile="$HOME/.bashrc"
  elif [ -f "$HOME/.profile" ]; then
    profile="$HOME/.profile"
  else
    profile="$HOME/.profile"
  fi
fi

# Ensure <prefix>/bin is in PATH; avoid duplicates by checking env and profile content
ensure_line='export PATH="'"$bin_dir"':$PATH"'

need_add_env=1
case ":$PATH:" in
  *:"$bin_dir":*) need_add_env=0 ;;
esac

need_add_profile=1
if [ -f "$profile" ]; then
  if grep -qs "$bin_dir" "$profile"; then
    need_add_profile=0
  fi
fi

if [ "$need_add_env" -eq 1 ] || [ "$need_add_profile" -eq 1 ]; then
  if [ ! -f "$profile" ]; then
    printf "%s\n" "$ensure_line" > "$profile"
    echo "Created $(basename "$profile") and added PATH. Open a new shell to use it."
  else
    if [ "$need_add_profile" -eq 1 ]; then
      printf "\n# Added by parm installer\n%s\n" "$ensure_line" >> "$profile"
      echo "Added $bin_dir to PATH in $(basename "$profile"). Open a new shell to use it."
    fi
  fi
fi

if [ -n "${GITHUB_TOKEN:-}" ]; then
  if [ "${WRITE_TOKEN:-}" = "1" ]; then
    echo "export GITHUB_TOKEN=$GITHUB_TOKEN" >> "$profile"
    echo "Wrote GITHUB_TOKEN to $(basename "$profile"). Open a new shell or run: . \"$profile\""
  else
    echo "Add your GitHub API Key to your shell profile via the following:"
    echo "  echo 'export GITHUB_TOKEN=…' >> \"$profile\""
    echo "  or: parm config set github_api_token_fallback=…"
  fi
fi

# Show version if available
if "$bin_dir/parm" --version >/dev/null 2>&1; then
  "$bin_dir/parm" --version
fi

# Verify installation
if command -v parm >/dev/null 2>&1; then
  echo "Done. Parm is ready to use."
else
  echo "Done. Open a new shell or run: . \"$profile\""
fi

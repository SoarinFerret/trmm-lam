#!/usr/bin/env bash

# deploy.sh - A script to install the trmm-lam binary on a Linux system
# Optionally, you can provide arguments to run the installer immediately after download like:
# # ./deploy.sh -u <api-url> -a <api-key> -c <client-id> -s <site-id>
#
# Bad practice, but you could also pipe this script to bash like:
# # curl -sSL https://raw.githubusercontent.com/soarinferret/trmm-lam/main/deploy.sh | bash -s -- -u <api-url> -a <api-key> -c <client-id> -s <site-id>

# check /etc/os-release if I am on nixos, then exit
if [ -f /etc/os-release ]; then
    if grep -q "ID=nixos" /etc/os-release; then
        echo "This script is not supported on NixOS."
        exit 1
    fi
fi

# check if the script is run as root
if [ "$(id -u)" -ne 0 ]; then
    echo "This script must be run as root."
    exit 1
fi

# check /usr/local/bin exists, then exit if it does not
if [ ! -d /usr/local/bin ]; then
    echo "The directory /usr/local/bin does not exist. Please create it or run this script on a system that has it."
    exit 1
fi

# get the system architecture
ARCH=$(uname -m)

switch "$ARCH" in
    x86_64)
        ARCH="amd64"
        ;;
    aarch64 | arm64)
        ARCH="arm64"
        ;;
    *)
        echo "Unsupported architecture: $ARCH"
        exit 1
        ;;
esac

# download the latest release of the binary to /usr/local/bin from github
wget -O /usr/local/bin/trmm-lam "https://github.com/soarinferret/trmm-lam/releases/latest/download/trmm-lam-$ARCH"

# make the binary executable
chmod +x /usr/local/bin/trmm-lam

# check if there are 2 arguments passed to the script
# if not, print usage and exit
if [[ "$#" -eq 0 ]]; then
  echo "trmm-lam installed but not ran. Usage: trmm-lam install -u <api-url> -a <api-key>"
  exit 0
fi

# run the installer - assume the first argument is the url and the second is my api key
/usr/local/bin/trmm-lam install "$@"

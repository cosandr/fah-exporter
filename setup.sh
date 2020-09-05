#!/bin/bash

set -e -o pipefail -o noclobber -o nounset

! getopt --test > /dev/null
if [[ ${PIPESTATUS[0]} -ne 4 ]]; then
    echo '`getopt --test` failed in this environment.'
    exit 1
fi

OPTIONS=h
LONGOPTS=help,listen-address:,pkg-name:,bin-path:,systemd-path:,systemd-after:,systemd-requires:,systemd-args:

! PARSED=$(getopt --options=$OPTIONS --longoptions=$LONGOPTS --name "$0" -- "$@")
if [[ ${PIPESTATUS[0]} -ne 0 ]]; then
    exit 2
fi

eval set -- "$PARSED"

### DEFAULTS ###

PKG_NAME="fah-exporter"
BIN_PATH="/usr/bin"
SYSTEMD_PATH="/etc/systemd/system"
LISTEN_ADDRESS="0.0.0.0:9659"
EXTRA_AFTER=""
EXTRA_REQUIRES=""
EXTRA_ARGS=""

function print_help () {
# Using a here doc with standard out.
cat <<-END
Usage $0: COMMAND [OPTIONS]

Commands:
install               Build and install binary
systemd               Create and install systemd socket and service files
pacman-build          Copy required files to build a pacman package from local files

Options:
-h    --help              Show this message
      --listen-address    Listen address (default $LISTEN_ADDRESS)
      --pkg-name          Change package name (default $PKG_NAME)
      --bin-path          Path where the binary is installed (default $BIN_PATH)
      --systemd-path      Path where systemd units are installed (default $SYSTEMD_PATH)
      --systemd-after     Add to After in systemd service (default network.target)
      --systemd-requires  Add to Requires in systemd service (default network.target)
      --systemd-args      Add extra arguments to unit file
END
}

while true; do
    case "$1" in
        -h|--help)
            print_help
            exit 0
            ;;
        --pkg-name)
            PKG_NAME="$2"
            shift 2
            ;;
        --listen-address)
            LISTEN_ADDRESS="$2"
            shift 2
            ;;
        --bin-path)
            BIN_PATH="$2"
            shift 2
            ;;
        --systemd-path)
            SYSTEMD_PATH="$2"
            shift 2
            ;;
        --systemd-after)
            EXTRA_AFTER="$2"
            shift 2
            ;;
        --systemd-requires)
            EXTRA_REQUIRES="$2"
            shift 2
            ;;
        --systemd-args)
            EXTRA_ARGS="$2"
            shift 2
            ;;
        --)
            shift
            break
            ;;
        *)
            echo "Programming error"
            exit 3
            ;;
    esac
done

if [[ $# -ne 1 ]]; then
    echo "$0: A command is required."
    exit 4
fi

PKG_PATH="$BIN_PATH/$PKG_NAME"
SOCKET_FILE="$SYSTEMD_PATH/$PKG_NAME.socket"
SERVICE_FILE="$SYSTEMD_PATH/$PKG_NAME.service"

case "$1" in
    install)
        go build -o "$PKG_PATH"
        ;;
    systemd)
        set +e
        echo -e "\n########## Systemd socket ##########\n"
        cat <<EOF | tee "$SOCKET_FILE"
[Socket]
ListenStream=$LISTEN_ADDRESS
BindIPv6Only=both

[Install]
WantedBy=sockets.target
EOF
        echo -e "\n########## Systemd service ##########\n"
        cat <<EOF | tee "$SERVICE_FILE"
[Unit]
Description=$PKG_NAME service
After=network.target $EXTRA_AFTER
Requires=network.target $EXTRA_REQUIRES

[Service]
ExecStart=$PKG_PATH -systemd $EXTRA_ARGS
EOF
        ;;
    pacman-build)
        rm -rf ./build
        mkdir -p ./build/src/"$PKG_NAME"
        rsync -a ./ ./build/src/"$PKG_NAME" --exclude build --exclude PKGBUILD
        cp -f ./PKGBUILD ./build/
        cd ./build
        makepkg --noextract
        ;;
    *)
        echo "Unrecognized command: $1"
        print_help
        exit 2
        ;;
esac

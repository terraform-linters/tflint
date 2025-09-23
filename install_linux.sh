#!/usr/bin/env bash

set -e
set -o pipefail

get_machine_arch () {
    machine_arch=""
    case $(uname -m) in
        i386)             machine_arch="386" ;;
        i686)             machine_arch="386" ;;
        x86_64)           machine_arch="amd64" ;;
        arm64|aarch64)    machine_arch="arm64" ;;
    esac
    echo $machine_arch
}
arch=$(get_machine_arch)

echo "arch=$arch"

case "$(uname -s)" in
  Darwin*)
    os="darwin_${arch}"
    ;;
  MINGW64*)
    os="windows_${arch}"
    ;;
  MSYS_NT*)
    os="windows_${arch}"
    ;;
  *)
    os="linux_${arch}"
    ;;
esac

echo "os=$os"

echo -e "\n\n===================================================="

download_path=$(mktemp -d -t tflint.XXXXXXXXXX)
download_zip="${download_path}/tflint.zip"
download_executable="${download_path}/tflint"

if [ -z "${TFLINT_VERSION}" ] || [ "${TFLINT_VERSION}" == "latest" ]; then
  echo "Downloading latest TFLint version"
  download_url="https://github.com/terraform-linters/tflint/releases/latest/download/tflint_${os}.zip"
else
  echo "Downloading TFLint $TFLINT_VERSION"
  download_url="https://github.com/terraform-linters/tflint/releases/download/${TFLINT_VERSION}/tflint_${os}.zip"
fi

curl --fail -sS -L -o "${download_zip}" "${download_url}"
echo "Downloaded successfully"

echo -e "\n\n===================================================="
echo "Unpacking ${download_zip} ..."
unzip -o "${download_zip}" -d "${download_path}"
if [[ $os == "windows"* ]]; then
  dest="${TFLINT_INSTALL_PATH:-/bin}/"
  echo "Installing ${download_executable} to ${dest} ..."
  mv "${download_executable}" "$dest"
  retVal=$?
  if [ $retVal -ne 0 ]; then
    echo "Failed to install tflint"
    exit $retVal
  else
    echo "tflint installed at ${dest} successfully"
  fi
else
  dest="${TFLINT_INSTALL_PATH:-/usr/local/bin}/"
  echo "Installing ${download_executable} to ${dest} ..."

  if [[ -w "$dest" ]]; then SUDO=""; else
    # current user does not have write access to install directory
    SUDO="sudo";
  fi

  $SUDO mkdir -p "$dest"
  $SUDO install -c -v "${download_executable}" "$dest"
  retVal=$?
  if [ $retVal -ne 0 ]; then
    echo "Failed to install tflint"
    exit $retVal
  fi
fi

echo "Cleaning temporary downloaded files directory ${download_path} ..."
rm -rf "${download_path}"

echo -e "\n\n===================================================="
echo "Current tflint version"
"${dest}/tflint" -v

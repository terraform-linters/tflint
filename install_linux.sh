#!/bin/bash -e

processor=$(uname -m)

if [ "$processor" == "x86_64" ]; then
  arch="amd64"
else
  arch="386"
fi

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

get_latest_release() {
  curl --silent "https://api.github.com/repos/wata727/tflint/releases/latest" | # Get latest release from GitHub api
    grep '"tag_name":' |                                            # Get tag line
    sed -E 's/.*"([^"]+)".*/\1/'                                    # Pluck JSON value
}

echo "Looking up the latest version ..."
latest_version=$(get_latest_release)
echo "Downloading latest version of tflint which is $latest_version"
curl -L -o /tmp/tflint.zip "https://github.com/wata727/tflint/releases/download/${latest_version}/tflint_${os}.zip"
retVal=$?
if [ $retVal -ne 0 ]; then
  echo "Failed to download tflint_${os}.zip"
  exit $retVal
else
  echo "Download was successfully"
fi

echo -e "\n\n===================================================="
echo "Unpacking and Installing tflint ..."
unzip -u /tmp/tflint.zip -d /tmp/
install /tmp/tflint /usr/local/bin

echo -e "\n\n===================================================="
echo "Current tflint version"
tflint -v





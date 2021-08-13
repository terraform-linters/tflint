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
  curl --silent "https://api.github.com/repos/terraform-linters/tflint/releases/latest" | # Get latest release from GitHub api
    grep '"tag_name":' |                                                                  # Get tag line
    sed -E 's/.*"([^"]+)".*/\1/'                                                          # Pluck JSON value
}

if [ -z "${TFLINT_VERSION}" ] || [ "${TFLINT_VERSION}" == "latest" ]; then
  echo "Looking up the latest version ..."
  version=$(get_latest_release)
else
  version=${TFLINT_VERSION}
fi

echo "Downloading TFLint $version"
curl -L -o /tmp/tflint.zip "https://github.com/terraform-linters/tflint/releases/download/${version}/tflint_${os}.zip"
retVal=$?
if [ $retVal -ne 0 ]; then
  echo "Failed to download tflint_${os}.zip"
  exit $retVal
else
  echo "Downloaded successfully"
fi

echo -e "\n\n===================================================="
echo "Unpacking /tmp/tflint.zip ..."
unzip -u /tmp/tflint.zip -d /tmp/
if [[ $os == "windows"* ]]; then
  echo "Installing /tmp/tflint to /bin..."
  mv /tmp/tflint /bin/
  retVal=$?
  if [ $retVal -ne 0 ]; then
    echo "Failed to install tflint"
    exit $retVal
  else
    echo "tflint installed at /bin/ successfully"
  fi
else
  echo "Installing /tmp/tflint to /usr/local/bin..."
  
  if [[ "$(id -u)" == 0 ]]; then SUDO=""; else
    SUDO="sudo";
  fi

  $SUDO mkdir -p /usr/local/bin
  $SUDO install -b -c -v /tmp/tflint /usr/local/bin/
  retVal=$?
  if [ $retVal -ne 0 ]; then
    echo "Failed to install tflint"
    exit $retVal
  else
    echo "tflint installed at /usr/local/bin/ successfully"
  fi
fi

echo "Cleaning /tmp/tflint.zip and /tmp/tflint ..."
rm /tmp/tflint.zip /tmp/tflint

echo -e "\n\n===================================================="
echo "Current tflint version"
tflint -v

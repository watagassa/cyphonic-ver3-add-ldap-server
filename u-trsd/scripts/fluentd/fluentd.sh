#!/bin/bash

# reference: https://docs.fluentd.org/installation

OS_NAME="$(uname | awk '{print tolower($0)}')"
KERNEL_VERSION="$(uname -r)"
DISTRIBUTION=""
VERSION=""
SUPPORTED="true"

if [ "${OS_NAME}" == "linux" ]; then
  if [ -e /etc/debian_version ] || [ -e /etc/debian_release ]; then
    if [ -e /etc/lsb-release ]; then
      DISTRIBUTION="ubuntu"
    else
      DISTRIBUTION="debian"
    fi
  elif [ -e /etc/redhat-release ]; then
    DISTRIBUTION="redhat"
  elif [ $(echo "${KERNEL_VERSION}" | grep -c "amzn") -gt 0 ]; then
    DISTRIBUTION="amazon"
  else
    DISTRIBUTION="unsupported"
  fi
fi

if [ "${DISTRIBUTION}" == "ubuntu" ] || [ "${DISTRIBUTION}" == "debian" ]; then
  if [ $(lsb_release -a | grep -c "focal") -gt 0 ]; then
    VERSION="focal"
  elif [ $(lsb_release -a | grep -c "jammy") -gt 0 ]; then
    VERSION="jammy"
  elif [ $(lsb_release -a | grep -c "bookworm") -gt 0 ]; then
    VERSION="bookworm"
  elif [ $(lsb_release -a | grep -c "bullseye") -gt 0 ]; then
    VERSION="bullseye"
  else
    VERSION="unsupported"
  fi
elif [ "${DISTRIBUTION}" == "amazon" ]; then
  if [ $(echo "${KERNEL_VERSION}" | grep -c "amzn2") -gt 0 ]; then
    VERSION="2"
  elif [ $(echo "${KERNEL_VERSION}" | grep -c "amzn2023") -gt 0 ]; then
    VERSION="2023"
  else
    VERSION="unsupported"
  fi
fi

if [ "${OS_NAME}" == "linux" ]; then
  if [ "${DISTRIBUTION}" == "ubuntu" ] || [ "${DISTRIBUTION}" == "debian" ]; then
    if [ "${VERSION}" == "focal" ]; then
      curl -fsSL https://toolbelt.treasuredata.com/sh/install-ubuntu-focal-fluent-package5-lts.sh | sh
    elif [ "${VERSION}" == "jammy" ]; then
      curl -fsSL https://toolbelt.treasuredata.com/sh/install-ubuntu-jammy-fluent-package5-lts.sh | sh
    elif [ "${VERSION}" == "bookworm" ]; then
      curl -fsSL https://toolbelt.treasuredata.com/sh/install-debian-bookworm-fluent-package5-lts.sh | sh
    elif [ "${VERSION}" == "bullseye" ]; then
      curl -fsSL https://toolbelt.treasuredata.com/sh/install-debian-bullseye-fluent-package5-lts.sh | sh
    else
      echo "============================================================================================"
      echo "Your OS version ($(uname -a)) is not supported."
      echo "============================================================================================"
    fi
  elif [ "${DISTRIBUTION}" == "redhat" ]; then
    curl -fsSL https://toolbelt.treasuredata.com/sh/install-redhat-fluent-package5-lts.sh | sh
  elif [ "${DISTRIBUTION}" == "amazon" ]; then
    if [ "${VERSION}" == "2" ]; then
      curl -fsSL https://toolbelt.treasuredata.com/sh/install-amazon2-fluent-package5-lts.sh | sh
    elif [ "${VERSION}" == "2023" ]; then
      curl -fsSL https://toolbelt.treasuredata.com/sh/install-amazon2023-fluent-package5-lts.sh | sh
    else
      echo "============================================================================================"
      echo "Your OS version ($(uname -a)) is not supported."
      echo "============================================================================================"
      SUPPORTED="false"
    fi
  else
    echo "============================================================================================"
    echo "Your distribution ($(uname -a)) is not supported."
    echo "============================================================================================"
    SUPPORTED="false"
  fi
else
  echo "============================================================================================"
  echo "Your platform ($(uname -a)) is not supported."
  echo "============================================================================================"
  SUPPORTED="false"
fi

if [ "${SUPPORTED}" == "true" ]; then
  echo "Generating configuration file..."
  echo "
<source>
  @type tail
  path "${DEBUG_LOG_FILE_PATH}"
  pos_file "${DEBUG_LOG_FILE_PATH}".pos
  tag cyphonic.${SERVICE_TYPE,,}.debug
  <parse>
    @type json
  </parse>
</source>

<source>
  @type tail
  path "${ERROR_LOG_FILE_PATH}"
  pos_file "${ERROR_LOG_FILE_PATH}".pos
  tag cyphonic.${SERVICE_TYPE,,}.error
  <parse>
    @type json
  </parse>
</source>

<match cyphonic.**>
  @type elasticsearch
  host elasticsearch
  port 9200
  logstash_format true
  logstash_prefix cyphonic
</match>" > ./scripts/fluentd/fluent.conf
fi

#!/bin/bash

set -u

SCRIPT_DIR=$( cd -- "$( dirname -- "${BASH_SOURCE[0]}" )" &> /dev/null && pwd )

source ${SCRIPT_DIR}/setup.sh
cd ${SCRIPT_DIR}/../terraform
terraform apply
if [[ $? != 0 ]]; then
  echo "terraform apply failed, not continuing"
  exit 1
fi

ip=$(terraform output -raw ip_address)
cd ${SCRIPT_DIR}/../
make build
check_ssh ${ip}
scp bin/math-visual-proofs-server root@${ip}:

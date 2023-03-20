#!/bin/bash

set -u

SCRIPT_DIR=$( cd -- "$( dirname -- "${BASH_SOURCE[0]}" )" &> /dev/null && pwd )

source ./setup.sh
cd ${SCRIPT_DIR}/terraform
terraform apply
if [[ $? != 0 ]]; then
  echo "terraform apply failed, not continuing"
  exit
fi

ip=$(terraform output -raw ip_address)
cd ../
echo ${ip} #| tee /tmp/hosts

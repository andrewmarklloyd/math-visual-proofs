#!/bin/bash

set -u

SCRIPT_DIR=$( cd -- "$( dirname -- "${BASH_SOURCE[0]}" )" &> /dev/null && pwd )

export TF_VAR_ssh_inbound_ip=$(curl -s ifconfig.me)
eval $(echo "export AWS_SECRET_ACCESS_KEY=op://math-visual-proofs/do-spaces-keys/SPACES_AWS_SECRET_ACCESS_KEY
export SPACES_URL=op://math-visual-proofs/do-spaces-keys/SPACES_URL
export AWS_ACCESS_KEY_ID=op://math-visual-proofs/do-spaces-keys/SPACES_AWS_ACCESS_KEY_ID
export TF_VAR_do_token=op://math-visual-proofs/digital-ocean/token
export CLOUDMQTT_SERVER_URL=op://math-visual-proofs/cloud-mqtt/CLOUDMQTT_SERVER_URL" | op inject)

cd ${SCRIPT_DIR}/../terraform/
tfenv install
tfenv use
terraform init

check_ssh() {
  ip=${1}
  success='false'
  echo "Checking for ssh access"
  until [ ${success} == 'true' ]; do
    ssh root@${ip} exit
    code=$(echo $?)
    if [[ ${code} == 0 ]]; then
      success='true'
    else
      echo "exit code: ${code}"
      sleep 5
    fi
  done
}

check_docker() {
  ip=${1}
  success='false'
  echo "Checking for docker running"
  until [ ${success} == 'true' ]; do
    ssh root@${ip} docker ps
    code=$(echo $?)
    if [[ ${code} == 0 ]]; then
      success='true'
    else
      echo "exit code: ${code}"
      sleep 5
    fi
  done
}

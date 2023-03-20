#!/bin/bash

set -eu

get_config() {
  vault=${1}
  op item get --vault ${vault} "cloud-mqtt" --fields type=concealed --format json
}

post() {
  endpoint=${1}
  payload=${2}
  curl -s -XPOST -u :${CLOUDMQTT_APIKEY} \
    -d "${payload}" \
    -H "Content-Type:application/json" https://api.cloudmqtt.com/api/${endpoint}
}

create_agent_user() {
  post user "{\"username\": \"${CLOUDMQTT_MATH_PROOFS_AGENT_USER}\",\"password\": \"${CLOUDMQTT_MATH_PROOFS_AGENT_PASSWORD}\"}"
  post acl "{\"type\":\"topic\",\"username\":\"${CLOUDMQTT_MATH_PROOFS_AGENT_USER}\",\"pattern\":\"math-visual-proofs/#\",\"read\":false,\"write\":true}"
}

create_server_user() {
  post user "{\"username\": \"${CLOUDMQTT_MATH_PROOFS_SERVER_USER}\",\"password\": \"${CLOUDMQTT_MATH_PROOFS_SERVER_PASSWORD}\"}"
  post acl "{\"type\":\"topic\",\"username\":\"${CLOUDMQTT_MATH_PROOFS_SERVER_USER}\",\"pattern\":\"math-visual-proofs/#\",\"read\":true,\"write\":false}"
}


config=$(get_config math-visual-proofs)

fields="CLOUDMQTT_MATH_PROOFS_SERVER_USER
CLOUDMQTT_MATH_PROOFS_SERVER_PASSWORD
CLOUDMQTT_MATH_PROOFS_AGENT_USER
CLOUDMQTT_MATH_PROOFS_AGENT_PASSWORD"
for f in ${fields}; do
  export ${f}=$(echo ${config} | jq -r ".[] | select(.label==\"${f}\").value")
done

create_server_user
create_agent_user

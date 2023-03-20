#!/bin/bash


sudo apt update
sudo apt install jq apt-transport-https ca-certificates curl software-properties-common awscli -y
curl -fsSL https://download.docker.com/linux/ubuntu/gpg | sudo gpg --dearmor -o /usr/share/keyrings/docker-archive-keyring.gpg
echo "deb [arch=$(dpkg --print-architecture) signed-by=/usr/share/keyrings/docker-archive-keyring.gpg] https://download.docker.com/linux/ubuntu $(lsb_release -cs) stable" | sudo tee /etc/apt/sources.list.d/docker.list > /dev/null
sudo apt update
apt-cache policy docker-ce
sudo apt install docker-ce -y

# Create user and immediately delete password
USERNAME=math-visual-proofs
useradd --create-home --shell "/bin/bash" --groups sudo,docker "${USERNAME}"
passwd --delete "${USERNAME}"

# Create SSH directory for sudo user and move keys over
home_directory="$(eval echo ~${USERNAME})"
mkdir --parents "${home_directory}/.ssh"
cp /root/.ssh/authorized_keys "${home_directory}/.ssh"
chmod 0700 "${home_directory}/.ssh"
chmod 0600 "${home_directory}/.ssh/authorized_keys"
chown --recursive "${USERNAME}":"${USERNAME}" "${home_directory}/.ssh"

# Disable root SSH login with password
sed --in-place 's/^PermitRootLogin.*/PermitRootLogin prohibit-password/g' /etc/ssh/sshd_config
if sshd -t -q; then systemctl restart sshd; fi

# install docker compose
mkdir -p /home/${USERNAME}/.docker/cli-plugins/
curl -SL https://github.com/docker/compose/releases/download/v2.3.3/docker-compose-linux-x86_64 -o /home/${USERNAME}/.docker/cli-plugins/docker-compose
chmod +x /home/${USERNAME}/.docker/cli-plugins/docker-compose
chown --recursive "${USERNAME}":"${USERNAME}" "${home_directory}/.docker"

docker pull manimcommunity/manim:stable

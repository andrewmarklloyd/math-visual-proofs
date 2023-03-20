resource "digitalocean_droplet" "renderer_server" {
  image  = "ubuntu-22-04-x64"
  name   = "renderer-server"
  region = "sfo3"
  monitoring = true
  # is this the smallest size?
  size   = "s-1vcpu-1gb"
  ssh_keys = [data.digitalocean_ssh_key.do.id]

  user_data = data.local_file.userdata.content
  tags = ["math-visual-proofs"]
}

output "ip_address" {
  value = digitalocean_droplet.renderer_server.ipv4_address
}

output "droplet_id" {
  value = digitalocean_droplet.renderer_server.id
}

data "local_file" "userdata" {
  filename = "./userdata.sh"
}

data "digitalocean_ssh_key" "do" {
  name = "DO SSH Key"
}

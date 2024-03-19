#!/bin/bash 

# ========================================================================
# Setup script for greenlight server.
#
# Usage:
#
# Copy this script to the server:
#   rsync -rP --delete ./remote/setup root@<IP-ADDRESS>:/root
#
# Use SSH to run the command, using the using the -t flag to force 
# pseudo-terminal allocation.
#
#   ssh -t root@<IP-ADDRESS> "bash /root/setup/01.sh"
#
# ========================================================================


# Exit immediately if any command returns a non-zero exit status, and treat
# unset variables as an error.
set -eu

# ========================================================================
# VARIABLES
# ========================================================================

# Set timezone for the server. A full list of available timezones can be found
# by running `timedatectl list-timezones`.
TIMEZONE=America/New_York

# Set username for DB user and prompt for password.
USERNAME=greenlight
read -p "Enter password for greenlight DB user: " DB_PASSWORD

# Force all output to be presented in en_US for the duration of this script. 
# This avoids any "setting locale failed" errors while this script is running, # before we have  installed support for all locales. Do not change this setting.
export LC_ALL=en_US.UTF-8

# ========================================================================
# INITIAL SETUP
# ========================================================================

# Enable universe and update software packages.
add-apt-repository --yes universe
apt update

# Set the system timezone and install all locales.
timedatectl set-timezone ${TIMEZONE}
apt --yes install locales-all

# Add the new user to the server and give sudo privileges.
# useradd --create-home --shell "/bin/bash" --groups sudo "${USERNAME}"

# Force a password to be set for the new user the first time they log in.
passwd --delete "${USERNAME}"
chage --lastday 0 "${USERNAME}"

# Copy SSH keys from the root user of the server to the new user, and change 
# their permissions. 
rsync --archive --chown=${USERNAME}:${USERNAME} /root/.ssh /home/${USERNAME}

# Configure the firewall to allow SSH, HTTP, and HTTPS traffic.
ufw allow 22
ufw allow 80/tcp
ufw allow 443/tcp
ufw --force enable

# Install fail2ban, which will automatically and temporarily ban an IP address
# if it makes too many failed SSH login attempts.
apt --yes install fail2ban

# ========================================================================
# DATABASE SETUP
# ========================================================================

# Install migrate CLI tool.
curl -L https://github.com/golang-migrate/migrate/releases/download/v4.14.1/migrate.linux-amd64.tar.gz | tar xvz
mv migrate.linux-amd64 /usr/local/bin/migrate

# Install PostgreSQL.
apt --yes install postgresql

# Set up the greenlight DB and create a user with the DB_PASSWORD as password.
sudo -i -u postgres psql -c "CREATE DATABASE greenlight"
sudo -i -u postgres psql -d greenlight -c "CREATE EXTENSION IF NOT EXISTS citext"
sudo -i -u postgres psql -d greenlight -c "CREATE ROLE greenlight WITH LOGIN PASSWORD '${DB_PASSWORD}'"

# Add DSN for connecting to greenlight database. This variable will be available
# as a system-wide environmental variable, stored in /etc/environment.
echo "GREENLIGHT_DB_DSN='postgres://greenlight:${DB_PASSWORD}@localhost/greenlight'" >> /etc/environment

# Install Caddy reverse proxy. This apparently starts the service automatically.
# See https://caddyserver.com/docs/install#debian-ubuntu-raspbian.
apt install -y debian-keyring debian-archive-keyring apt-transport-https
curl -1sLf 'https://dl.cloudsmith.io/public/caddy/stable/gpg.key' | sudo gpg --dearmor -o /usr/share/keyrings/caddy-stable-archive-keyring.gpg
curl -1sLf 'https://dl.cloudsmith.io/public/caddy/stable/debian.deb.txt' | sudo tee /etc/apt/sources.list.d/caddy-stable.list
apt update
apt --yes install caddy

# Upgrade all packages. Using the --force-confnew flag will replace config files
# with new ones if they are availabe.
apt --yes -o Dpkg::Options::="--force-confnew" upgrade

echo "Script complete! Rebooting..."
reboot
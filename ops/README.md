# Website Deployment

Automated deployment system for the Sol Armada website using systemd and GitHub releases.

## Overview

This deployment system automatically:
- Fetches the latest release from GitHub
- Verifies binary integrity with SHA256 checksums
- Installs and restarts the service
- Rolls back automatically on failure
- Keeps the 3 most recent releases

## Files

- `updater/website-updater.sh` - Update script that fetches and installs releases
- `../systemd/website.service` - Main website systemd service
- `../systemd/website-updater.service` - Update service
- `../systemd/website-updater.timer` - Timer to check for updates every 5 minutes

## Initial Setup

### 1. Create User and Directories

```bash
sudo useradd -r -s /bin/false website
sudo mkdir -p /var/lib/website /etc/website /opt/website-releases
sudo chown website:website /var/lib/website /opt/website-releases
sudo chmod 755 /var/lib/website /opt/website-releases
```

### 2. Install Updater Script

```bash
sudo cp ops/updater/website-updater.sh /usr/local/bin/
sudo chmod +x /usr/local/bin/website-updater.sh
```

### 3. Install Systemd Units

```bash
sudo cp systemd/website.service /etc/systemd/system/
sudo cp systemd/website-updater.service /etc/systemd/system/
sudo cp systemd/website-updater.timer /etc/systemd/system/
sudo systemctl daemon-reload
```

### 4. Configure Environment

```bash
sudo cp website.env.example /etc/website/website.env
sudo chown root:website /etc/website/website.env
sudo chmod 640 /etc/website/website.env
sudo nano /etc/website/website.env  # Edit with your configuration
```

### 5. Enable and Start Services

```bash
# Enable services to start on boot
sudo systemctl enable website.service
sudo systemctl enable website-updater.timer

# Start the updater timer (will check for updates)
sudo systemctl start website-updater.timer

# Manually trigger first update to download and install
sudo systemctl start website-updater.service

# Check status
sudo systemctl status website.service
sudo journalctl -u website -f
```

## Usage

### Manual Update

```bash
sudo systemctl start website-updater.service
```

### Check Updater Logs

```bash
sudo journalctl -u website-updater.service -n 50
```

### View Website Logs

```bash
sudo journalctl -u website -f
```

### Restart Website

```bash
sudo systemctl restart website.service
```

### Check Current Version

```bash
cat /var/lib/website/current-version
```

### List Available Releases

```bash
ls -la /opt/website-releases/
```

## Rollback

If an update fails, the system automatically rolls back. To manually rollback:

```bash
# Install previous version
sudo install -m 755 /opt/website-releases/v1.0.0/website /opt/website

# Restart service
sudo systemctl restart website

# Update version file
echo "v1.0.0" | sudo tee /var/lib/website/current-version
```

## Architecture Override

For ARM64 servers:

```bash
# Create updater environment file
echo "ARCH=arm64" | sudo tee /etc/website/updater.env
```

## How It Works

1. **Timer triggers** every 5 minutes (with 45s random delay)
2. **Updater checks** GitHub API for latest release tag
3. **If new version found:**
   - Downloads binary and SHA256 checksum
   - Verifies checksum
   - Installs to `/opt/website-releases/{version}/`
   - Copies to `/opt/website`
   - Restarts service
   - Waits 8 seconds for health check
   - If restart or health check fails → automatic rollback
4. **Cleanup:** Removes old releases (keeps latest 3)

## Security Features

- Runs as dedicated `website` user
- Systemd hardening (NoNewPrivileges, ProtectHome, etc.)
- File locking prevents concurrent updates
- SHA256 checksum verification
- Automatic rollback on failure
- Rate limiting via timer randomization

## Troubleshooting

### Service Won't Start

```bash
# Check detailed status
sudo systemctl status website.service
sudo journalctl -u website -n 100 --no-pager

# Check environment file
sudo cat /etc/website/website.env

# Check binary permissions
ls -la /opt/website

# Test binary directly
sudo -u website /opt/website
```

### Updater Fails

```bash
# Check updater logs
sudo journalctl -u website-updater.service -n 100

# Verify dependencies
command -v curl jq sha256sum systemctl flock

# Test GitHub API access
curl -fsSL https://api.github.com/repos/Sol-Armada/website/releases/latest
```

### No Updates Happening

```bash
# Check timer status
sudo systemctl status website-updater.timer

# Check when timer will fire next
sudo systemctl list-timers | grep website

# Manually trigger update
sudo systemctl start website-updater.service
```

## Configuration

### Environment Variables

See `website.env.example` for all available configuration options.

Required variables:
- `JWT_SECRET`
- `DISCORD_CLIENT_ID`
- `DISCORD_CLIENT_SECRET`
- `DISCORD_GUILD_ID`
- `DATABASE_DSN`

### Updater Overrides

Optional overrides in `/etc/website/updater.env`:
- `REPO_OWNER` - GitHub organization (default: Sol-Armada)
- `REPO_NAME` - Repository name (default: website)
- `ARCH` - Architecture amd64/arm64 (default: amd64)
- `INSTALL_PATH` - Binary location (default: /opt/website)

## Monitoring

### Check Update Timer

```bash
sudo systemctl status website-updater.timer
sudo journalctl -u website-updater.timer -f
```

### Check Service Health

```bash
sudo systemctl is-active website.service
curl http://localhost:8080/health
```

### View All Logs

```bash
sudo journalctl -u website -u website-updater.service -f
```

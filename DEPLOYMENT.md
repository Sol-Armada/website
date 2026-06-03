# Quick Deployment Guide

## What Was Created

### Deployment Scripts
- ✅ `ops/updater/website-updater.sh` - Auto-update script
- ✅ `ops/updater/updater.env.example` - Updater config example
- ✅ `ops/README.md` - Comprehensive deployment documentation

### Systemd Service Files
- ✅ `systemd/website.service` - Main service
- ✅ `systemd/website-updater.service` - Update service
- ✅ `systemd/website-updater.timer` - Auto-update timer (every 5 min)

### Configuration
- ✅ `website.env.example` - Environment variables template

## Quick Start (Production Server)

```bash
# 1. Setup user and directories
sudo useradd -r -s /bin/false website
sudo mkdir -p /var/lib/website /etc/website /opt/website-releases
sudo chown website:website /var/lib/website /opt/website-releases

# 2. Install updater
sudo cp ops/updater/website-updater.sh /usr/local/bin/
sudo chmod +x /usr/local/bin/website-updater.sh

# 3. Install systemd services
sudo cp systemd/*.service systemd/*.timer /etc/systemd/system/
sudo systemctl daemon-reload

# 4. Configure environment
sudo cp website.env.example /etc/website/website.env
sudo chown root:website /etc/website/website.env
sudo chmod 640 /etc/website/website.env
sudo nano /etc/website/website.env  # EDIT THIS with real values

# 5. Enable and start
sudo systemctl enable website.service website-updater.timer
sudo systemctl start website-updater.timer
sudo systemctl start website-updater.service  # First update

# 6. Verify
sudo systemctl status website.service
sudo journalctl -u website -f
```

## How It Works

1. **GitHub Release Workflow** (`.github/workflows/release-binary.yml`)
   - Builds frontend → `api/dist/`
   - Builds Go binary with embedded frontend
   - Creates release with linux/amd64 and linux/arm64 binaries

2. **Auto-Update System** (sol-bot style)
   - Timer checks GitHub every 5 minutes
   - Downloads latest release + SHA256 checksum
   - Verifies integrity
   - Installs to `/opt/website-releases/{version}/`
   - Restarts service
   - Auto-rollback on failure

3. **Production Deployment**
   - Single binary at `/opt/website`
   - Runs as `website` user
   - Environment from `/etc/website/website.env`
   - Logs to journald

## Manual Operations

```bash
# Manual update
sudo systemctl start website-updater.service

# Check version
cat /var/lib/website/current-version

# View logs
sudo journalctl -u website -f
sudo journalctl -u website-updater.service -n 50

# Restart
sudo systemctl restart website.service

# Rollback to v1.0.0
sudo install -m 755 /opt/website-releases/v1.0.0/website /opt/website
sudo systemctl restart website
echo "v1.0.0" | sudo tee /var/lib/website/current-version
```

## Required Environment Variables

Must be set in `/etc/website/website.env`:

- `JWT_SECRET` - Session signing key
- `DISCORD_CLIENT_ID` - OAuth client ID
- `DISCORD_CLIENT_SECRET` - OAuth client secret
- `DISCORD_GUILD_ID` - Discord server ID
- `DISCORD_REDIRECT_URI` - OAuth callback URL
- `DATABASE_DSN` - PostgreSQL connection string
- `REDIS_ADDR` - Redis server address
- `ADMIN_ROLE_ID` - Discord admin role ID

See `website.env.example` for all options.

## Testing The Workflow

1. Commit all changes
2. Create and push a test tag:
   ```bash
   git tag v0.0.1
   git push origin v0.0.1
   ```
3. Watch GitHub Actions build
4. Verify release artifacts attached
5. On server, manually trigger update:
   ```bash
   sudo systemctl start website-updater.service
   ```

## Architecture Support

- **AMD64**: Default, automatically detected
- **ARM64**: Set `ARCH=arm64` in `/etc/website/updater.env`

Both architectures are built by the GitHub workflow.

## Documentation

See `ops/README.md` for comprehensive documentation including:
- Detailed setup instructions
- Troubleshooting guide
- Security features
- Monitoring procedures
- Configuration options

---

**Status**: ✅ Deployment system ready!  
**Pattern**: Matches sol-bot deployment exactly  
**Features**: Auto-update, rollback, health checks, release retention

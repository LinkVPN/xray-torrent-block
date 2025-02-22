# Torrent Blocker

[![en](https://img.shields.io/badge/lang-en-red)](https://github.com/LinkVPN/xray-torrent-block/blob/main/README.md)
[![ru](https://img.shields.io/badge/lang-ru-blue)](https://github.com/LinkVPN/xray-torrent-block/blob/main/README.ru.md)

Torrent Blocker is an application designed to block torrent usage by users of the  panel. The application analyzes logs, detects torrent activity, and temporarily blocks the user, sending notifications to the administrator via Telegram, and optionally to the user.
It can also work with other panels and directly with Xray.

## Features:

- Monitoring logs of nodes and the panel for torrent usage.
- IP address blocking at the system level. Maximum block speed (no abuse reports!).
- Sending notifications via Telegram to both the administrator and the user.
- Configurable through a configuration file.
- Uses UFW for blocking.
- Configurable block duration.
- Supports temporary blocking with automatic unblocking.
- Simple installation and setup via systemd.

## Preparation

### Xray Configuration

- Enable logging. Section `log`
  ```json
    "log": {
      "access": "/var/lib/marzban-node/access.log",
      "error": "/var/lib/marzban-node/error.log",
      "loglevel": "error",
      "dnsLog": false
    },
  ```
- Configure bittorrent traffic tagging. Section `routing`. Add the rule:

  ```json
        {
          "protocol": [
            "bittorrent"
          ],
          "outboundTag": "TORRENT",
          "type": "field"
        },
  ```

  Here, `TORRENT` is the tag that the application will use to filter logs.

- Configure bittorrent traffic blocking. Section `outbounds`. Send all traffic to blackhole:
  ```json
      {
        "protocol": "blackhole",
        "tag": "TORRENT"
      },
  ```
  Unfortunately, this blocking only effectively handles about 20% of bittorrent traffic.

###  Configuration

- On the server where the panel is hosted, create the folder `/var/lib/-node`:

  ```bash
  mkdir -p /var/lib/-node
  ```

- Add a new volume to the `/opt//docker-compose.yml` file:

  ```yaml
  volumes:
    - /var/lib/:/var/lib/
    - /var/lib/-node:/var/lib/-node #новый volume
  ```

- Restart the panel with the following command:
  ```bash
  docker compose down --remove-orphans; docker compose up -d
  ```

### Node Configuration

Ensure that the volume is correctly mounted in `docker-compose.yml`:

```yaml
volumes:
  - /var/lib/-node:/var/lib/-node
```

By default, this volume is present, ensuring logs are accessible on the host.

## Installation

To automatically install the application, follow these steps:

- Run the installation script:
  ```bash
  bash <(curl -fsSL https://raw.githubusercontent.com/LinkVPN/xray-torrent-block/main/install.sh)
  ```
- The script will automatically install all dependencies, download the latest release, ask for the admin `Token` and `Chat ID`, and start the service.
- After installation, you can control the application via systemd:
  ```bash
  systemctl start/status/stop torrent-blocker
  ```

## Configuration

After installation, you can configure the application's behavior via the configuration file located at: `/opt/torrent-blocker/config.yaml`.

Key configuration parameters:

- **LogFile** — the path to the log file to be monitored. Default: `/var/lib/marzban-node/access.log`
- **BlockDuration** — the duration of the user's block in minutes. Default: `10`
- **TorrentTag** — the tag used to identify log entries related to torrents. Default: `TORRENT`
- **SendAdminMessage** — whether to send notifications to the administrator. Optional. Default: `true`
- **AdminBotToken** — the admin bot token for notifications. Optional.
- **AdminChatID** — the admin chat (or user) ID to which notifications will be sent. Optional.
- **SendUserMessage** — whether to send notifications to the user. Optional. Default: `false`
- **BotToken** — the bot token for sending messages to the user via Telegram. Optional.
- **TidRegex** — regular expression to extract the user's `CHAT_ID` from the log entry. Optional.
- **UserMessageTemplate** — the message template for notifying the user. Optional.
- **UsernameRegex** — regular expression to extract the user's login from the log entry. Optional.

An example configuration file with detailed comments is available at `/opt/torrent-blocker/config.yaml`.

### Example for Sending Notifications to Users:

If the user's `CHAT_ID` is included in their login on the Marzban panel, you can configure the application to send notifications directly to the user.

For example, if the user's login in Marzban looks like this: `kutovoys_tgid-1234111`, you can set up the following in `config.yaml`:

- **TidRegex**: `tgid-(\\d+)`
- **UsernameRegex**: `email: \\d+\\.(\\w+)_tgid-`

In this case, the administrator will receive notifications with the username `kutovoys`, and the user will also be notified directly via Telegram when they are blocked.

## Contributing

We welcome contributions from the community! If you have ideas for improvements or have found a bug, please create an issue or submit a pull request on GitHub.

<p align="center">
  <img width=100 src="https://github.com/kn-lim/seigetsu-bot/blob/main/images/seigetsu.png?raw=true"></img>
</p>

# seigetsu-bot

![Go](https://img.shields.io/github/go-mod/go-version/kn-lim/seigetsu-bot)
[![Release](https://img.shields.io/github/v/release/kn-lim/seigetsu-bot)](https://github.com/kn-lim/seigetsu-bot/releases)
![Build](https://github.com/kn-lim/seigetsu-bot/actions/workflows/release.yaml/badge.svg)
[![Go Report Card](https://goreportcard.com/badge/github.com/kn-lim/seigetsu-bot)](https://goreportcard.com/report/github.com/kn-lim/seigetsu-bot)
![License](https://img.shields.io/github/license/kn-lim/seigetsu-bot)

A personal Discord bot to handle miscellaneous tasks.

## Packages Used

- [discordgo](https://github.com/bwmarrin/discordgo/)
- [aws-sdk-go-v2](https://github.com/aws/aws-sdk-go-v2/)

## Current Uses

- Manages a Pixelmon Minecraft server

## Environment Variables

| Name | Description |
|-|-|
| `DISCORD_BOT_TOKEN` | Discord Bot Token |
| `RCON_PASSWORD` | RCON Password of Pixelmon Service |
| `PIXELMON_NAME` | AWS Name Tag of Pixelmon EC2 Instance |
| `PIXELMON_INSTANCE_ID` | AWS Instance ID of Pixelmon EC2 Instance |
| `PIXELMON_REGION` | AWS Region of Pixelmon EC2 Instance |
| `PIXELMON_HOSTED_ZONE_ID` | AWS Hosted Zone ID of Domain |
| `PIXEMON_DOMAIN` | Domain of Pixelmon Server |
| `PIXEMON_SUBDOMAIN` | Subdomain of Pixelmon Server |

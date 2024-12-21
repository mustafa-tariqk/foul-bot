# FoulBot

Me and my friends assign points anytime someone does something "foul." We need a way to track and assign these points to determine who is the "foul sport of the year."

## Prerequisites

Auth:

* [Discord Guild ID](https://en.wikipedia.org/wiki/Template:Discord_server#Getting_Guild_ID)
* Discord App ID
* Discord Bot Token

Hardware:

* Internet
* Random computer you can leave on forever / as much as you can

## Discord Specific Setup

Watch [this video](https://youtu.be/Oy5HGvrxM4o) if you don't know how to get an App ID or Bot Token.

The scopes you need to generate an invite link via oauth2 generator are: `bot` + `application.commands`

I have been setting the role permissions to `Administrator` mainly because I'm too lazy to think about it anymore.

## Installation

1. Download latest release executable
2. Create an empty `config.json` file
3. Fill it with the Guild ID, App ID and Bot Token. Reference below.
4. Set the executable to run on startup

## config.json

```json
{
    "DISCORD_APPLICATION_ID": "",
    "DISCORD_GUILD_ID": "",
    "DISCORD_TOKEN": ""
}
```

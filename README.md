# FoulBot

Me and my friends assign points anytime someone does something "foul." We need a way to track and assign these points to determine who is the "foul sport of the year."

## Prerequisites

Software:

* [Git](https://git-scm.com/downloads)
* [Golang](https://go.dev/doc/install)
* [Make](https://www.gnu.org/software/make/#download)

Auth:

* [Discord Guild ID](https://en.wikipedia.org/wiki/Template:Discord_server#Getting_Guild_ID)
* Discord App ID
* Discord Bot Token

Hardware:

* Internet
* Random computer you can leave on forever
* (Cloud server also works)

## Discord Specific Setup

Watch [this video](https://youtu.be/Oy5HGvrxM4o) if you don't know how to get an App ID or Bot Token.

The scopes you need to generate an invite link via oauth2 generator are: `bot` + `application.commands`

I have been setting the role permissions to `Administrator` mainly because I'm too lazy to think about it anymore.

## Installation

1. Clone the repo: `git clone https://github.com/mustafa-tariqk/foul-bot.git`
2. Create an empty .env: `touch .env`
3. Fill it with the Guild ID, App ID and Bot Token as follows. Reference below.
4. Run the app with `make run` or `go run .` if not on a unix system.

## .env

```sh
DISCORD_TOKEN=
DISCORD_GUILD_ID=
DISCORD_APPLICATION_ID=
```

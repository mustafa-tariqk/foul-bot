package main

import (
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"

	"github.com/bwmarrin/discordgo"
	"github.com/joho/godotenv"
	_ "github.com/mattn/go-sqlite3"
)

func main() {
	godotenv.Load()
	guildId := os.Getenv("GUILD_ID")
	token := os.Getenv("DISCORD_TOKEN")

	bot, err := discordgo.New("Bot " + token)
	if err != nil {
		log.Fatal(err)
	}

	bot.AddHandler(func(s *discordgo.Session, i *discordgo.InteractionCreate) {
		if i.Type == discordgo.InteractionApplicationCommand {
			switch i.ApplicationCommandData().Name {
			case "own":
				// Retrieve the command options
				options := i.ApplicationCommandData().Options
				user := options[0].UserValue(s)
				number := options[1].IntValue()
				reason := options[2].StringValue()

				// Format the response message
				response := fmt.Sprintf("Own <@%s> %+d for %s", user.ID, number, reason)

				// Create buttons
				owningButton := discordgo.Button{
					Label:    "Owning",
					Style:    discordgo.PrimaryButton,
					CustomID: "owning",
				}
				notOwningButton := discordgo.Button{
					Label:    "Not Owning",
					Style:    discordgo.SecondaryButton,
					CustomID: "not_owning",
				}

				// Create action row
				actionRow := discordgo.ActionsRow{
					Components: []discordgo.MessageComponent{
						owningButton,
						notOwningButton,
					},
				}

				// Create interaction response data
				responseData := discordgo.InteractionResponseData{
					Content:    response,
					Components: []discordgo.MessageComponent{actionRow},
				}

				// Create interaction response
				interactionResponse := discordgo.InteractionResponse{
					Type: discordgo.InteractionResponseChannelMessageWithSource,
					Data: &responseData,
				}

				// Respond to the interaction
				s.InteractionRespond(i.Interaction, &interactionResponse)
			}
		}
	})

	bot.Identify.Intents = discordgo.IntentsAllWithoutPrivileged

	err = bot.Open()
	if err != nil {
		log.Fatal(err)
	}
	defer bot.Close()

	if err != nil {
		log.Fatal(err)
	}

	_, err = bot.ApplicationCommandCreate(bot.State.User.ID, guildId, &discordgo.ApplicationCommand{
		Name:        "own",
		Description: "Displays ownership",
		Options: []*discordgo.ApplicationCommandOption{
			{
				Type:        discordgo.ApplicationCommandOptionUser,
				Name:        "user",
				Description: "The user to mention",
				Required:    true,
			},
			{
				Type:        discordgo.ApplicationCommandOptionInteger,
				Name:        "number",
				Description: "An integer value",
				Required:    true,
			},
			{
				Type:        discordgo.ApplicationCommandOptionString,
				Name:        "reason",
				Description: "The reason for gaining",
				Required:    true,
			},
		},
	})
	if err != nil {
		log.Fatal(err)
	}

	fmt.Println("Bot is running")

	sc := make(chan os.Signal, 1)
	signal.Notify(sc, syscall.SIGINT, syscall.SIGTERM, os.Interrupt)
	<-sc
}

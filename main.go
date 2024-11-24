package main

import (
	"fmt"
	"log"
	"os"
	"os/signal"
	"sort"
	"syscall"

	"github.com/bwmarrin/discordgo"
	"github.com/joho/godotenv"
)

func main() {
	godotenv.Load()
	guildId := os.Getenv("DISCORD_GUILD_ID")
	token := os.Getenv("DISCORD_TOKEN")
	appId := os.Getenv("DISCORD_APPLICATION_ID")

	points := make(map[string]int64)

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

				// TODO: add to counter ~after~ poll
				points[user.ID] += number
			case "leaderboard":
				type userPoints struct {
					userID string
					points int64
				}
				pairs := make([]userPoints, 0, len(points))
				for id, score := range points {
					pairs = append(pairs, userPoints{id, score})
				}

				// Sort by points descending
				sort.Slice(pairs, func(i, j int) bool {
					return pairs[i].points > pairs[j].points
				})

				// Get top 10 or all if less
				if len(pairs) > 10 {
					pairs = pairs[:10]
				}

				// Number emojis for ranking
				numbers := []string{":one:", ":two:", ":three:", ":four:", ":five:", ":six:", ":seven:", ":eight:", ":nine:", ":ten:"}

				description := ""
				for i, pair := range pairs {
					user, err := s.User(pair.userID)
					if err != nil {
						fmt.Printf("Error fetching user: %v\n", err)
						continue
					}
					description += fmt.Sprintf("%s %s: %d\n", numbers[i], user.Username, pair.points)
				}

				embed := &discordgo.MessageEmbed{
					Title:       "Leaderboard",
					Description: description,
				}

				s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
					Type: discordgo.InteractionResponseChannelMessageWithSource,
					Data: &discordgo.InteractionResponseData{
						Embeds: []*discordgo.MessageEmbed{embed},
					},
				})
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

	commands := []*discordgo.ApplicationCommand{
		{
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
		},
		{
			Name:        "leaderboard",
			Description: "Displays the leaderboard",
			Options:     []*discordgo.ApplicationCommandOption{},
		},
	}

	_, err = bot.ApplicationCommandBulkOverwrite(appId, guildId, commands)
	if err != nil {
		log.Fatalf("could not register commands: %s", err)
	}

	fmt.Println("Bot is running")

	sc := make(chan os.Signal, 1)
	signal.Notify(sc, syscall.SIGINT, syscall.SIGTERM, os.Interrupt)
	<-sc
}

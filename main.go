package main

import (
	"encoding/json"
	"fmt"
	"log"
	"os"
	"os/signal"
	"sort"
	"syscall"

	"github.com/bwmarrin/discordgo"
	"github.com/joho/godotenv"
)

var (
	POINTS_JSON            = "points.json"
	DISCORD_TOKEN          = "DISCORD_TOKEN"
	DISCORD_GUILD_ID       = "DISCORD_GUILD_ID"
	DISCORD_APPLICATION_ID = "DISCORD_APPLICATION_ID"
)

func loadPoints() map[string]int64 {
	points := make(map[string]int64)

	if _, err := os.Stat(POINTS_JSON); os.IsNotExist(err) {
		data, err := json.Marshal(points)
		if err != nil {
			log.Fatalf("could not marshal empty points: %s", err)
		}
		if err := os.WriteFile(POINTS_JSON, data, 0644); err != nil {
			log.Fatalf("could not create points file: %s", err)
		}
	} else {
		data, err := os.ReadFile(POINTS_JSON)
		if err != nil {
			log.Fatalf("could not read points file: %s", err)
		}
		if err := json.Unmarshal(data, &points); err != nil {
			log.Fatalf("could not unmarshal points: %s", err)
		}
	}

	return points
}

func loadEnv() (*discordgo.Session, map[string]int64, string, string) {
	godotenv.Load()
	bot, err := discordgo.New("Bot " + os.Getenv(DISCORD_TOKEN))
	if err != nil {
		log.Fatal(err)
	}
	points := loadPoints()

	guildId := os.Getenv(DISCORD_GUILD_ID)
	appId := os.Getenv(DISCORD_APPLICATION_ID)

	return bot, points, guildId, appId
}

func savePoints(points map[string]int64) {
	data, err := json.Marshal(points)
	if err != nil {
		log.Fatalf("could not marshal points: %s", err)
	}
	if err := os.WriteFile(POINTS_JSON, data, 0644); err != nil {
		log.Fatalf("could not write points file: %s", err)
	}
}

func establishCommands(bot *discordgo.Session, guildId string, appId string) {
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

	_, err := bot.ApplicationCommandBulkOverwrite(appId, guildId, commands)
	if err != nil {
		log.Fatalf("could not register commands: %s", err)
	}

	bot.Identify.Intents = discordgo.IntentsAllWithoutPrivileged
}

func create_leaderboard(points map[string]int64, s *discordgo.Session) *discordgo.MessageEmbed {
	type userPoints struct {
		userID string
		points int64
	}
	pairs := make([]userPoints, 0, len(points))
	for id, score := range points {
		pairs = append(pairs, userPoints{id, score})
	}

	sort.Slice(pairs, func(i, j int) bool {
		return pairs[i].points > pairs[j].points
	})

	if len(pairs) > 10 {
		pairs = pairs[:10]
	}

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

	return embed
}

func main() {
	bot, points, guildId, appId := loadEnv()

	bot.AddHandler(func(s *discordgo.Session, i *discordgo.InteractionCreate) {
		if i.Type == discordgo.InteractionApplicationCommand {
			options := i.ApplicationCommandData().Options
			switch i.ApplicationCommandData().Name {
			case "own":
				user := options[0].UserValue(s)
				number := options[1].IntValue()
				reason := options[2].StringValue()

				fmt.Printf("Own <@%s> %+d for %s\n", user.ID, number, reason)

				// TODO: add to counter ~after~ poll
				points[user.ID] += number
				savePoints(points)
			case "leaderboard":
				s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
					Type: discordgo.InteractionResponseChannelMessageWithSource,
					Data: &discordgo.InteractionResponseData{
						Embeds: []*discordgo.MessageEmbed{create_leaderboard(points, s)},
					},
				})
			}
		}
	})

	err := bot.Open()
	if err != nil {
		log.Fatal(err)
	}
	defer bot.Close()

	establishCommands(bot, guildId, appId)
	fmt.Println("Bot is running...")

	// Runs forever, shut down with ctrl+c
	sc := make(chan os.Signal, 1)
	signal.Notify(sc, syscall.SIGINT, syscall.SIGTERM, os.Interrupt)
	<-sc
}

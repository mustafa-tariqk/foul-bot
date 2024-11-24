package main

import (
	"encoding/json"
	"fmt"
	"log"
	"os"
	"os/signal"
	"sort"
	"sync"
	"syscall"
	"time"

	"github.com/bwmarrin/discordgo"
	"github.com/joho/godotenv"
)

var (
	POINTS_JSON            = "points.json"
	DISCORD_TOKEN          = "DISCORD_TOKEN"
	DISCORD_GUILD_ID       = "DISCORD_GUILD_ID"
	DISCORD_APPLICATION_ID = "DISCORD_APPLICATION_ID"
	activePolls            = make(map[string]*VotePoll)
	pollsMutex             sync.RWMutex
)

type VotePoll struct {
	MessageID string
	ChannelID string
	UserID    string
	Points    int64
	Reason    string
	ExpiresAt time.Time
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

				// Create poll message
				err := s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
					Type: discordgo.InteractionResponseChannelMessageWithSource,
					Data: &discordgo.InteractionResponseData{
						Content: fmt.Sprintf("Own giving <@%s> %+d points for %s\nPoll closes in 1 hour",
							user.ID, number, reason),
					},
				})
				if err != nil {
					return
				}
				pollMsg, err := s.InteractionResponse(i.Interaction)
				if err != nil {
					return
				}

				s.MessageReactionAdd(i.ChannelID, pollMsg.ID, "👍")
				s.MessageReactionAdd(i.ChannelID, pollMsg.ID, "👎")

				// Create poll
				poll := &VotePoll{
					MessageID: pollMsg.ID,
					ChannelID: i.ChannelID,
					UserID:    user.ID,
					Points:    number,
					Reason:    reason,
					ExpiresAt: time.Now().Add(1 * time.Hour),
				}

				// Store poll
				pollsMutex.Lock()
				activePolls[pollMsg.ID] = poll
				pollsMutex.Unlock()

				time.AfterFunc(1*time.Hour, func() {
					concludePoll(s, poll, points)
				})
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

	bot.AddHandler(func(s *discordgo.Session, r *discordgo.MessageReactionAdd) {
		pollsMutex.RLock()
		_, exists := activePolls[r.MessageID]
		pollsMutex.RUnlock()

		if !exists || r.UserID == s.State.User.ID {
			return
		}

		// TODO: removing duplicates?
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

func concludePoll(s *discordgo.Session, poll *VotePoll, points map[string]int64) {
	pollsMutex.Lock()
	delete(activePolls, poll.MessageID)
	pollsMutex.Unlock()

	// Get reactions
	upVotes, _ := s.MessageReactions(poll.ChannelID, poll.MessageID, "👍", 100, "", "")
	downVotes, _ := s.MessageReactions(poll.ChannelID, poll.MessageID, "👎", 100, "", "")

	// Compare votes (excluding bot's reaction)
	if len(upVotes)-1 > len(downVotes)-1 {
		points[poll.UserID] += poll.Points
		savePoints(points)
		s.ChannelMessageSend(poll.ChannelID,
			fmt.Sprintf("<@%s> gaining %+d points", poll.UserID, poll.Points))
	} else {
		s.ChannelMessageSend(poll.ChannelID,
			fmt.Sprintf("<@%s> not gaining", poll.UserID))
	}
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

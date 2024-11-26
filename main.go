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
	VERSION                = "dev"
	DISCORD_TOKEN          = "DISCORD_TOKEN"
	DISCORD_GUILD_ID       = "DISCORD_GUILD_ID"
	DISCORD_APPLICATION_ID = "DISCORD_APPLICATION_ID"
	POINTS_JSON            = "points.json"
	POLLS_JSON             = "polls.json"
	POLL_LENGTH            = 24 * time.Hour
	pollsMutex             sync.RWMutex
	activePolls            = make(map[string]*VotePoll)
	NUMBERS                = []string{":one:", ":two:", ":three:", ":four:", ":five:", ":six:", ":seven:", ":eight:", ":nine:", ":ten:"}
)

type userPoints struct {
	userID string
	points int64
}

type VotePoll struct {
	MessageID string
	ChannelID string
	UserID    string
	Points    int64
	Reason    string
	ExpiresAt time.Time
}

type StoredPolls struct {
	Polls map[string]*VotePoll `json:"polls"`
}

func main() {
	bot, points, guildId, appId := loadEnv()

	handleInputs(bot, points)
	handlePollsMutex(bot)

	err := bot.Open()
	if err != nil {
		log.Fatal(err)
	}
	defer bot.Close()

	handleExpiredPolls(bot, points)

	establishCommands(bot, guildId, appId)
	fmt.Println("Bot is running...")

	// Runs forever, shut down with ctrl+c
	sc := make(chan os.Signal, 1)
	signal.Notify(sc, syscall.SIGINT, syscall.SIGTERM, os.Interrupt)
	<-sc
	fmt.Println("Bot is shutting down...")
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

func loadPolls() map[string]*VotePoll {
	polls := make(map[string]*VotePoll)
	if _, err := os.Stat(POLLS_JSON); os.IsNotExist(err) {
		savePolls(polls)
		return polls
	}

	data, err := os.ReadFile(POLLS_JSON)
	if err != nil {
		log.Fatalf("could not read polls file: %s", err)
	}

	var storedPolls StoredPolls
	if err := json.Unmarshal(data, &storedPolls); err != nil {
		log.Fatalf("could not unmarshal polls: %s", err)
	}

	return storedPolls.Polls
}

func savePolls(polls map[string]*VotePoll) {
	storedPolls := StoredPolls{Polls: polls}
	data, err := json.Marshal(storedPolls)
	if err != nil {
		log.Fatalf("could not marshal polls: %s", err)
	}
	if err := os.WriteFile(POLLS_JSON, data, 0644); err != nil {
		log.Fatalf("could not write polls file: %s", err)
	}
}

func handleExpiredPolls(bot *discordgo.Session, points map[string]int64) {
	polls := loadPolls()
	now := time.Now()

	for _, poll := range polls {
		if now.After(poll.ExpiresAt) {
			concludePoll(bot, poll, points)
		} else {
			pollsMutex.Lock()
			activePolls[poll.MessageID] = poll
			pollsMutex.Unlock()

			timeUntilExpiry := time.Until(poll.ExpiresAt)
			time.AfterFunc(timeUntilExpiry, func() {
				concludePoll(bot, poll, points)
			})
		}
	}
}

func handleInputs(bot *discordgo.Session, points map[string]int64) {
	bot.AddHandler(func(s *discordgo.Session, i *discordgo.InteractionCreate) {
		if i.Type == discordgo.InteractionApplicationCommand {
			options := i.ApplicationCommandData().Options
			switch i.ApplicationCommandData().Name {
			case "own":
				user := options[0].UserValue(s)
				number := options[1].IntValue()
				reason := options[2].StringValue()

				err := s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
					Type: discordgo.InteractionResponseChannelMessageWithSource,
					Data: &discordgo.InteractionResponseData{
						Content: fmt.Sprintf("Own <@%s> %+d points for %s\nPoll closes in 24 hours",
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

				poll := &VotePoll{
					MessageID: pollMsg.ID,
					ChannelID: i.ChannelID,
					UserID:    user.ID,
					Points:    number,
					Reason:    reason,
					ExpiresAt: time.Now().Add(POLL_LENGTH),
				}

				pollsMutex.Lock()
				activePolls[pollMsg.ID] = poll
				savePolls(activePolls)
				pollsMutex.Unlock()

				time.AfterFunc(POLL_LENGTH, func() {
					concludePoll(s, poll, points)
				})
			case "leaderboard":
				s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
					Type: discordgo.InteractionResponseChannelMessageWithSource,
					Data: &discordgo.InteractionResponseData{
						Embeds: []*discordgo.MessageEmbed{create_leaderboard(points, s)},
					},
				})
			case "version":
				s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
					Type: discordgo.InteractionResponseChannelMessageWithSource,
					Data: &discordgo.InteractionResponseData{
						Content: fmt.Sprintf("Current version: %s", VERSION),
					},
				})
			}
		}
	})
}

func concludePoll(s *discordgo.Session, poll *VotePoll, points map[string]int64) {
	pollsMutex.Lock()
	delete(activePolls, poll.MessageID)
	savePolls(activePolls)
	pollsMutex.Unlock()

	upVotes, _ := s.MessageReactions(poll.ChannelID, poll.MessageID, "👍", 100, "", "")
	downVotes, _ := s.MessageReactions(poll.ChannelID, poll.MessageID, "👎", 100, "", "")

	if len(upVotes) > len(downVotes) {
		points[poll.UserID] += poll.Points
		savePoints(points)
		s.ChannelMessageSend(poll.ChannelID,
			fmt.Sprintf("<@%s> gaining %+d", poll.UserID, poll.Points))
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
	return create_leaderboard_string(pairs, s)
}

func create_leaderboard_string(pairs []userPoints, s *discordgo.Session) *discordgo.MessageEmbed {
	description := ""
	for i, pair := range pairs {
		user, err := s.User(pair.userID)
		if err != nil {
			log.Fatal(err)
			continue
		}
		description += fmt.Sprintf("%s %s: %d\n", NUMBERS[i], user.Username, pair.points)
	}
	return &discordgo.MessageEmbed{
		Title:       "Leaderboard",
		Description: description,
	}
}

func handlePollsMutex(bot *discordgo.Session) {
	bot.AddHandler(func(s *discordgo.Session, r *discordgo.MessageReactionAdd) {
		pollsMutex.RLock()
		_, exists := activePolls[r.MessageID]
		pollsMutex.RUnlock()

		if !exists || r.UserID == s.State.User.ID {
			return
		}

		// TODO: removing duplicates?
	})
}

func establishCommands(bot *discordgo.Session, guildId string, appId string) {
	commands := []*discordgo.ApplicationCommand{
		{
			Name:        "own",
			Description: "Accuse someone of gaining",
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
			Description: "Displays a top 10 leaderboard",
			Options:     []*discordgo.ApplicationCommandOption{},
		},
		{
			Name:        "version",
			Description: "Displays the current version",
			Options:     []*discordgo.ApplicationCommandOption{},
		},
	}
	_, err := bot.ApplicationCommandBulkOverwrite(appId, guildId, commands)
	if err != nil {
		log.Fatalf("could not register commands: %s", err)
	}
	bot.Identify.Intents = discordgo.IntentsAllWithoutPrivileged
}

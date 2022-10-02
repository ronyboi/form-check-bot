package main

import (
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"

	"github.com/bwmarrin/discordgo"
	"github.com/joho/godotenv"
)

// Variables used for command line parameters
var (
	Token string
	dg *discordgo.Session
)

func init() {
	err := godotenv.Load("discordToken.env")
	
	if err != nil {
		log.Fatalf("Error loading .env file")
	}

	Token = os.Getenv("TOKEN")

	// Create a new Discord session using the provided bot token.
	dg, err = discordgo.New("Bot " + Token)
	if err != nil {
		fmt.Println("error creating Discord session,", err)
		return
	}
}

// Commands 
var (
	commands = []*discordgo.ApplicationCommand{
		{
			Name: "form-check",
			Description: "Command to initiate a new form check",
			Options: []*discordgo.ApplicationCommandOption{
				{
					Type: discordgo.ApplicationCommandOptionString,
					Name: "description",
					Description: "Describe your lift, what felt strong? what felt weak? Any other comments?",
					Required: true,
				},
				{
					Type: discordgo.ApplicationCommandOptionAttachment,
					Name: "video",
					Description: "Attach a clip of your lift",
					Required: true,
				},
			},
		},
	}

	commandHandlers = map[string]func(s *discordgo.Session, i *discordgo.InteractionCreate){
		"form-check": func(s *discordgo.Session, i *discordgo.InteractionCreate) {
			options := i.ApplicationCommandData().Options

			optionMap := make(map[string]*discordgo.ApplicationCommandInteractionDataOption, len(options))
			for _, opt := range options {
				optionMap[opt.Name] = opt
			}

			attachments := i.ApplicationCommandData().Resolved.Attachments;

			var attachment *discordgo.MessageAttachment;

			for key, value := range attachments {
				log.Printf("%s %s", key, value.URL)
				attachment = value
			}

			s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
				Type: discordgo.InteractionResponseChannelMessageWithSource,
				Data: &discordgo.InteractionResponseData{
					Content: "Creating a form check thread...",
				},
			})

			invokee := i.Interaction.Member.User.Username
			invokeeID := i.Interaction.Member.User.ID

			threadHeader := fmt.Sprintf("Form Check for %s", invokee)
			messageCreator := fmt.Sprintf("**Description**\n%s\n<@%s>\nLink to the clip of a lift: \n%s", optionMap["description"].StringValue(), invokeeID, attachment.URL)

			message, _ := s.ChannelMessageSend(i.ChannelID, messageCreator);

			// genEmbed := embed.NewEmbed().SetTitle(threadHeader).SetDescription(messageCreator).SetThumbnail(attachment.URL, attachment.ProxyURL).SetColor(0x1c1c1c)

			// genEmbed.Video = &discordgo.MessageEmbedVideo{
			// 	URL: attachment.ProxyURL,
			// 	Height: attachment.Height,
			// 	Width: attachment.Width,
			// }

			// message, _ := s.ChannelMessageSendEmbed(i.ChannelID, genEmbed.MessageEmbed)
			s.MessageThreadStart(i.ChannelID, message.ID, threadHeader, 60)
		},
	}
)

func init() {
	dg.AddHandler(func(s *discordgo.Session, i *discordgo.InteractionCreate)  {
			if h, ok := commandHandlers[i.ApplicationCommandData().Name]; ok {
				h(s, i)
			}
		})
}

func main() {
	dg.AddHandler(func(s *discordgo.Session, r *discordgo.Ready) {
		log.Printf("Logged in as: %v#%v", s.State.User.Username, s.State.User.Discriminator)
	})
	// In this example, we only care about receiving message events.
	dg.Identify.Intents = discordgo.IntentsGuildMessages

	// Open a websocket connection to Discord and begin listening.
	err := dg.Open()
	if err != nil {
		fmt.Println("error opening connection,", err)
		return
	}

	log.Println("Adding commands...")
	registeredCommands := make([]*discordgo.ApplicationCommand, len(commands))
	for i, v := range commands {
		cmd, err := dg.ApplicationCommandCreate(dg.State.User.ID, "", v)
		if err != nil {
			log.Panicf("Cannot create '%v' command: %v", v.Name, err)
		}
		registeredCommands[i] = cmd
	}

	// Wait here until CTRL-C or other term signal is received.
	fmt.Println("Bot is now running.  Press CTRL-C to exit.")
	sc := make(chan os.Signal, 1)
	signal.Notify(sc, syscall.SIGINT, syscall.SIGTERM, os.Interrupt, os.Kill)
	<-sc

	// Cleanly close down the Discord session.
	defer dg.Close()

	log.Println("Removing commands...")
		// // We need to fetch the commands, since deleting requires the command ID.
		// // We are doing this from the returned commands on line 375, because using
		// // this will delete all the commands, which might not be desirable, so we
		// // are deleting only the commands that we added.
		// registeredCommands, err := s.ApplicationCommands(s.State.User.ID, *GuildID)
		// if err != nil {
		// 	log.Fatalf("Could not fetch registered commands: %v", err)
		// }

	for _, v := range registeredCommands {
		err := dg.ApplicationCommandDelete(dg.State.User.ID, "", v.ID)
		if err != nil {
			log.Panicf("Cannot delete '%v' command: %v", v.Name, err)
		}
	}
}
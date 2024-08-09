package main

import (
	"context"
	"encoding/binary"
	"fmt"
	"io"
	"log"
	"os"
	"os/exec"
	"os/signal"
	"strings"
	"syscall"
	"time"

	"github.com/LucasSnatiago/LMusicBot/config"
	"github.com/LucasSnatiago/LMusicBot/music"
	"github.com/bwmarrin/discordgo"
	"github.com/lrstanley/go-ytdlp"
)

var (
	BotConfig  *config.BotConfig
	StopSignal chan os.Signal
)

func init() {
	// If yt-dlp isn't installed yet, download and cache it for further use.
	ytdlp.MustInstall(context.TODO(), nil)
	BotConfig = config.New()
}

func main() {
	session, err := discordgo.New("Bot " + BotConfig.DiscordToken)
	if err != nil {
		log.Fatal("Bot could not run:", err)
	}

	session.Identify.Intents |= discordgo.IntentsGuildMessages
	session.Identify.Intents |= discordgo.IntentDirectMessages
	session.Identify.Intents |= discordgo.IntentsGuildVoiceStates

	session.AddHandler(LoginMessage)
	session.AddHandler(HandleMessages)

	// Open a websocket connection to Discord and begin listening.
	err = session.Open()
	if err != nil {
		fmt.Println("error opening connection:", err)
		return
	}

	// Graceful shutdown
	fmt.Println("Use CTRL + C to stop the Bot")
	StopSignal = make(chan os.Signal, 1)
	signal.Notify(StopSignal, syscall.SIGINT, syscall.SIGTERM, os.Interrupt)
	<-StopSignal

	session.Close()
	fmt.Println("Bot is off")
}

func LoginMessage(s *discordgo.Session, r *discordgo.Ready) {
	log.Printf("Logged in as %s", r.User.String())

	// Set the playing status.
	s.UpdateGameStatus(0, fmt.Sprintf("%shelp", BotConfig.BotPrefix))
}

func HandleMessages(s *discordgo.Session, m *discordgo.MessageCreate) {
	// Ignore all messages created by the bot itself and ignore all messages without the bot prefix
	if m.Author.ID == s.State.User.ID && !strings.HasPrefix(m.Content, BotConfig.BotPrefix) {
		return
	}

	// Get the Guild ID
	channel, err := s.Channel(m.ChannelID)
	if err != nil {
		log.Printf("Failed to get channel: %v", err)
		return
	}
	guildID := channel.GuildID

	// From now on, remove bot prefix
	command := strings.Replace(m.Content, BotConfig.BotPrefix, "", 1)
	commands := strings.Split(command, " ")

	switch commands[0] {
	case "ping":
		s.ChannelMessageSend(m.ChannelID, "Pong!")
	case "play":
		search_term := strings.Join(commands[1:], " ")
		go playSong(search_term, m.ChannelID, guildID, s, m)
	case "help":
		s.ChannelMessageSend(m.ChannelID, "Available commands:\n\nping: bot answers pong!\nhelp: show this message")
	case "shutdown":
		StopSignal <- syscall.SIGTERM
	}
}

func playSong(search, channelID, guildID string, s *discordgo.Session, m *discordgo.MessageCreate) {
	music_byte_array := music.GetSongByteArray(search)

	// Connecting to the user voice chat
	vc, err := s.ChannelVoiceJoin(guildID, getUserVoiceChannelID(s, m), false, true)
	if err != nil {
		s.ChannelMessageSend(channelID, "Failed to connect to the voice chat")
	}
	defer vc.Disconnect()

	// Sleep for a specified amount of time before playing the sound
	time.Sleep(250 * time.Millisecond)

	err = sendAudioToVoiceChannel(vc, music_byte_array)
	if err != nil {
		s.ChannelMessageSend(channelID, "Failed to send audio to the voice chat")
		log.Print(err)
	}

	// Sleep for a specificed amount of time before ending.
	time.Sleep(250 * time.Millisecond)
}

func sendAudioToVoiceChannel(vc *discordgo.VoiceConnection, audioBuffer io.Reader) error {
	opusEncoder := exec.Command("ffmpeg", "-i", "pipe:0", "-f", "s16le", "-ar", "48000", "-ac", "2", "pipe:1")
	opusEncoder.Stdin = audioBuffer

	opusData, err := opusEncoder.StdoutPipe()
	if err != nil {
		return fmt.Errorf("failed to create ffmpeg pipe: %w", err)
	}

	err = opusEncoder.Start()
	if err != nil {
		return fmt.Errorf("failed to start ffmpeg process: %w", err)
	}

	vc.Speaking(true)
	defer vc.Speaking(false)

	var opuslen int16
	for {
		err := binary.Read(opusData, binary.LittleEndian, &opuslen)
		if err != nil {
			log.Print("Could not read audio buffer:", err)
		}

		// Create a dinamic buffer
		buf := make([]byte, opuslen)
		err = binary.Read(opusData, binary.LittleEndian, &buf)
		if err != nil {
			if err == io.EOF || err == io.ErrUnexpectedEOF {
				break
			}
			return fmt.Errorf("error reading opus data: %w", err)
		}

		vc.OpusSend <- buf
	}

	return opusEncoder.Wait()
}

func getUserVoiceChannelID(s *discordgo.Session, m *discordgo.MessageCreate) string {
	// Verifying if the message author is in a voice channel
	vs, err := s.State.VoiceState(m.GuildID, m.Author.ID)
	if err != nil {
		log.Printf("Failed to get voice instance: %v", err)
		return ""
	}

	// If message author is not in any voice channel
	if vs == nil || vs.ChannelID == "" {
		log.Printf("User %s is not in a voice channel.", m.Author.Username)
		s.ChannelMessageSend(m.ChannelID, fmt.Sprintf("User %s is not in a voice channel.", m.Author.Username))
		return ""
	}

	// Get voice channel information
	channel, err := s.State.Channel(vs.ChannelID)
	if err != nil {
		log.Printf("Failed to get channel info: %v", err)
		return ""
	}

	return channel.ID
}

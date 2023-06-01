package muzikas

import (
	"fmt"
	"io"
	"log"
	"strings"
	"time"

	"github.com/bwmarrin/discordgo"
	"github.com/jonas747/dca"
)

type BotStatus int32

const (
	Resting BotStatus = 0
	Playing           = 1
	Paused            = 2
	Err               = 3
)

type MuzikasBot struct {
	session       *discordgo.Session
	voiceConn     *discordgo.VoiceConnection
	queue         chan *Song
	queueList     []string
	skipInterrupt chan bool
	currentStream *dca.StreamingSession
	botStatus     BotStatus
}

func NewMuzikasBot(session *discordgo.Session) *MuzikasBot {
	return &MuzikasBot{
		session:       session,
		queue:         make(chan *Song, 100),
		skipInterrupt: make(chan bool, 1),
		botStatus:     Resting,
	}
}

func (mb *MuzikasBot) Start() {
	log.Println(`
  ♫♫♫♫♫♫♫♫♫♫♫♫♫♫♫♫♫♫♫♫♫♫♫♫♫♫♫♫♫♫
  ♫♫♫♫♫♫♫♫STARTING MUZIKAS♫♫♫♫♫♫
  ♫♫♫♫♫♫♫♫♫♫♫♫♫♫♫♫♫♫♫♫♫♫♫♫♫♫♫♫♫♫
  `)
	mb.session.AddHandler(messageCreatedHandler(mb))
	mb.session.Open()
}

func messageCreatedHandler(mb *MuzikasBot) func(*discordgo.Session, *discordgo.MessageCreate) {
	return func(s *discordgo.Session, m *discordgo.MessageCreate) {
		switch strings.Split(m.Message.Content, " ")[0] {
		case "!play":
			c, _ := s.State.Channel(m.Message.ChannelID)
			g, _ := s.State.Guild(c.GuildID)
			for _, vs := range g.VoiceStates {
				if vs.UserID == m.Message.Author.ID {
					if mb.voiceConn == nil {
						conn, err := mb.session.ChannelVoiceJoin(g.ID, vs.ChannelID, false, true)
						if err != nil {
							log.Println("Error connecting to voice channel")
							s.ChannelMessageSend(m.Message.ChannelID, "Error connecting to voice channel")
							log.Println(err.Error())
							return
						}
						mb.voiceConn = conn
					}
					song, err := GetSongInfo(strings.Split(m.Message.Content, " ")[1])
					if err != nil {
						log.Println("Error fetching song to play")
						s.ChannelMessageSend(m.Message.ChannelID, "Error fetching song to play")
						return
					}
					s.ChannelMessageSend(m.Message.ChannelID, fmt.Sprintf("Playing song: %v", song.name))
					mb.enqueue(song)
					mb.playSong()
				}
			}
		case "!skip":
			mb.skip()
		case "!list":
			var sngList strings.Builder
			for _, sng := range mb.queueList {
				sngList.WriteString(sng)
			}
			s.ChannelMessageSend(m.Message.ChannelID, sngList.String())
		case "!queue":
			song, err := GetSongInfo(strings.Split(m.Message.Content, " ")[1])
			if err != nil {
				log.Println("Error fetching song to enqueue")
				s.ChannelMessageSend(m.Message.ChannelID, "Error fetching song to enqueue")
				return
			}
			s.ChannelMessageSend(m.Message.ChannelID, fmt.Sprintf("Added song to queue: %v", song.name))
			mb.enqueue(song)
		case "!stop":
			s.ChannelMessageSend(m.Message.ChannelID, "Stopping music")
			mb.stop()
		case "!pause":
			mb.pause()
		case "!unpause":
			mb.unpause()
		default:
			log.Println("Unknown command")
		}
	}
}

func (mb *MuzikasBot) playSong() {
	song := mb.dequeue()

	log.Printf("Playing song: %v \n", song.name)

	options := dca.StdEncodeOptions
	options.RawOutput = true
	options.Bitrate = 96
	options.Application = "lowdelay"

	encodingSession, err := dca.EncodeFile(song.downloadUrl, options)
	if err != nil {
		log.Println("Error encoding from yt url")
		log.Println(err.Error())
		return
	}
	defer encodingSession.Cleanup()

	time.Sleep(250 * time.Millisecond)
	err = mb.voiceConn.Speaking(true)

	if err != nil {
		log.Println("Error connecting to discord voice")
		log.Println(err.Error())
	}

	done := make(chan error)
	stream := dca.NewStream(encodingSession, mb.voiceConn, done)
	mb.currentStream = stream
	log.Println("Created stream, waiting on finish or err")

	mb.botStatus = Playing

	select {
	case err := <-done:
		log.Println("Song done")
		if err != nil && err != io.EOF {
			mb.botStatus = Err
			log.Println(err.Error())
			return
		}
		mb.voiceConn.Speaking(false)
		break
	case _ = <-mb.skipInterrupt:
		log.Println("Song interrupted, stop playing")
		mb.voiceConn.Speaking(false)
		return
	}
	mb.voiceConn.Speaking(false)

	if len(mb.queue) == 0 {
		time.Sleep(250 * time.Millisecond)
		log.Println("Audio done")
		mb.stop()
		mb.botStatus = Resting
		return
	}

	time.Sleep(250 * time.Millisecond)
	log.Println("Play next in queue")
	go mb.playSong()
	return
}

func (mb *MuzikasBot) skip() {
	if len(mb.queue) == 0 {
		mb.stop()
	} else {
		if len(mb.skipInterrupt) == 0 {
			mb.skipInterrupt <- true
			mb.playSong()
		}
	}
}

func (mb *MuzikasBot) enqueue(song *Song) {
	log.Printf("Queueing song %v", song.name)
	songString := fmt.Sprintf("-- :%v \n", song.name)
	mb.queueList = append(mb.queueList, songString)
	mb.queue <- song
}

func (mb *MuzikasBot) dequeue() *Song {
	mb.queueList = mb.queueList[1:]
	return <-mb.queue
}

func (mb *MuzikasBot) stop() {
	mb.voiceConn.Disconnect()
	mb.voiceConn = nil
	mb.botStatus = Resting
}

func (mb *MuzikasBot) pause() {
	if mb.currentStream != nil {
		mb.currentStream.SetPaused(true)
		log.Println("Paused playback")
	}
}

func (mb *MuzikasBot) unpause() {
	if mb.currentStream != nil {
		mb.currentStream.SetPaused(false)
		log.Println("Unpaused playback")
	}
}

package youtubejoinplugin

import (
	"encoding/json"
	"errors"
	"log"
	"sync"

	"google.golang.org/api/youtube/v3"

	"github.com/iopred/bruxism"
)

// YouTubeJoinPlugin is a plugin that monitors channels, and when they go live, will join the service
type YouTubeJoinPlugin struct {
	sync.RWMutex
	ytLiveChannel *bruxism.YTLiveChannel
	liveVideoChan chan *youtube.Video
	Channels      map[string]bool
}

// Help returns a list of help strings that are printed when the user requests them.
func (p *YouTubeJoinPlugin) Help(bot *bruxism.Bot, service bruxism.Service, message bruxism.Message, detailed bool) []string {
	return nil
}

// Message handler.
func (p *YouTubeJoinPlugin) Message(bot *bruxism.Bot, service bruxism.Service, message bruxism.Message) {
	if service.IsBotOwner(message) || service.IsChannelOwner(message) && bruxism.MatchesCommand(service, "leave", message) {
		p.Unmonitor(message.Channel())
	}
}

// Load will load plugin state from a byte array.
func (p *YouTubeJoinPlugin) Load(bot *bruxism.Bot, service bruxism.Service, data []byte) error {
	if service.Name() != bruxism.YouTubeServiceName {
		panic("YouTubeJoin plugin only supports YouTube.")
	}

	if data != nil {
		if err := json.Unmarshal(data, p); err != nil {
			log.Println("Error loading data", err)
		}
	}

	for channel, _ := range p.Channels {
		p.ytLiveChannel.MonitorAll(channel, p.liveVideoChan)
	}

	go p.Run(bot, service)

	return nil
}

// Run will poll YouTube for channels going live and send messages.
func (p *YouTubeJoinPlugin) Run(bot *bruxism.Bot, service bruxism.Service) {
	p.RLock()
	lvc := p.liveVideoChan
	p.RUnlock()
	for {
		v := <-lvc
		p.RLock()
		if p.Channels[v.Snippet.ChannelId] && v.LiveStreamingDetails != nil && v.LiveStreamingDetails.ActiveLiveChatId != "" {
			service.(*bruxism.YouTube).JoinVideo(v)
		}
		p.RUnlock()
	}
}

func (p *YouTubeJoinPlugin) Monitor(channel string) error {
	p.Lock()
	defer p.Unlock()

	if p.Channels[channel] {
		return errors.New("already tracking that channel")
	}

	p.Channels[channel] = true

	return p.ytLiveChannel.MonitorAll(channel, p.liveVideoChan)
}

func (p *YouTubeJoinPlugin) Unmonitor(channel string) error {
	p.Lock()
	defer p.Unlock()

	if !p.Channels[channel] {
		return errors.New("not tracking that channel")
	}

	delete(p.Channels, channel)

	p.ytLiveChannel.UnmonitorAll(channel, p.liveVideoChan)

	service.(*bruxism.YouTube).LeaveAll(channel)
}

// Save will save plugin state to a byte array.
func (p *YouTubeJoinPlugin) Save() ([]byte, error) {
	return json.Marshal(p)
}

// Stats will return the stats for a plugin.
func (p *YouTubeJoinPlugin) Stats(bot *bruxism.Bot, service bruxism.Service, message bruxism.Message) []string {
	return nil
}

// Name returns the name of the plugin.
func (p *YouTubeJoinPlugin) Name() string {
	return "YouTubeJoin"
}

// New will create a new YouTubeLive plugin.
func New(ytLiveChannel *bruxism.YTLiveChannel) *YouTubeJoinPlugin {
	return &YouTubeJoinPlugin{
		ytLiveChannel: ytLiveChannel,
		liveVideoChan: make(chan *youtube.Video),
		Channels:      map[string]bool{},
	}
}

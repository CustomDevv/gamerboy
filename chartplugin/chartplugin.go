package chartplugin

import (
	"bytes"
	"fmt"
	"image/color"
	"log"
	"math/rand"
	"strings"

	"github.com/gonum/plot"
	"github.com/gonum/plot/plotter"
	"github.com/gonum/plot/plotutil"
	"github.com/iopred/bruxism"
)

type chartPlugin struct {
	bruxism.SimplePlugin
}

var randomDirection = []string{
	"up",
	"down",
	"flat",
}

var randomY = []string{
	"interest",
	"care",
	"success",
	"fail",
	"happiness",
	"sadness",
	"money",
}

var randomX = []string{
	"time",
	"releases",
	"days",
	"years",
}

func (p *chartPlugin) random(list []string) string {
	return list[rand.Intn(len(list))]
}

func (p *chartPlugin) randomChart(service bruxism.Service) string {
	ticks := ""
	if service.Name() == bruxism.DiscordServiceName {
		ticks = "`"
	}

	return fmt.Sprintf("%s%schart %s %s, %s%s", ticks, service.CommandPrefix(), p.random(randomDirection), p.random(randomY), p.random(randomX), ticks)
}

func (p *chartPlugin) helpFunc(bot *bruxism.Bot, service bruxism.Service, message bruxism.Message, detailed bool) []string {
	help := bruxism.CommandHelp(service, "chart", "<up|down|flat> <vertical message>, <horizontal message>", "Creates a chart trending in the desired direction.")

	if detailed {
		help = append(help, []string{
			"Examples:",
			bruxism.CommandHelp(service, "chart", "down interest, time", "Creates a chart showing declining interest over time")[0],
		}...)
	}

	return help
}

func (p *chartPlugin) messageFunc(bot *bruxism.Bot, service bruxism.Service, message bruxism.Message) {
	if service.IsMe(message) {
		return
	}

	if bruxism.MatchesCommand(service, "chart", message) {
		query, parts := bruxism.ParseCommand(service, message)
		if len(parts) == 0 {
			service.SendMessage(message.Channel(), fmt.Sprintf("Invalid chart eg: %s", p.randomChart(service)))
			return
		}

		start, end := 0.5, 0.5

		switch parts[0] {
		case "up":
			start, end = 0, 1
		case "down":
			start, end = 1, 0
		case "flat":
		case "straight":
		default:
			service.SendMessage(message.Channel(), fmt.Sprintf("Invalid chart direction. eg: %s", p.randomChart(service)))
			return
		}

		axes := strings.Split(query[len(parts[0]):], ",")
		if len(axes) != 2 {
			service.SendMessage(message.Channel(), fmt.Sprintf("Invalid chart axis labels eg: %s", p.randomChart(service)))
			return
		}

		pl, err := plot.New()
		if err != nil {
			service.SendMessage(message.Channel(), fmt.Sprintf("Error making chart, sorry! eg: %s", p.randomChart(service)))
			return
		}

		service.Typing(message.Channel())

		pl.Y.Label.Text = axes[0]
		pl.X.Label.Text = axes[1]

		num := 5 + rand.Intn(15)

		start *= float64(num)
		end *= float64(num)

		pts := make(plotter.XYs, num)
		for i := range pts {
			pts[i].X = float64(i) + rand.Float64()*0.5 - 0.2
			pts[i].Y = start + float64(end-start)/float64(num-1)*float64(i) + rand.Float64()*0.5 - 0.25
		}

		pl.X.Tick.Label.Color = color.Transparent
		pl.Y.Tick.Label.Color = color.Transparent

		pl.X.Min = -0.5
		pl.X.Max = float64(num) + 0.5

		pl.Y.Min = -0.5
		pl.Y.Max = float64(num) + 0.5

		err = plotutil.AddLinePoints(pl, pts)
		if err != nil {
			service.SendMessage(message.Channel(), fmt.Sprintf("Sorry %s, there was a problem creating your chart.", message.UserName()))
			return
		}

		w, err := pl.WriterTo(320, 240, "png")
		if err != nil {
			service.SendMessage(message.Channel(), fmt.Sprintf("Sorry %s, there was a problem creating your chart.", message.UserName()))
			return
		}

		b := &bytes.Buffer{}
		w.WriteTo(b)

		go func() {
			url, err := bot.UploadToImgur(b, "chart.png")
			if err == nil {
				if service.Name() == bruxism.DiscordServiceName {
					service.SendMessage(message.Channel(), fmt.Sprintf("Here's your chart <@%s>: %s", message.UserID(), url))
				} else {
					service.SendMessage(message.Channel(), fmt.Sprintf("Here's your chart %s: %s", message.UserName(), url))
				}
			} else {
				// If imgur failed and we're on Discord, try file send instead!
				if service.Name() == bruxism.DiscordServiceName {
					service.SendFile(message.Channel(), "comic.png", b)
					return
				}

				log.Println("Error uploading comic: ", err)
				service.SendMessage(message.Channel(), fmt.Sprintf("Sorry %s, there was a problem uploading the comic to imgur.", message.UserName()))
			}
		}()
	}
}

// New will create a new comic plugin.
func New() bruxism.Plugin {
	p := &chartPlugin{
		SimplePlugin: *bruxism.NewSimplePlugin("Chart"),
	}
	p.MessageFunc = p.messageFunc
	p.HelpFunc = p.helpFunc
	return p
}

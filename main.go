package main

import (
	"flag"
	"fmt"
	ui "github.com/gizak/termui/v3"
	"github.com/gizak/termui/v3/widgets"
	"log"
	"math"
	"net/http"
	"strconv"
	"time"
)

type Stats struct {
	Index        int
	Site         string
	Low          int64
	Avg          int64
	High         int64
	RequestCount int
	RequestSec   float64
}

var requestSec int
var startedAt time.Time
var callers int
var max bool

func main() {
	startedAt = time.Now()
	flag.IntVar(&requestSec, "r", 10, "Request a sec")
	flag.IntVar(&callers, "c", 5, "Request callers")
	flag.BoolVar(&max, "max", false, "Max requests")
	flag.Parse()
	sites := flag.Args()

	// If max is set just for now increase request/sec and callers to a lot
	if max {
		requestSec = 10000
		callers = 1000
	}

	c := make(chan Stats)
	for i, site := range sites {
		go pollSite(i, site, c)
	}

	if err := ui.Init(); err != nil {
		log.Fatalf("failed to initialize termui: %v", err)
	}
	defer ui.Close()

	table := widgets.NewTable()
	table.TextStyle = ui.NewStyle(ui.ColorWhite)
	table.SetRect(0, 0, 160, 10)
	var rows [][]string
	rows = append(rows, []string{
		"Site",
		"High",
		"Avg",
		"Low",
		"Requests",
		"Req/sec",
	})

	for _, _ = range sites {
		rows = append(rows, []string{
			"_",
			"_",
			"_",
			"_",
			"_",
			"_",
		})
	}

	go func() {
		for stat := range c {
			rows[stat.Index+1][0] = stat.Site
			rows[stat.Index+1][1] = strconv.FormatInt(stat.High, 10)
			rows[stat.Index+1][2] = strconv.FormatInt(stat.Avg, 10)
			rows[stat.Index+1][3] = strconv.FormatInt(stat.Low, 10)
			rows[stat.Index+1][4] = fmt.Sprintf("%d", stat.RequestCount)
			rows[stat.Index+1][5] = fmt.Sprintf("%f", math.Round(stat.RequestSec*100)/100)
			table.Rows = rows
			ui.Render(table)
		}
	}()

	uiEvents := ui.PollEvents()
	for {
		e := <-uiEvents
		switch e.ID {
		case "q", "<C-c>":
			return
		}
	}
}

func pollSite(index int, site string, outChan chan Stats) {
	c := make(chan int64)
	go startSiteCaller(site, c)
	var timeList []int64
	go func() {
		for _ = range time.Tick(time.Second) {
			l, a, h := getStats(timeList)
			length := len(timeList)
			outChan <- Stats{
				index,
				site,
				l,
				a,
				h,
				length,
				float64(length) / time.Since(startedAt).Seconds(),
			}
		}
	}()

	for t := range c {
		timeList = append(timeList, t)
	}
}

func startSiteCaller(url string, out chan int64) {
	in := make(chan string, 2)
	// Spawn site callers to handle call load
	for i := 1; i <= callers; i++ {
		go startCallWorker(in, out)
	}

	for _ = range time.Tick(time.Second / time.Duration(requestSec)) {
		in <- url
	}
}

func startCallWorker(ch chan string, out chan int64) {
	for url := range ch {
		t, err := getSiteRespTime(url)
		if err != nil {
			println(err.Error())
		} else {
			out <- t
		}
	}
}

func getSiteRespTime(url string) (int64, error) {
	start := time.Now()
	_, err := http.Get(url)
	if err != nil {
		return 0, err
	}
	respTime := time.Since(start)

	return respTime.Milliseconds(), nil
}

func getStats(times []int64) (int64, int64, int64) {
	var low int64 = 0
	var total int64 = 0
	var high int64 = 0
	for _, t := range times {
		if low > t || low == 0 {
			low = t
		}
		total = total + t
		if high < t || high == 0 {
			high = t
		}
	}

	var avg int64
	if len(times) > 0 {
		avg = total / int64(len(times))
	} else {
		avg = 0
	}

	return low, avg, high
}

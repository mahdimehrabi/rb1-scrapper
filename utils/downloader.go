package utils

import (
	"context"
	"fmt"
	"github.com/gocolly/colly"
	"golang.org/x/time/rate"
	"log/slog"
	"math/rand"
	"net/url"
	"rb-scrapper/entity"
	"strings"
	"sync"
	"time"
)

const (
	numWorkers       = 10000
	rateLimit        = 1000 //maximum 1000 image download per second (iran network cant reach maximum but its possible in a server with good resources and internet)
	downloadQueueCap = 100000
)

type Downloader interface {
	Download(pathsChan chan string)
}

type SearchEngine struct {
	Name       string
	SearchURL  string
	ResultAttr string
	Extractor  func(*colly.HTMLElement) string
}

var searchEngines = []SearchEngine{
	{"Google", "https://www.google.com/search?tbm=isch&q=%s", "img",
		func(element *colly.HTMLElement) string {
			return element.Attr("src")
		}},
	{"Bing", "https://www.bing.com/images/search?q=%s", "a.iusc", extractImageURLFromBing},
}

func extractImageURLFromBing(e *colly.HTMLElement) string {
	data := e.Attr("m")
	start := strings.Index(data, `"murl":"`)
	if start == -1 {
		return ""
	}
	start += len(`"murl":"`)
	end := strings.Index(data[start:], `"`)
	if end == -1 {
		return ""
	}
	return data[start : start+end]
}

type DownloadResizer struct {
	logger      *slog.Logger
	count       uint64
	targetCount uint64
	limiter     *rate.Limiter
	mtx         *sync.Mutex
	rand        *rand.Rand
	ctx         context.Context
	cancelCtx   context.CancelFunc
	resultChan  chan *entity.URL
	queries     []string
}

func NewDownloadResizer(targetCount uint64, slog *slog.Logger, queries []string) *DownloadResizer {
	s := rand.NewSource(time.Now().UnixNano())
	ctx, cancelFunc := context.WithCancel(context.Background())
	return &DownloadResizer{
		targetCount: targetCount,
		logger:      slog,
		limiter:     rate.NewLimiter(rate.Limit(rateLimit), rateLimit),
		mtx:         &sync.Mutex{},
		rand:        rand.New(rand.New(s)),
		ctx:         ctx,
		cancelCtx:   cancelFunc,
		queries:     queries,
	}
}

// Download: returns channel of file paths
func (d *DownloadResizer) Download(imageURLsChan chan *entity.URL) {
	d.resultChan = imageURLsChan

	c := colly.NewCollector(
		colly.Async(true),
	)
	c.AllowURLRevisit = true
	c.SetRequestTimeout(time.Second * 2)

loop:
	for {
		select {
		case <-d.ctx.Done():
			break loop
		default:
			query := d.queries[d.rand.Intn(len(d.queries))]
			engine := searchEngines[d.rand.Intn(len(searchEngines))]
			c.OnHTML(engine.ResultAttr, func(e *colly.HTMLElement) {
				imgURL := engine.Extractor(e)
				if imgURL != "" {
					d.resultChan <- &entity.URL{
						URL:   imgURL,
						Query: query,
					}
				}
			})
			searchURL := fmt.Sprintf(engine.SearchURL, url.QueryEscape(query))

			if err := c.Visit(searchURL); err != nil {
				d.logger.Error("error in sending request to %s error:%s", searchURL, err.Error())
				continue
			}
			c.Wait()
		}
	}
	close(d.resultChan)
	d.logger.Info("Finished downloading and processing images.")
	return
}

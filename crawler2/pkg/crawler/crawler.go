package crawler

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/pkg/errors"

	"github.com/mitasimo/gbgo_bestpractice/crawler2/pkg/parser"
)

type CrawlResult struct {
	Err error
	Msg string
}

type Crawler struct {
	sync.Mutex
	visited   map[string]string
	maxDepth  int
	depthLock sync.RWMutex // блокировка для maxDepth
}

func NewCrawler(maxDepth int) *Crawler {
	return &Crawler{
		visited:  make(map[string]string),
		maxDepth: maxDepth,
	}
}

// lesson1
func (c *Crawler) chechDepth(depth int) (r bool) {
	c.depthLock.RLock()
	r = depth <= c.maxDepth
	c.depthLock.RUnlock()
	return
}

// lesson1
func (c *Crawler) IncDepth(d int) {
	c.depthLock.Lock()
	c.maxDepth += d
	c.depthLock.Unlock()
}

// рекурсивно сканируем страницы
func (c *Crawler) Run(ctx context.Context, url string, results chan<- CrawlResult, depth int) {
	// просто для того, чтобы успевать следить за выводом программы, можно убрать :)
	time.Sleep(2 * time.Second)

	// проверяем что контекст исполнения актуален
	select {
	case <-ctx.Done():
		return

	default:
		// проверка глубины
		//if depth >= c.maxDepth {
		//	return
		//}
		// lesson 1
		if !c.chechDepth(depth) {
			return
		}

		page, err := parser.Parse(ctx, url)
		if err != nil {
			// ошибку отправляем в канал, а не обрабатываем на месте
			results <- CrawlResult{
				Err: errors.Wrapf(err, "parse page %s", url),
			}
			return
		}

		title := parser.PageTitle(ctx, page)
		links := parser.PageLinks(ctx, nil, page)

		// блокировка требуется, т.к. мы модифицируем мапку в несколько горутин
		c.Lock()
		c.visited[url] = title
		c.Unlock()

		// отправляем результат в канал, не обрабатывая на месте
		results <- CrawlResult{
			Err: nil,
			Msg: fmt.Sprintf("%s -> %s\n", url, title),
		}

		// рекурсивно ищем ссылки
		for link := range links {
			// если ссылка не найдена, то запускаем анализ по новой ссылке
			if c.checkVisited(link) {
				continue
			}

			go c.Run(ctx, link, results, depth+1)
		}
	}
}

func (c *Crawler) checkVisited(url string) bool {
	c.Lock()
	defer c.Unlock()

	_, ok := c.visited[url]
	return ok
}

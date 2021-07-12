package main

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/pkg/errors"
)

type crawlResult struct {
	err error
	msg string
}

type crawler struct {
	sync.Mutex
	visited   map[string]string
	maxDepth  int
	depthLock sync.RWMutex // блокировка для maxDepth
}

func newCrawler(maxDepth int) *crawler {
	return &crawler{
		visited:  make(map[string]string),
		maxDepth: maxDepth,
	}
}

// lesson1
func (c *crawler) chechDepth(depth int) (r bool) {
	c.depthLock.RLock()
	r = depth <= c.maxDepth
	c.depthLock.RUnlock()
	return
}

// lesson1
func (c *crawler) IncDepth(d int) {
	c.depthLock.Lock()
	c.maxDepth += d
	c.depthLock.Unlock()
}

// рекурсивно сканируем страницы
func (c *crawler) run(ctx context.Context, url string, results chan<- crawlResult, depth int) {
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

		page, err := parse(ctx, url)
		if err != nil {
			// ошибку отправляем в канал, а не обрабатываем на месте
			results <- crawlResult{
				err: errors.Wrapf(err, "parse page %s", url),
			}
			return
		}

		title := pageTitle(ctx, page)
		links := pageLinks(ctx, nil, page)

		// блокировка требуется, т.к. мы модифицируем мапку в несколько горутин
		c.Lock()
		c.visited[url] = title
		c.Unlock()

		// отправляем результат в канал, не обрабатывая на месте
		results <- crawlResult{
			err: nil,
			msg: fmt.Sprintf("%s -> %s\n", url, title),
		}

		// рекурсивно ищем ссылки
		for link := range links {
			// если ссылка не найдена, то запускаем анализ по новой ссылке
			if c.checkVisited(link) {
				continue
			}

			go c.run(ctx, link, results, depth+1)
		}
	}
}

func (c *crawler) checkVisited(url string) bool {
	c.Lock()
	defer c.Unlock()

	_, ok := c.visited[url]
	return ok
}

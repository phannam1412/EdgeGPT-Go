package CookieManager

import (
	"encoding/json"
	"io"
	"os"

	"github.com/gabriel-vasile/mimetype"
	"github.com/pavel-one/EdgeGPT-Go/internal/Logger"
)

var log = Logger.NewLogger("CookieManager")

type CookieItem struct {
	CurrentUsed int
	Path        string
	Json        []map[string]any
}

type Manager struct {
	Cookies []*CookieItem
}

func NewManager() (*Manager, error) {
	var cookies []*CookieItem

	//entries, err := os.ReadDir(cookiesPath)
	//if err != nil {
	//	return nil, err
	//}
	//
	//for _, e := range entries {
	//	if e.IsDir() {
	//		continue
	//	}
	//	filepath := cookiesPath + "/" + e.Name()
	filepath := os.Getenv("BINGCHAT_COOKIE")

	f, err := os.Open(filepath)
	if err != nil {
		return nil, err
	}
	cookiesJSON, err := io.ReadAll(f)
	if err != nil {
		return nil, err
	}

	m := mimetype.Detect(cookiesJSON)
	if !m.Is("application/json") {
		return nil, err
	}

	var parse []map[string]any
	err = json.Unmarshal(cookiesJSON, &parse)
	if err != nil {
		return nil, err
	}

	if err := f.Close(); err != nil {
		return nil, err
	}

	cookies = append(cookies, &CookieItem{
		CurrentUsed: 0,
		Path:        filepath,
		Json:        parse,
	})
	//}

	return &Manager{Cookies: cookies}, nil
}

func (m *Manager) GetBestCookie() []map[string]any {
	o := m.Cookies[m.findMinIndex()]

	log.Infoln("Getting new cookies:", o.Path)

	return o.Json
}

func (m *Manager) findMinIndex() (index int) {
	arr := m.Cookies

	for i, s := range arr {
		if s.CurrentUsed < arr[index].CurrentUsed {
			index = i
		}
	}

	m.Cookies[index].CurrentUsed++
	return index
}

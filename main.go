package main

import (
	"crypto/tls"
	"flag"
	"fmt"
	"log"
	"net/http"
	"net/url"
	"os"
	"time"

	"tg-gemini-bot/geminis"

	"context"

	"github.com/google/generative-ai-go/genai"
	"go.uber.org/zap"
	"google.golang.org/api/iterator"
	"google.golang.org/api/option"

	"golang.org/x/net/proxy"
	tele "gopkg.in/telebot.v3"
)

func buildClientWithProxy(addr string) (*http.Client, error) {
	auth := &proxy.Auth{
		User:     "",
		Password: "",
	}

	dialer, err := proxy.SOCKS5("tcp", string(addr), auth, proxy.Direct)
	if err != nil {
		return nil, err
	}

	httpTransport := &http.Transport{Dial: dialer.Dial}
	hc := &http.Client{Transport: httpTransport}

	return hc, nil
}

type APIKeyProxyTransport struct {
	APIKey    string
	Transport http.RoundTripper
	ProxyURL  string
}

func detectImageType(filename string) (string, error) {
	file, err := os.Open(filename)
	if err != nil {
		return "", err
	}
	defer file.Close()

	buffer := make([]byte, 512)
	_, err = file.Read(buffer)
	if err != nil {
		return "", err
	}

	if buffer[0] == 0xFF && buffer[1] == 0xD8 && buffer[2] == 0xFF {
		return "jpeg", nil
	} else if buffer[0] == 0x89 && buffer[1] == 0x50 && buffer[2] == 0x4E && buffer[3] == 0x47 {
		return "png", nil
	} else if buffer[0] == 0x47 && buffer[1] == 0x49 && buffer[2] == 0x46 && buffer[3] == 0x38 {
		return "gif", nil
	} else if buffer[1] == 0x50 && buffer[2] == 0x4E && buffer[3] == 0x47 {
		return "png", nil
	} else {
		return "Unknown", nil
	}
}
func (t *APIKeyProxyTransport) RoundTrip(req *http.Request) (*http.Response, error) {
	rt := t.Transport
	if rt == nil {
		rt = http.DefaultTransport
	}

	// 如果提供了 ProxyURL，则对 Transport 设置代理
	if t.ProxyURL != "" {
		proxyURL, err := url.Parse(t.ProxyURL)
		if err != nil {
			return nil, err
		}
		dialer, err := proxy.FromURL(proxyURL, proxy.Direct)
		if err != nil {
			return nil, err
		}
		rt = &http.Transport{
			Dial: dialer.Dial,
			TLSClientConfig: &tls.Config{
				InsecureSkipVerify: true,
			},
		}
	}

	// 克隆请求以避免修改原始请求
	newReq := *req
	args := newReq.URL.Query()
	args.Set("key", t.APIKey)
	newReq.URL.RawQuery = args.Encode()
	// 执行 HTTP 请求，并处理可能的错误
	resp, err := rt.RoundTrip(&newReq)
	if err != nil {
		// 返回网络请求中的错误
		return nil, fmt.Errorf("error during round trip: %v", err)
	}

	return resp, nil
}

func main() {
	var proxyaddr string
	var proxyurl string
	flag.StringVar(&proxyaddr, "proxy", "", "SOCKS5 proxy address")
	flag.Parse()
	var pref tele.Settings

	if proxyaddr != "" {
		client, err := buildClientWithProxy(proxyaddr)
		if err != nil {
			log.Fatal(err)
			return
		}
		pref = tele.Settings{
			Token:  os.Getenv("TELEBOT_TOKEN"),
			Poller: &tele.LongPoller{Timeout: 10 * time.Second},
			Client: client,
		}
	} else {
		pref = tele.Settings{
			Token:  os.Getenv("TELEBOT_TOKEN"),
			Poller: &tele.LongPoller{Timeout: 10 * time.Second},
		}
	}
	if proxyaddr != "" {
		proxyurl = "socks5://" + proxyaddr
	} else {
		proxyurl = ""
	}
	logger, err := zap.NewProduction()
	if err != nil {
		panic("Failed to create logger: " + err.Error())
	}
	defer logger.Sync()

	b, err := tele.NewBot(pref)
	if err != nil {
		zap.Error(err)
		return
	}
	b.Handle(tele.OnPhoto, func(c tele.Context) error {
		var (
			text  = c.Text()
			photo = c.Message().Photo
		)
		client1 := &http.Client{Transport: &APIKeyProxyTransport{
			APIKey:    os.Getenv("GEMINI_API_KEY"),
			Transport: nil,
			ProxyURL:  proxyurl,
		}}
		ctx := context.Background()
		client, err := genai.NewClient(ctx, option.WithHTTPClient(client1))
		if err != nil {
			logger.Fatal("Error while creating client", zap.Error(err))
		}
		defer client.Close()
		if err != nil {
			fmt.Println("解析JSON失败:", err)
		}
		//for _, photo := range  {
		b.Download(&photo.File, photo.FileID)
		imageType, err := detectImageType(photo.FileID)
		if err != nil {
			fmt.Println("Error:", err)
			return err
		}
		data, err := os.ReadFile(photo.FileID)
		if err != nil {
			logger.Fatal("", zap.Error(err))
		}
		geminis.GetPhotoResponse(c, client.GenerativeModel("gemini-1.5-pro-vision"), imageType, text, data)
		os.Remove(photo.FileID)
		//}
		//c.Send(resp, &tele.SendOptions{ParseMode: tele.ModeMarkdown})
		return err
	})

	b.Handle(tele.OnText, func(c tele.Context) error {
		client1 := &http.Client{Transport: &APIKeyProxyTransport{
			APIKey:    os.Getenv("GEMINI_API_KEY"),
			Transport: nil,
			ProxyURL:  proxyurl,
		}}
		ctx := context.Background()
		client, err := genai.NewClient(ctx, option.WithHTTPClient(client1))
		//client, err := genai.NewClient(ctx, option.WithAPIKey(os.Getenv("GEMINI_API_KEY")))
		if err != nil {
			logger.Fatal("Error while creating client", zap.Error(err))
		}
		defer client.Close()
		var (
			text = c.Text()
		)
		if c.Message().Photo == nil {
			geminis.GetTextResponse(c, client.GenerativeModel("models/gemini-1.5-pro"), text)
			client.ListModels(ctx)
			iter := client.ListModels(ctx)
			for {
				m, err := iter.Next()
				if err == iterator.Done {
					break
				}
				if err != nil {
					panic(err)
				}
				fmt.Println(m.Name, m.Description)
			}

			//c.Send(resp, &tele.SendOptions{ParseMode: tele.ModeMarkdown})
		}
		return err
	})
	b.Start()
}

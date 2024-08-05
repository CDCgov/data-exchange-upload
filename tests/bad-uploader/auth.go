package main

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	neturl "net/url"
	"sync"
	"time"

	"golang.org/x/oauth2"
)

type SAMSTokenSource struct {
	username string
	password string
	url      string
	token    *oauth2.Token
	lock     sync.Mutex
}

type SAMSToken struct {
	AccessToken  string   `json:"access_token"`
	TokenType    string   `json:"token_type"`
	ExpiresIn    int      `json:"expires_in"`
	RefreshToken string   `json:"refresh_token"`
	Scope        string   `json:"scope"`
	Resource     []string `json:"resource"`
}

func (sts *SAMSTokenSource) Token() (*oauth2.Token, error) {
	sts.lock.Lock()
	defer sts.lock.Unlock()

	if sts.token != nil && time.Now().Before(sts.token.Expiry) {
		return sts.token, nil
	}

	tStart := time.Now()
	defer func(tStart time.Time) { fmt.Println("Auth took ", time.Since(tStart).Seconds(), " seconds") }(tStart)

	body := neturl.Values{
		"username": []string{sts.username},
		"password": []string{sts.password},
	}

	resp, err := http.PostForm(sts.url, body)
	if err != nil {
		return nil, err
	}

	defer resp.Body.Close()
	b, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	t := &SAMSToken{}
	if err := json.Unmarshal(b, t); err != nil {
		return nil, err
	}

	sts.token = &oauth2.Token{
		AccessToken:  t.AccessToken,
		TokenType:    t.TokenType,
		RefreshToken: t.RefreshToken,
		Expiry:       time.Now().Add(time.Duration(t.ExpiresIn) * time.Second),
	}
	return sts.token, nil
}

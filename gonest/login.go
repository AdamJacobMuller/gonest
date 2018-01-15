package gonest

import (
	"bytes"
	"encoding/json"
	"fmt"
	"github.com/AdamJacobMuller/golib"
	log "github.com/sirupsen/logrus"
	"github.com/tcnksm/go-input"
	"net/http"
	"net/url"
	"os"
	"path"
	"strings"
)

type LoginRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

type Nest struct {
	httpClient http.Client

	Email     string `json:"email"`
	Password  string `json:"password"`
	CZToken   string `json:"czToken"`
	Website_2 string `json:"website2"`
}

func (n *Nest) Save() error {
	home := os.Getenv("HOME")
	return golib.SaveFile(path.Join(home, ".gonest.json"), n)
}

func (n *Nest) Load() error {
	home := os.Getenv("HOME")
	return golib.LoadFile(path.Join(home, ".gonest.json"), n)
}

func (n *Nest) Login() error {
	czTokenOk, err := n.TestCZToken()
	if err != nil {
		return err
	}

	if !czTokenOk {
		err = n.GetCZToken()
		if err != nil {
			return err
		}
	}

	website2Ok, err := n.TestWebsite2()
	if err != nil {
		return err
	}

	if !website2Ok {
		err = n.GetWebsite2()
		if err != nil {
			return err
		}
	}

	return nil
}

func (n *Nest) TestCZToken() (bool, error) {
	if n.CZToken == "" {
		return false, nil
	}

	req, err := http.NewRequest("GET", "https://home.nest.com/session", nil)
	if err != nil {
		return false, err
	}

	req.Header.Add("Origin", "https://home.nest.com")
	req.Header.Add("Authorization", fmt.Sprintf("Basic %s", n.CZToken))
	req.Header.Add("Referer", "https://home.nest.com/")
	resp, err := n.httpClient.Do(req)
	if err != nil {
		return false, err
	}

	log.WithFields(log.Fields{
		"status": resp.Status,
	}).Info("got czToken validation response")

	if resp.StatusCode == 200 {
		return true, nil
	} else {
		return false, nil
	}
}

func (n *Nest) GetCZToken() error {
	ui := &input.UI{
		Writer: os.Stdout,
		Reader: os.Stdin,
	}

	var err error
	var email string
	var password string

	email = os.Getenv("NEST_EMAIL")
	password = os.Getenv("NEST_PASSWORD")

	if email == "" {
		email, err = ui.Ask("Email", &input.Options{
			Required:  true,
			Loop:      true,
			Mask:      false,
			HideOrder: true,
		})
		if err != nil {
			panic(err)
		}
	}

	if password == "" {
		password, err = ui.Ask("Password", &input.Options{
			Required:  true,
			Loop:      true,
			Mask:      false,
			HideOrder: true,
		})
		if err != nil {
			panic(err)
		}
	}

	loginRequest := LoginRequest{
		Email:    email,
		Password: password,
	}

	postBody, err := json.Marshal(loginRequest)
	if err != nil {
		panic(err)
	}

	var b bytes.Buffer
	b.Write(postBody)
	req, err := http.NewRequest("POST", "https://home.nest.com/session", &b)
	if err != nil {
		panic(err)
	}

	req.Header.Add("Content-Type", "application/json")
	req.Header.Add("Origin", "https://home.nest.com")
	req.Header.Add("Referer", "https://home.nest.com/")
	resp, err := n.httpClient.Do(req)
	if err != nil {
		panic(err)
	}

	fmt.Printf("Response: %s\n", resp.Status)

	for _, cookie := range resp.Cookies() {
		if cookie.Name == "cztoken" {
			n.CZToken = cookie.Value
		}
	}
	return nil
}

func (n *Nest) TestWebsite2() (bool, error) {
	if n.Website_2 == "" {
		return false, nil
	}

	req, err := http.NewRequest("GET", "https://home.nest.com/dropcam/api/login", nil)
	if err != nil {
		return false, err
	}

	req.Header.Add("Origin", "https://home.nest.com")
	req.Header.Add("Cookie", fmt.Sprintf("website_2=%s", n.Website_2))
	req.Header.Add("Referer", "https://home.nest.com/")
	resp, err := n.httpClient.Do(req)
	if err != nil {
		return false, err
	}

	log.WithFields(log.Fields{
		"status": resp.Status,
	}).Info("got website_2 validation response")

	if resp.StatusCode == 200 {
		return true, nil
	} else {
		return false, nil
	}
}

func (n *Nest) GetWebsite2() error {
	form := url.Values{}
	form.Add("access_token", n.CZToken)
	req, err := http.NewRequest("POST", "https://home.nest.com/dropcam/api/login", strings.NewReader(form.Encode()))
	if err != nil {
		return err
	}

	req.Header.Add("Content-Type", "application/x-www-form-urlencoded")
	req.Header.Add("Origin", "https://home.nest.com")
	req.Header.Add("Referer", "https://home.nest.com/")
	resp, err := n.httpClient.Do(req)
	if err != nil {
		return err
	}

	fmt.Printf("Response: %s\n", resp.Status)

	for _, cookie := range resp.Cookies() {
		if cookie.Name == "website_2" {
			n.Website_2 = cookie.Value
		}
	}
	return nil
}

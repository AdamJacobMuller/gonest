package gonest

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/AdamJacobMuller/golib"
	log "github.com/sirupsen/logrus"
	"github.com/tcnksm/go-input"
	"io/ioutil"
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

	DumpRawRequest  bool
	DumpRawResponse bool

	Email     string `json:"email"`
	Password  string `json:"password"`
	CZToken   string `json:"czToken"`
	Website_2 string `json:"website2"`
	N         string `json:"n"`
	UserToken string `json:"user_token"`
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
	return nil
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
	req.Header.Add("User-agent", "Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_3) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/79.0.3945.130 Safari/537.36")
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

/*
{"status":"VERIFICATION_PENDING","2fa_token":"","truncated_phone_number":"3172"}
*/

type NestSessionResponse struct {
	Status   string `json:"status"`
	TfaToken string `json:"2fa_token"`
	Phone    string `json:"truncated_phone_number"`
}

type AccessTokenResponse struct {
	Status      string `json:"status"`
	AccessToken string `json:"access_token"`
}

type TFAVerifyRequest struct {
	TfaToken string `json:"2fa_token"`
	Pin      string `json:"pin"`
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
	req.Header.Add("User-agent", "Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_3) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/79.0.3945.130 Safari/537.36")
	resp, err := n.httpClient.Do(req)
	if err != nil {
		panic(err)
	}

	fmt.Printf("Response: %s\n", resp.Status)

	respBody, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		panic(err)
	}

	var nsr NestSessionResponse
	err = json.Unmarshal(respBody, &nsr)
	if err != nil {
		panic(err)
	}

	fmt.Printf("NSR: %#v\n", nsr)
	if nsr.Status == "VERIFICATION_PENDING" {
		var tfa_cookie_name string
		var tfa_cookie_value string
		var tfa_cookie_found bool
		var tfa_pin string
		for _, cookie := range resp.Cookies() {
			if cookie.Path == "/api/0.1/2fa/verify_pin" {
				tfa_cookie_name = cookie.Name
				tfa_cookie_value = cookie.Value
				tfa_cookie_found = true
				break
			}
		}
		if tfa_cookie_found == false {
			return errors.New("unable to locate tfa cookie")
		}
		fmt.Printf("%s=%s\n", tfa_cookie_name, tfa_cookie_value)
		tfa_pin, err = ui.Ask("2FA Pin", &input.Options{
			Required:  true,
			Loop:      true,
			Mask:      false,
			HideOrder: true,
		})
		if err != nil {
			panic(err)
		}

		tfa_verify_request := TFAVerifyRequest{
			TfaToken: nsr.TfaToken,
			Pin:      tfa_pin,
		}

		postBody, err := json.Marshal(tfa_verify_request)
		if err != nil {
			panic(err)
		}

		var b bytes.Buffer
		b.Write(postBody)
		req, err := http.NewRequest("POST", "https://home.nest.com/api/0.1/2fa/verify_pin", &b)
		if err != nil {
			panic(err)
		}

		req.AddCookie(&http.Cookie{Name: tfa_cookie_name, Value: tfa_cookie_value})
		req.Header.Add("Content-Type", "application/json")
		req.Header.Add("Origin", "https://home.nest.com")
		req.Header.Add("User-agent", "Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_3) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/79.0.3945.130 Safari/537.36")
		req.Header.Add("Referer", "https://home.nest.com/")
		resp, err = n.httpClient.Do(req)
		if err != nil {
			panic(err)
		}

		respBody, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			panic(err)
		}

		var nsr NestSessionResponse
		err = json.Unmarshal(respBody, &nsr)
		if err != nil {
			panic(err)
		}
		for _, cookie := range resp.Cookies() {
			if cookie.Name == "cztoken" {
				n.CZToken = cookie.Value
			}
		}
	} else {
		for _, cookie := range resp.Cookies() {
			if cookie.Name == "cztoken" {
				n.CZToken = cookie.Value
			}
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
	req.Header.Add("User-agent", "Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_3) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/79.0.3945.130 Safari/537.36")
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
	req.Header.Add("User-agent", "Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_3) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/79.0.3945.130 Safari/537.36")
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

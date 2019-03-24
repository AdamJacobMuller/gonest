package gonest

import (
	log "github.com/sirupsen/logrus"

	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
	"strings"
)

func (n *Nest) Delete(url string) (*http.Response, error) {
	request, err := http.NewRequest("DELETE", url, nil)
	if err != nil {
		log.WithFields(log.Fields{
			"error": err,
			"url":   url,
		}).Error("failed to create new DELETE request")
		return nil, err
	}

	request.Header.Add("Origin", "https://home.nest.com")
	request.Header.Add("Referer", "https://home.nest.com/")
	request.Header.Add("Cookie", fmt.Sprintf("cztoken=%s; website_2=%s", n.CZToken, n.Website_2))
	response, err := n.httpClient.Do(request)
	if err != nil {
		log.WithFields(log.Fields{
			"error": err,
			"url":   url,
		}).Error("failed to Do request")
		return nil, err
	}
	response.Body.Close()

	return response, nil
}
func (n *Nest) GetJSONUnmarsahl(url string, result interface{}) error {
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		log.WithFields(log.Fields{
			"url":   url,
			"error": err,
		}).Error("GetJSONUnmarsahl http.NewRequest failed")
		return err
	}

	req.Header.Add("Content-Type", "application/x-www-form-urlencoded")
	req.Header.Add("Origin", "https://home.nest.com")
	req.Header.Add("Referer", "https://home.nest.com/")
	req.Header.Add("Cookie", fmt.Sprintf("cztoken=%s; website_2=%s", n.CZToken, n.Website_2))
	resp, err := n.httpClient.Do(req)
	if err != nil {
		log.WithFields(log.Fields{
			"url":   url,
			"error": err,
		}).Error("GetJSONUnmarsahl HTTP request failed")
		return err
	}

	if resp.StatusCode != 200 {
		log.WithFields(log.Fields{
			"url":    url,
			"status": resp.Status,
		}).Error("GetJSONUnmarsahl invalid status code returned")
		return errors.New("invalid status code returned")
	}

	respBodyBytes, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		log.WithFields(log.Fields{
			"url":   url,
			"error": err,
		}).Error("GetJSONUnmarsahl ioutil.ReadAll failed")
		return err
	}
	resp.Body.Close()

	err = json.Unmarshal(respBodyBytes, result)
	if err != nil {
		log.WithFields(log.Fields{
			"url":   url,
			"error": err,
		}).Error("GetJSONUnmarsahl json.Unmarshal failed")
		return err
	}

	return nil
}

func (n *Nest) PostFormJSONUnmarsahl(url string, form url.Values, result interface{}) error {
	req, err := http.NewRequest("POST", url, strings.NewReader(form.Encode()))
	if err != nil {
		log.WithFields(log.Fields{
			"url":   url,
			"error": err,
		}).Error("PostJSONUnmarsahl http.NewRequest failed")
		return err
	}

	req.Header.Add("Content-Type", "application/x-www-form-urlencoded")
	req.Header.Add("Origin", "https://home.nest.com")
	req.Header.Add("Referer", "https://home.nest.com/")
	req.Header.Add("Cookie", fmt.Sprintf("cztoken=%s; website_2=%s", n.CZToken, n.Website_2))
	resp, err := n.httpClient.Do(req)
	if err != nil {
		log.WithFields(log.Fields{
			"url":   url,
			"error": err,
		}).Error("PostJSONUnmarsahl HTTP request failed")
		return err
	}

	if resp.StatusCode != 200 {
		log.WithFields(log.Fields{
			"url":    url,
			"status": resp.Status,
		}).Error("PostFormJSONUnmarsahl invalid status code returned")
		return errors.New("invalid status code returned")
	}

	respBodyBytes, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		log.WithFields(log.Fields{
			"url":   url,
			"error": err,
		}).Error("PostJSONUnmarsahl ioutil.ReadAll failed")
		return err
	}
	resp.Body.Close()

	err = json.Unmarshal(respBodyBytes, result)
	if err != nil {
		log.WithFields(log.Fields{
			"url":   url,
			"error": err,
		}).Error("PostJSONUnmarsahl json.Unmarshal failed")
		return err
	}

	return nil
}

func (n *Nest) PostJSONUnmarsahl(url string, post interface{}, result interface{}) error {
	postBody, err := json.Marshal(post)
	if err != nil {
		log.WithFields(log.Fields{
			"url":   url,
			"error": err,
		}).Error("PostJSONUnmarsahl json.Marshal failed")
		return err
	}

	var b bytes.Buffer
	b.Write(postBody)

	req, err := http.NewRequest("POST", url, &b)
	if err != nil {
		log.WithFields(log.Fields{
			"url":   url,
			"error": err,
		}).Error("PostJSONUnmarsahl http.NewRequest failed")
		return err
	}

	req.Header.Add("Origin", "https://home.nest.com")
	req.Header.Add("Referer", "https://home.nest.com/")
	req.Header.Add("Cookie", fmt.Sprintf("cztoken=%s; website_2=%s", n.CZToken, n.Website_2))
	resp, err := n.httpClient.Do(req)
	if err != nil {
		log.WithFields(log.Fields{
			"url":   url,
			"error": err,
		}).Error("PostJSONUnmarsahl HTTP request failed")
		return err
	}

	if resp.StatusCode != 200 {
		log.WithFields(log.Fields{
			"status": resp.Status,
		}).Error("invalid status code returned")
		return errors.New("PostJSONUnmarsahl invalid status code returned")
	}

	respBodyBytes, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		log.WithFields(log.Fields{
			"url":   url,
			"error": err,
		}).Error("PostJSONUnmarsahl ioutil.ReadAll failed")
		return err
	}
	resp.Body.Close()

	err = json.Unmarshal(respBodyBytes, result)
	if err != nil {
		log.WithFields(log.Fields{
			"url":   url,
			"error": err,
		}).Error("PostJSONUnmarsahl json.Unmarshal failed")
		return err
	}

	return nil
}

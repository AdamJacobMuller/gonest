package gonest

import (
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"time"

	log "github.com/sirupsen/logrus"
)

func (c Clip) Delete() error {
	url := fmt.Sprintf("https://home.nest.com/dropcam/api/clips/%d", c.ID)
	log.WithFields(log.Fields{
		"id":  c.ID,
		"url": url,
	}).Info("deleting clip")
	response, err := c.nest.Delete(url)
	if err != nil {
		log.WithFields(log.Fields{
			"url": url,
		}).Error("failed to delete clip")
		return err
	} else {
		if response.StatusCode == 200 {
			log.WithFields(log.Fields{
				"url":    url,
				"status": response.Status,
			}).Info("clip deleted")
			return nil
		} else {
			log.WithFields(log.Fields{
				"url":    url,
				"status": response.Status,
			}).Info("failed to delete clip")
			return errors.New("failed to delete clip")
		}
	}
}

func (c Clip) Save(filename string) error {
	fh, err := os.Create(filename)
	if err != nil {
		log.WithFields(log.Fields{
			"error": err,
			"id":    c.ID,
			"url":   c.DownloadURL,
		}).Error("failed open file for clip download")
		return err
	}

	attempts := 0
	for {
		request, err := http.NewRequest("GET", c.DownloadURL, nil)
		if err != nil {
			log.WithFields(log.Fields{
				"error": err,
				"id":    c.ID,
				"url":   c.DownloadURL,
			}).Error("failed create new HTTP request for clip download")
			return err
		}

		response, err := c.nest.httpClient.Do(request)
		if err != nil {
			log.WithFields(log.Fields{
				"error": err,
				"id":    c.ID,
				"url":   c.DownloadURL,
			}).Error("failed to fetch clip")
			return err
		}
		defer response.Body.Close()

		if response.StatusCode == 404 {
			log.WithFields(log.Fields{
				"status":   response.Status,
				"id":       c.ID,
				"url":      c.DownloadURL,
				"attempts": attempts,
			}).Info("waiting for file")
			attempts += 1
			if attempts > 300 {
				return errors.New("unable to save clip: clip not processed after 300 seconds")
			}
			time.Sleep(time.Second)
			continue
		}

		if response.StatusCode != 200 {
			log.WithFields(log.Fields{
				"status":   response.Status,
				"id":       c.ID,
				"url":      c.DownloadURL,
				"attempts": attempts,
			}).Info("hopefully temporary error fetching clip")
			attempts += 1
			if attempts > 300 {
				return errors.New("unable to save clip: too many errors")
			}
			time.Sleep(time.Second)
			continue
		}

		log.WithFields(log.Fields{
			"id":       c.ID,
			"filename": filename,
			"url":      c.DownloadURL,
		}).Info("saving file")
		_, err = io.Copy(fh, response.Body)
		if err != nil {
			log.WithFields(log.Fields{
				"id":       c.ID,
				"filename": filename,
				"error":    err,
				"url":      c.DownloadURL,
			}).Info("saving file failed")
			response.Body.Close()
			return err
		}
		response.Body.Close()
		break
	}

	return nil
}

type Clip struct {
	nest *Nest

	PublicLink  string `json:"public_link"`
	DownloadURL string `json:"download_url"`
	ID          int    `json:"id"`
}

// https://home.nest.com/dropcam/api/visible_clips
func (n *Nest) ListClips() ([]*Clip, error) {
	var clipList []*Clip
	err := n.GetJSONUnmarsahl("https://home.nest.com/dropcam/api/clips/request", &clipList)
	if err != nil {
		return nil, err
	}
	return clipList, nil
}

// https://home.nest.com/camera/50f668e4151745988da09a704458d7f6/clips
func (n *Nest) CreateClip(uuid string, start time.Time, length int) (*Clip, error) {
	form := url.Values{}
	form.Add("uuid", uuid)
	form.Add("start_date", fmt.Sprintf("%d", start.Unix()))
	form.Add("length", fmt.Sprintf("%d", length))
	form.Add("is_public", "true")
	form.Add("is_time_lapse", "false")
	form.Add("donate_video", "false")

	var clipList []Clip
	var err error

	err = n.PostFormJSONUnmarsahl("https://home.nest.com/dropcam/api/clips/request", form, &clipList)
	if err != nil {
		return nil, err
	}

	if len(clipList) == 0 {
		return nil, errors.New("got 0 length clip list")
	}

	clip := clipList[0]
	clip.nest = n

	return &clip, nil

}

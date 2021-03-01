package social

import (
	"bytes"
	"errors"
	"io/ioutil"
	"log"
	"net/http"
	"strconv"
)

type Social struct {
	ServiceURL      string
	ServiceKey      string
	ServiceSecret   string
	TwitterEnabled  bool
	DiscordEnabled  bool
	SlackEnabled    bool
	TelegramEnabled bool
}

// NewSocialChannel creates a new Social Media Connector
func NewSocialChannel(serviceURL string, serviceKey string, serviceSecret string, twitterEnabled bool, discordEnabled bool, slackEnabled bool, telegramEnabled bool) (*Social, error) {
	social := &Social{
		ServiceURL:      serviceURL,
		ServiceKey:      serviceKey,
		ServiceSecret:   serviceSecret,
		TwitterEnabled:  twitterEnabled,
		DiscordEnabled:  discordEnabled,
		SlackEnabled:    slackEnabled,
		TelegramEnabled: telegramEnabled,
	}

	url := social.ServiceURL + "/status"
	resp, err := http.Get(url)
	if err != nil {
		log.Fatalln(err)
	}

	if resp.StatusCode != 200 {
		return nil, errors.New("Social webservice return code is not 200. Webservice url: " + url)
	}

	return social, nil
}

// SendMessage publishes message on enabled social medias
func (social *Social) SendMessage(message string) error {
	if social.DiscordEnabled {
		err := social.sendMessageSocial(message, "discord")
		if err != nil {
			return err
		}
	}
	if social.TwitterEnabled {
		err := social.sendMessageSocial(message, "twitter")
		if err != nil {
			return err
		}
	}
	if social.TelegramEnabled {
		err := social.sendMessageSocial(message, "telegram")
		if err != nil {
			return err
		}
	}
	if social.SlackEnabled {
		err := social.sendMessageSocial(message, "slack")
		if err != nil {
			return err
		}
	}
	return nil
}

func (social *Social) sendMessageSocial(message string, socialMedia string) error {

	url := social.ServiceURL + "/send/" + socialMedia
	jsonStr := []byte("{\"message\":\"" + message + "\"}")

	req, err := http.NewRequest("POST", url, bytes.NewBuffer(jsonStr))
	if err != nil {
		return errors.New("Could not create post request. " + err.Error())
	}
	req.Header.Set("key", social.ServiceKey)
	req.Header.Set("secret", social.ServiceSecret)

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return errors.New("Could not create post request. " + err.Error())
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		return errors.New("Invalid return code: " + strconv.Itoa(resp.StatusCode))
	}

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		log.Fatal(err)
	}
	log.Println("Message sent: ", message)
	log.Println(string(body))

	return nil
}

package main

import (
	"bytes"
	"crypto/md5"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"time"

	log "github.com/sirupsen/logrus"
)

var (
	nextSleep  = 3 * time.Minute
	retrySleep = 1 * time.Minute

	bingBase = "https://www.bing.com"
	bingMkts = []string{"fr-dz", "es-ar", "en-au", "nl-be", "fr-be", "es-bo", "bs-ba", "pt-br", "en-ca", "fr-ca", "cs-cz", "es-cl", "es-co", "es-cr", "sr-latn-me", "en-cy", "da-dk", "de-de", "es-ec", "et-ee", "en-eg", "es-sv", "es-es", "fr-fr", "es-gt", "en-gulf", "es-hn", "en-hk", "hr-hr", "en-in", "id-id", "en-ie", "is-is", "it-it", "en-jo", "lv-lv", "en-lb", "lt-lt", "hu-hu", "en-my", "en-mt", "es-mx", "fr-ma", "nl-nl", "en-nz", "es-ni", "en-ng", "nb-no", "de-at", "en-pk", "es-pa", "es-py", "es-pe", "en-ph", "pl-pl", "pt-pt", "es-pr", "es-do", "ro-md", "ro-ro", "en-sa", "de-ch", "en-sg", "sl-si", "sk-sk", "en-za", "sr-latn-rs", "en-lk", "fr-ch", "fi-fi", "sv-se", "fr-tn", "tr-tr", "en-gb", "en-us", "es-uy", "es-ve", "vi-vn", "el-gr", "ru-by", "bg-bg", "ru-kz", "ru-ru", "uk-ua", "he-il", "ar-iq", "ar-sa", "ar-ly", "ar-eg", "ar-gulf", "th-th", "ko-kr", "zh-cn", "zh-tw", "ja-jp", "zh-hk"} // https://www.microsoft.com/en-in/locale.aspx

	token  = ""
	chatID = "" // get by https://api.telegram.org/bot{token}/getUpdates
)

func init() {
	// redis is optional
	if err := r.Info(ctx).Err(); err == nil {
		p = true
	}

	token = os.Getenv("BOT_TOKEN")
	chatID = os.Getenv("CHAT_ID")
	checkTelegram()
}

func checkTelegram() {
	// check token
	url := fmt.Sprintf("https://api.telegram.org/bot%v/getUpdates", token)
	rsp, err := http.Get(url)
	if err != nil {
		panic(err)
	}
	defer rsp.Body.Close()

	var r struct {
		Ok     bool          `json:"ok"`
		Result []interface{} `json:"result"`
	}
	err = json.NewDecoder(rsp.Body).Decode(&r)
	if err != nil {
		panic(err)
	}
	if !r.Ok {
		panic("invalid BOT_TOKEN")
	}

	// check chat id
	url2 := fmt.Sprintf("https://api.telegram.org/bot%v/getChat?chat_id=%v", token, chatID)
	rsp2, err := http.Get(url2)
	if err != nil {
		panic(err)
	}
	defer rsp2.Body.Close()

	var c struct {
		Ok     bool `json:"ok"`
		Result struct {
			ID int64 `json:"id"`
		} `json:"result"`
	}
	err = json.NewDecoder(rsp2.Body).Decode(&c)
	if err != nil {
		panic(err)
	}
	if !c.Ok || fmt.Sprintf("%v", c.Result.ID) != chatID {
		panic("invalid CHAT_ID")
	}
}

type BingRsp struct {
	Images []struct {
		Startdate     string        `json:"startdate"`
		Fullstartdate string        `json:"fullstartdate"`
		Enddate       string        `json:"enddate"`
		URL           string        `json:"url"`
		Urlbase       string        `json:"urlbase"`
		Copyright     string        `json:"copyright"`
		Copyrightlink string        `json:"copyrightlink"`
		Title         string        `json:"title"`
		Quiz          string        `json:"quiz"`
		Wp            bool          `json:"wp"`
		Hsh           string        `json:"hsh"`
		Drk           int           `json:"drk"`
		Top           int           `json:"top"`
		Bot           int           `json:"bot"`
		Hs            []interface{} `json:"hs"`
	} `json:"images"`
	Tooltips struct {
		Loading  string `json:"loading"`
		Previous string `json:"previous"`
		Next     string `json:"next"`
		Walle    string `json:"walle"`
		Walls    string `json:"walls"`
	} `json:"tooltips"`
}

func imageDuplicate(path string) (dup bool, hash string, err error) {
	if dup, err = strSeen(path); dup || err != nil {
		return
	}

	// download check hash
	rsp, err := http.Get(bingBase + path)
	if err != nil {
		return
	}
	defer rsp.Body.Close()

	body, err := io.ReadAll(rsp.Body)
	if err != nil {
		return
	}

	h := md5.Sum(body)
	hash = hex.EncodeToString(h[:])
	dup, err = strSeen(hash)
	return
}

func postBingCh(msg string) error {
	var ip struct {
		ChatID string `json:"chat_id"`
		Text   string `json:"text"`
	}
	ip.Text = msg
	ip.ChatID = chatID

	data, err := json.Marshal(ip)
	if err != nil {
		return err
	}

	url := fmt.Sprintf("https://api.telegram.org/bot%v/sendMessage", token)
	req, err := http.NewRequest("POST", url, bytes.NewBuffer(data))
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return err
	}
	resp.Body.Close()
	return nil
}

func dailyBing() {
	sleep := nextSleep
	remote := "https://www.bing.com/HPImageArchive.aspx?format=js&idx=0&n=1&mkt=" // 1 is the image count

	path := ""
	hash := ""
	msg := ""
	i := 0 // market index
	for {
		var r BingRsp
		rsp, err := http.Get(remote + bingMkts[i])
		if err != nil {
			goto retry
		}

		err = json.NewDecoder(rsp.Body).Decode(&r)
		if err != nil {
			goto retry
		}
		rsp.Body.Close() // discard

		if len(r.Images) != 1 {
			log.Warnf("format changed, image returned %v", len(r.Images))
			goto retry
		}

		path = r.Images[0].URL
		if dup, h, err := imageDuplicate(path); err != nil {
			goto retry
		} else {
			if dup {
				goto nextMkt
			}
			hash = h
		}

		msg = fmt.Sprintf("%v\n%v\n%v", r.Images[0].Startdate, r.Images[0].Copyright, bingBase+path)
		err = postBingCh(msg)
		if err != nil {
			goto retry
		}

		sleep = nextSleep
	nextMkt:
		i++
		if i == len(bingMkts) {
			i = 0
		}
		goto sleep
	retry:
		sleep = retrySleep
	sleep:
		setSeen(path, hash) // only set after successful post
		time.Sleep(sleep)
	}
}

func main() {
	dailyBing()
}

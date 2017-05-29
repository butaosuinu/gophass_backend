package events

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"net/url"
	"time"
)

const (
	CONNPASS_URL   = "https://connpass.com/api/v1/event/"
	DOORKEEPER_URL = "https://api.doorkeeper.jp/events/"
)

type Query struct {
	Keywords []string
	Address  string
	Month    time.Time
}

// 最終的な検索結果
type ResultEvent struct {
	Title     string
	Address   string
	StartDate time.Time
	Url       string
	Limit     int
	Accepted  int
}

type ResConnpass struct {
	ResultsReturned int `json:"results_returned"`
	Events          []struct {
		EventURL      string `json:"event_url"`
		EventType     string `json:"event_type"`
		OwnerNickname string `json:"owner_nickname"`
		Series        struct {
			URL   string `json:"url"`
			ID    int    `json:"id"`
			Title string `json:"title"`
		} `json:"series"`
		UpdatedAt        time.Time `json:"updated_at"`
		Lat              string    `json:"lat"`
		StartedAt        time.Time `json:"started_at"`
		HashTag          string    `json:"hash_tag"`
		Title            string    `json:"title"`
		EventID          int       `json:"event_id"`
		Lon              string    `json:"lon"`
		Waiting          int       `json:"waiting"`
		Limit            int       `json:"limit"`
		OwnerID          int       `json:"owner_id"`
		OwnerDisplayName string    `json:"owner_display_name"`
		Description      string    `json:"description"`
		Address          string    `json:"address"`
		Catch            string    `json:"catch"`
		Accepted         int       `json:"accepted"`
		EndedAt          time.Time `json:"ended_at"`
		Place            string    `json:"place"`
	} `json:"events"`
	ResultsStart     int `json:"results_start"`
	ResultsAvailable int `json:"results_available"`
}

type ResDoorkeeper []struct {
	Event struct {
		Title        string    `json:"title"`
		ID           int       `json:"id"`
		StartsAt     time.Time `json:"starts_at"`
		EndsAt       time.Time `json:"ends_at"`
		VenueName    string    `json:"venue_name"`
		Address      string    `json:"address"`
		Lat          string    `json:"lat"`
		Long         string    `json:"long"`
		TicketLimit  int       `json:"ticket_limit"`
		PublishedAt  time.Time `json:"published_at"`
		UpdatedAt    time.Time `json:"updated_at"`
		Group        int       `json:"group"`
		Banner       string    `json:"banner"`
		Description  string    `json:"description"`
		PublicURL    string    `json:"public_url"`
		Participants int       `json:"participants"`
		Waitlisted   int       `json:"waitlisted"`
	} `json:"event"`
}

func buildUrlConnpass(query Query) string {
	u, err := url.Parse(CONNPASS_URL)
	if err != nil {
		log.Fatal(err)
	}
	val := url.Values{}
	if query.Keywords != nil {
		for _, v := range query.Keywords {
			val.Add("keyword", v)
		}
	}
	if query.Address != "" {
		val.Add("keyword", query.Address)
	}
	if query.Month.IsZero() == false {
		val.Add("ym", query.Month.Format("200601"))
	}

	u.RawQuery = val.Encode()
	return u.String()
}

func buildUrlDoorkeeper(query Query) string {
	u, err := url.Parse(DOORKEEPER_URL)
	if err != nil {
		log.Fatal(err)
	}
	val := url.Values{}
	if query.Keywords != nil {
		for _, v := range query.Keywords {
			val.Add("q", v)
		}
	}
	if query.Address != "" {
		val.Add("q", query.Address)
	}
	if query.Month.IsZero() == false {
		val.Add("since", "")
		val.Add("until", query.Month.Format("2006-01-02T15:04:05-0700"))
	}

	u.RawQuery = val.Encode()
	return u.String()
}

func searchConnpass(query Query) (*ResConnpass, error) {
	res, err := http.Get(buildUrlConnpass(query))
	if err != nil {
		fmt.Errorf("Error: %s", err)
		return &ResConnpass{}, err
	}

	defer res.Body.Close()
	if res.StatusCode != 200 {
		fmt.Errorf("StatusCode=%d", res.StatusCode)
		return &ResConnpass{}, err
	}

	body, err := ioutil.ReadAll(res.Body)
	if err != nil {
		fmt.Errorf("Error : %s", err)
		return &ResConnpass{}, err
	}

	var r = ResConnpass{}
	rBody, err := jsonParse(body, r)
	return rBody.(*ResConnpass), err
}

func searchDoorkeeper(query Query) (ResDoorkeeper, error) {
	res, err := http.Get(buildUrlDoorkeeper(query))
	if err != nil {
		fmt.Errorf("Error : %s", err)
		return nil, err
	}

	defer res.Body.Close()
	if res.StatusCode != 200 {
		fmt.Errorf("StatusCode=%d", res.StatusCode)
		return nil, err
	}

	body, err := ioutil.ReadAll(res.Body)
	if err != nil {
		fmt.Errorf("Error : %s", err)
		return nil, err
	}

	var r = ResDoorkeeper{}
	rBody, err := jsonParse(body, r)
	return rBody.(ResDoorkeeper), err
}

func jsonParse(jsonBlob []byte, res interface{}) (interface{}, error) {
	err := json.Unmarshal(jsonBlob, &res)
	if err != nil {
		fmt.Errorf("Error : %s", err)
		return nil, err
	}

	return res, nil
}

// SearchEvents serches events on some API and DB
func SearchEvents(query Query) ([]ResultEvent, error) {
	cCon := make(chan *ResConnpass)
	cDok := make(chan ResDoorkeeper)

	go func() {
		resCon, err := searchConnpass(query)
		if err != nil {
			cCon <- &ResConnpass{}
			return
		}
		cCon <- resCon
	}()

	go func() {
		resDok, err := searchDoorkeeper(query)
		if err != nil {
			cDok <- nil
			return
		}
		cDok <- resDok
	}()

	rCon := <-cCon
	rDok := <-cDok

	var resultData []ResultEvent

	for _, v := range rCon.Events {
		var event = ResultEvent{
			Title:     v.Title,
			Address:   v.Address,
			StartDate: v.StartedAt,
			Url:       v.EventURL,
			Limit:     v.Limit,
			Accepted:  v.Accepted,
		}
		resultData = append(resultData, event)
	}

	for _, v := range rDok {
		var event = ResultEvent{
			Title:     v.Event.Title,
			Address:   v.Event.Address,
			StartDate: v.Event.StartsAt,
			Url:       v.Event.PublicURL,
			Limit:     v.Event.TicketLimit,
			Accepted:  v.Event.Participants,
		}
		resultData = append(resultData, event)
	}

	return resultData, nil
}

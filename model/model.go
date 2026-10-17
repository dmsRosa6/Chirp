package model

import "encoding/json"

type Subject struct {
	Pattern     string       `json:"pattern"`
	Subscribers []Subscriber `json:"subscribers"`
	Groups      []Group      `json:"groups"`
}

type Subscriber struct {
	SessionID uint64 `json:"session"`
	ID        string `json:"id"`
}

type Group struct {
	ID          string       `json:"id"`
	Subscribers []Subscriber `json:"subscribers"`
}

func (s Subject) String() string {
	b, err := json.MarshalIndent(s, "", "  ")
	if err != nil {
		return `{"error":"` + err.Error() + `"}`
	}
	return string(b)
}

func (s Subject) JSON() ([]byte, error) {
	return json.Marshal(s)
}

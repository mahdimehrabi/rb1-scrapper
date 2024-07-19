package entity

import "encoding/json"

type URL struct {
	Name  string `json:"name"`
	Query string `json:"query"`
}

func (u *URL) JSON() ([]byte, error) {
	return json.Marshal(u)
}

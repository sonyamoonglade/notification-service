package entity

type Template struct {
	EventID uint64 `json:"event_id"`
	Text    string `json:"text"`
}

type Templates struct {
	Templates []Template `json:"templates"`
}

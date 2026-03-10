package models

type Url struct {
	Id           int64  `json:"id"`
	Username     string `json:"username"`
	Short_code   string `json:"short_code"`
	Original_url string `json:"original_url"`
}

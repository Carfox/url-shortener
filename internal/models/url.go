// STRUCTS
package models

import "time"

type UrlData struct {
	ID        string    `json:"id" bson:"_id,omitempty"`
	Url       string    `json:"url" bson:"url"`
	ShortCode string    `json:"short_code" bson:"short_code"`
	CreatedAt time.Time `json:"created_at" bson:"created_at"`
	UpdatedAt time.Time `json:"updated_at" bson:"updated_at"`
	Clicks    int       `json:"access_count" bson:"access_count"`
}

type UrlResponse struct {
	ID        string    `json:"id"`
	Url       string    `json:"url"`
	ShortCode string    `json:"short_code"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

type UrlRequest struct {
	Url string `json:"url"`
}

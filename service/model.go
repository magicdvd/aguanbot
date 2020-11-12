package service

import (
	rawJson "encoding/json"
)

type UserToken struct {
	User
	Token   string
	Refresh string
}

type User struct {
	Phone string
}

type RespData struct {
	Msg   string             `json:"msg"`
	State int                `json:"state"`
	Data  rawJson.RawMessage `json:"data"`
}

type Post struct {
	ID               uint        `json:"id"`
	AccountID        uint        `json:"accountId"`
	TagAgreeCount    int         `json:"-"`
	Content          string      `json:"content"`
	Title            string      `json:"title"`
	Status           int         `json:"status"`
	CreateAt         string      `json:"createAt"`
	PostAt           string      `json:"postAt"`
	LastTagAt        string      `json:"lastTagAt"`
	Album            interface{} `json:"album"`
	ContentType      int         `json:"contentType"`
	CountOfInvisible int         `json:"countOfInvisible"`
	Author           interface{} `json:"author"`
	Tags             []struct {
		Change       int         `json:"change"`
		Current      interface{} `json:"current"`
		ID           uint        `json:"id"`
		Content      string      `json:"content"`
		CountOfAgree int         `json:"countOfAgree"`
		CreateAt     string      `json:"createAt"`
		PostID       int         `json:"postId"`
		Status       int         `json:"status"`
		CreatorID    int         `json:"creatorId"`
		Agreed       int         `json:"agreed"`
		Creator      interface{} `json:"creator"`
		Enhancement  int         `json:"enhancement"`
	} `json:"tags"`
	Order           interface{} `json:"order"`
	AssignedInfo    interface{} `json:"assignedInfo"`
	FavoriteAt      interface{} `json:"favoriteAt"`
	OrderTime       interface{} `json:"orderTime"`
	CountOfFavorite int         `json:"countOfFavorite"`
	FavoriteStatus  int         `json:"favoriteStatus"`
	PaymentAt       interface{} `json:"paymentAt"`
	Item            interface{} `json:"item"`
	InRecommend     int         `json:"inRecommend"`
}

type Data struct {
	NextOffset int    `json:"nextOffset"`
	Posts      []Post `json:"posts"`
}

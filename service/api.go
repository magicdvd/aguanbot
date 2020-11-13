package service

import (
	"context"
	"fmt"
	"io/ioutil"
	"net/http"
	"strings"

	"github.com/whatisfaker/zaptrace/log"
)

type API interface {
	Login(context.Context, string, string) (*UserToken, error)
	GetList(context.Context, *UserToken) ([]Post, *UserToken, error)
	Follow(context.Context, *UserToken, uint) (*UserToken, error)
	UnFollow(context.Context, *UserToken, uint) (*UserToken, error)
	AgreeTag(context.Context, *UserToken, uint) (*UserToken, error)
}
type aguanAPI struct {
	Domain string
	log    *log.Factory
}

var _ API = (*aguanAPI)(nil)

func NewAguanAPI(log *log.Factory) *aguanAPI {
	return &aguanAPI{
		Domain: "https://api.aguan.net",
		log:    log,
	}
}

func refreshToken(res *http.Response, old *UserToken) *UserToken {
	authToken := res.Header.Get("Authorization")
	refreshToken := res.Header.Get("Refresh")
	if authToken != "" && refreshToken != "" {
		return &UserToken{
			Token:   authToken,
			Refresh: refreshToken,
		}
	}
	return old
}

func (c *aguanAPI) Login(ctx context.Context, user string, password string) (*UserToken, error) {
	url := c.Domain + "/api/public/signIn"
	payload := strings.NewReader(fmt.Sprintf("{\n\t\"phoneNumber\":\"%s\",\n\t\"password\":\"%s\"\n}", user, password))
	client := &http.Client{}
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, url, payload)
	if err != nil {
		return nil, err
	}
	req.Header.Add("Content-Type", "application/json")
	res, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer res.Body.Close()
	body, err := ioutil.ReadAll(res.Body)
	if err != nil {
		return nil, err
	}
	token := refreshToken(res, nil)
	if token == nil {
		return nil, fmt.Errorf("incorrect login token (%s) (%s)", string(body), user)
	}
	return token, nil
}

func (c *aguanAPI) GetList(ctx context.Context, u *UserToken) ([]Post, *UserToken, error) {
	url := c.Domain + "/api/private/getRollingPosts"
	payload := strings.NewReader("")
	client := &http.Client{}
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, url, payload)
	if err != nil {
		return nil, nil, err
	}
	req.Header.Add("Content-Type", "application/json")
	req.Header.Set("Authorization", u.Token)
	req.Header.Set("Refresh", u.Refresh)
	res, err := client.Do(req)
	if err != nil {
		return nil, nil, err
	}
	defer res.Body.Close()
	// body, err := ioutil.ReadAll(res.Body)
	// if err != nil {
	// 	return nil, nil, err
	// }
	decoder := json.NewDecoder(res.Body)
	data := RespData{}
	err = decoder.Decode(&data)
	if err != nil {
		return nil, nil, fmt.Errorf("GetList %v", err)
	}
	if data.State != 1 {
		return nil, nil, fmt.Errorf("GetList error state %v", data)
	}
	data2 := Data{}
	err = json.Unmarshal(data.Data, &data2)
	if err != nil {
		return nil, nil, fmt.Errorf("GetList parse data %v", err)
	}
	token := refreshToken(res, u)
	return data2.Posts, token, nil
}

func (c *aguanAPI) Follow(ctx context.Context, u *UserToken, userID uint) (*UserToken, error) {
	url := c.Domain + "/api/private/setRelation"
	payload := strings.NewReader(fmt.Sprintf("{\n\t\"targetId\":%d,\n\t\"status\":1\n}", userID))
	client := &http.Client{}
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, url, payload)
	if err != nil {
		return nil, err
	}
	req.Header.Add("Content-Type", "application/json")
	req.Header.Set("Authorization", u.Token)
	req.Header.Set("Refresh", u.Refresh)
	res, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer res.Body.Close()
	decoder := json.NewDecoder(res.Body)
	data := RespData{}
	err = decoder.Decode(&data)
	if err != nil {
		return nil, fmt.Errorf("Follow %v", err)
	}
	if data.State != 1 {
		return nil, fmt.Errorf("Follow error state %v", data)
	}
	token := refreshToken(res, u)
	return token, nil
}

func (c *aguanAPI) UnFollow(ctx context.Context, u *UserToken, userID uint) (*UserToken, error) {
	url := c.Domain + "/api/private/setRelation"
	payload := strings.NewReader(fmt.Sprintf("{\n\t\"targetId\":%d,\n\t\"status\":0\n}", userID))
	client := &http.Client{}
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, url, payload)
	if err != nil {
		return nil, err
	}
	req.Header.Add("Content-Type", "application/json")
	req.Header.Set("Authorization", u.Token)
	req.Header.Set("Refresh", u.Refresh)
	res, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer res.Body.Close()
	decoder := json.NewDecoder(res.Body)
	data := RespData{}
	err = decoder.Decode(&data)
	if err != nil {
		return nil, fmt.Errorf("UnFollow %v", err)
	}
	if data.State != 1 {
		return nil, fmt.Errorf("UnFollow error state %v", data)
	}
	token := refreshToken(res, u)
	return token, nil
}

func (c *aguanAPI) AgreeTag(ctx context.Context, u *UserToken, tagID uint) (*UserToken, error) {
	url := c.Domain + "/api/private/agreeTag"
	payload := strings.NewReader(fmt.Sprintf("{\n\t\"tagId\"\t:%d\n}", tagID))
	client := &http.Client{}
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, url, payload)
	if err != nil {
		return nil, err
	}
	req.Header.Add("Content-Type", "application/json")
	req.Header.Set("Authorization", u.Token)
	req.Header.Set("Refresh", u.Refresh)
	res, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer res.Body.Close()
	decoder := json.NewDecoder(res.Body)
	data := RespData{}
	err = decoder.Decode(&data)
	if err != nil {
		return nil, fmt.Errorf("AgreeTag %v", err)
	}
	if data.State != 1 {
		return nil, fmt.Errorf("AgreeTag error state %v", data)
	}
	token := refreshToken(res, u)
	return token, nil
}

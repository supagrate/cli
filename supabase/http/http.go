package http

import (
	"io"
	"net/http"
	"strings"
)

const supabaseUrl = "https://api.supabase.com/v1"

type SupabaseClient struct {
	c           http.Client
	accessToken string
}

func (c *SupabaseClient) Get(url string) (resp []byte, err error) {
	path := supabaseUrl + url
	req, err := http.NewRequest("GET", path, nil)

	if err != nil {
		return nil, err

	}

	res, err := c.Do(req)

	if err != nil {
		return nil, err
	}

	body, err := io.ReadAll(res.Body)

	if err != nil {
		return nil, err
	}

	defer res.Body.Close()

	return body, nil

}

func (c *SupabaseClient) Post(url, contentType string, body io.Reader) (resp *http.Response, err error) {
	req, err := http.NewRequest("POST", supabaseUrl+url, body)

	if err != nil {
		return nil, err

	}

	return c.Do(req)
}

func (c *SupabaseClient) Do(req *http.Request) (*http.Response, error) {
	bearer := "Bearer " + c.accessToken
	req.Header.Add("Authorization", bearer)
	req.Header.Add("Content-Type", "application/json")
	return c.c.Do(req)
}

func NewSupabaseClient(accessToken string) SupabaseClient {
	return SupabaseClient{
		c:           http.Client{},
		accessToken: strings.Trim(accessToken, " \n"),
	}
}

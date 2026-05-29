package csustkit

import (
	"fmt"
	"net/http"
	"net/http/cookiejar"
)

type ServiceDomain string

const (
	ServiceAuthServer ServiceDomain = "authServer"
	ServiceEhall      ServiceDomain = "ehall"
	ServiceCampusCard ServiceDomain = "campusCard"
)

type Client struct {
	httpClient *http.Client
	sso        SSOHelper
	campusCard CampusCardHelper
}

func NewClient() (*Client, error) {
	jar, err := cookiejar.New(nil)
	if err != nil {
		return nil, err
	}

	c := &Client{
		httpClient: &http.Client{
			Jar: jar,
			CheckRedirect: func(req *http.Request, via []*http.Request) error {
				if len(via) >= 30 {
					return fmt.Errorf("stopped after 30 redirects")
				}
				return nil
			},
		},
	}
	c.sso.client = c
	c.campusCard.client = c

	return c, nil
}

func (c *Client) SSO() *SSOHelper {
	return &c.sso
}

func (c *Client) CampusCard() *CampusCardHelper {
	return &c.campusCard
}

func (c *Client) makeURL(domain ServiceDomain, path string) string {
	if path == "" || path[0] != '/' {
		path = "/" + path
	}

	switch domain {
	case ServiceAuthServer:
		return "https://authserver.csust.edu.cn" + path
	case ServiceEhall:
		return "https://ehall.csust.edu.cn" + path
	case ServiceCampusCard:
		return "https://hxyxh5.csust.edu.cn" + path
	default:
		return path
	}
}

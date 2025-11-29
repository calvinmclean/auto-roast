package twchart

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/calvinmclean/babyapi"
	"github.com/calvinmclean/twchart"
)

type Probes []twchart.Probe

type Client struct {
	client    *babyapi.Client[*session]
	sessionID string
}

// the alias is used because the twchart.Session has UnmarshalText which invalidates UnmarshalJSON
type sessionAlias twchart.Session
type session struct {
	babyapi.DefaultResource
	Session sessionAlias
}

func NewClient(addr string) Client {
	client := babyapi.NewClient[*session](addr, "/sessions")
	return Client{client: client}
}

func (c *Client) CreateSession(ctx context.Context, beanName string, probes Probes) (string, error) {
	resp, err := c.client.Post(ctx, &session{
		Session: sessionAlias(twchart.Session{
			Name:   beanName,
			Date:   time.Now(),
			Probes: []twchart.Probe(probes),
		}),
	})
	if err != nil {
		return "", err
	}

	c.sessionID = resp.Data.GetID()

	return resp.Data.GetID(), nil
}

func (c Client) SetStartTime(ctx context.Context, startTime time.Time) error {
	_, err := c.client.Patch(ctx, c.sessionID, &session{Session: sessionAlias(twchart.Session{
		StartTime: startTime,
	})})
	return err
}

func (c Client) AddEvent(ctx context.Context, note string, now time.Time) error {
	e := twchart.Event{Note: note, Time: now}

	url, _ := c.client.URL(c.sessionID)
	url += "/add-event"

	return c.makeRequest(ctx, url, e)
}

func (c Client) AddStage(ctx context.Context, name string, now time.Time) error {
	s := twchart.Stage{Name: name, Start: now}

	url, _ := c.client.URL(c.sessionID)
	url += "/add-stage"

	return c.makeRequest(ctx, url, s)
}

func (c Client) Done(ctx context.Context) error {
	url, _ := c.client.URL(c.sessionID)
	url += "/done"

	return c.makeRequest(ctx, url, map[string]any{"time": time.Now()})
}

func (c Client) makeRequest(ctx context.Context, url string, body any) error {
	var bodyReader io.Reader = http.NoBody
	if body != nil {
		bodyBytes, err := json.Marshal(body)
		if err != nil {
			return fmt.Errorf("error encoding body: %w", err)
		}

		bodyReader = bytes.NewReader(bodyBytes)
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, url, bodyReader)
	if err != nil {
		return fmt.Errorf("error creating request: %w", err)
	}
	resp, err := c.client.MakeGenericRequest(req, nil)
	if err != nil {
		return fmt.Errorf("error making request: %w", err)
	}
	if resp.Response.StatusCode != http.StatusOK {
		return fmt.Errorf("unexpected status code: %d, response: %v", resp.Response.StatusCode, resp.Body)
	}

	return nil
}

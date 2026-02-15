package twchart

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	"github.com/calvinmclean/babyapi"
	"github.com/calvinmclean/twchart"
)

type Probes []twchart.Probe

type Client struct {
	client    *babyapi.Client[*session]
	sessionID string
}

type session struct {
	// include NilResource so we don't implement Render/Bind which are not needed
	*babyapi.NilResource
	twchart.Session
}

func (s session) GetID() string {
	return s.Session.GetID()
}

func NewClient(addr string) *Client {
	client := babyapi.NewClient[*session](addr, "/sessions")
	return &Client{client: client}
}

func (c *Client) CreateSession(ctx context.Context, beanName string, probes Probes) (string, error) {
	resp, err := c.client.Post(ctx, &session{
		Session: twchart.Session{
			Name:   beanName,
			Type:   twchart.SessionTypeCoffee,
			Date:   time.Now(),
			Probes: []twchart.Probe(probes),
		},
	})
	if err != nil {
		return "", err
	}

	c.sessionID = resp.Data.GetID()

	return resp.Data.GetID(), nil
}

func (c Client) SetStartTime(ctx context.Context, startTime time.Time) error {
	_, err := c.client.Patch(ctx, c.sessionID, &session{Session: twchart.Session{
		StartTime: startTime,
	}})
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

	req.Header.Add("Content-Type", "application/json")

	resp, err := c.client.MakeGenericRequest(req, nil)
	if err != nil {
		return fmt.Errorf("error making request: %w", err)
	}
	if resp.Response.StatusCode != http.StatusNoContent {
		return fmt.Errorf("unexpected status code: %d, response: %v", resp.Response.StatusCode, resp.Body)
	}

	return nil
}

// ParseProbes parses a string in the format "1=Name,2=Name,..." into twchart.Probes.
func ParseProbes(input string) (Probes, error) {
	var probes Probes
	entries := strings.SplitSeq(input, ",")
	for entry := range entries {
		parts := strings.SplitN(entry, "=", 2)
		if len(parts) != 2 {
			return nil, fmt.Errorf("invalid probe entry: %q", entry)
		}
		posStr := strings.TrimSpace(parts[0])
		name := strings.TrimSpace(parts[1])

		var pos twchart.ProbePosition
		_, err := fmt.Sscanf(posStr, "%d", &pos)
		if err != nil || pos <= twchart.ProbePositionNone {
			return nil, fmt.Errorf("invalid probe position: %q", posStr)
		}
		probes = append(probes, twchart.Probe{Name: name, Position: pos})
	}
	return probes, nil
}

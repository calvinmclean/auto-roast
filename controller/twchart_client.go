package controller

import (
	"context"
	"time"

	"github.com/calvinmclean/autoroast/twchart"
)

type twchartClient interface {
	CreateSession(ctx context.Context, beanName string, probes twchart.Probes) (string, error)
	SetStartTime(ctx context.Context, startTime time.Time) error
	AddEvent(ctx context.Context, note string, now time.Time) error
	AddStage(ctx context.Context, name string, now time.Time) error
	Done(ctx context.Context) error
}

type noopTWChartClient struct{}

var _ twchartClient = noopTWChartClient{}

// AddEvent implements twchartClient.
func (n noopTWChartClient) AddEvent(ctx context.Context, note string, now time.Time) error {
	return nil
}

// AddStage implements twchartClient.
func (n noopTWChartClient) AddStage(ctx context.Context, name string, now time.Time) error {
	return nil
}

// CreateSession implements twchartClient.
func (n noopTWChartClient) CreateSession(ctx context.Context, beanName string, probes twchart.Probes) (string, error) {
	return "", nil
}

// Done implements twchartClient.
func (n noopTWChartClient) Done(ctx context.Context) error {
	return nil
}

// SetStartTime implements twchartClient.
func (n noopTWChartClient) SetStartTime(ctx context.Context, startTime time.Time) error {
	return nil
}

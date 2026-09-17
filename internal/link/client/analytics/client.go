// Package analytics provides a gRPC client adapter to communicate with analytics-service.
package analytics

import (
	"context"
	"fmt"

	"github.com/hel1th/giga-url-shortener/internal/link/domain"
	"github.com/hel1th/giga-url-shortener/internal/link/usecase"
	analyticsv1 "github.com/hel1th/giga-url-shortener/proto/analytics/v1"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

var _ usecase.AnalyticsClient = (*Client)(nil)

// Client implements usecase.AnalyticsClient via gRPC.
type Client struct {
	grpcClient analyticsv1.AnalyticsServiceClient
}

// NewClient creates a new analytics gRPC client adapter.
func NewClient(cc grpc.ClientConnInterface) *Client {
	return &Client{
		grpcClient: analyticsv1.NewAnalyticsServiceClient(cc),
	}
}

// GetStats requests click statistics for a short link from the analytics service.
func (c *Client) GetStats(ctx context.Context, shortCode string) (*domain.LinkStats, error) {
	resp, err := c.grpcClient.GetLinkSummary(ctx, &analyticsv1.GetLinkSummaryRequest{ShortCode: shortCode})
	if err == nil {
		return respToStats(resp), nil
	}

	st, ok := status.FromError(err)
	if ok {
		switch st.Code() {
		case codes.NotFound:
			return nil, domain.ErrNotFound
		case codes.DeadlineExceeded, codes.Canceled:
			return nil, fmt.Errorf("analytics service timeout: %w", err)
		}
	}
	return nil, fmt.Errorf("analytics gRPC call failed: %w", err)
}

func respToStats(resp *analyticsv1.GetLinkSummaryResponse) *domain.LinkStats {
	return &domain.LinkStats{
		ShortCode:    resp.GetShortCode(),
		TotalClicks:  resp.GetTotalClicks(),
		UniqueClicks: resp.GetUniqueClicks(),
		Browsers:     resp.GetBrowsers(),
		OS:           resp.GetOs(),
		Devices:      resp.GetDevices(),
	}
}

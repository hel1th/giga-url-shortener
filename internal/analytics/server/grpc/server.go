package grpc

import (
	"context"

	"github.com/hel1th/giga-url-shortener/internal/analytics/domain"
	analyticsv1 "github.com/hel1th/giga-url-shortener/proto/analytics/v1"
)

type Server struct {
	analyticsv1.UnimplementedAnalyticsServiceServer
	uc AnalyticsUseCase
}

func NewServer(uc AnalyticsUseCase) *Server {
	return &Server{uc: uc}
}

type AnalyticsUseCase interface {
	GetStats(ctx context.Context, shortCode string) (*domain.LinkSummary, error)
}

func (s *Server) GetLinkSummary(
	ctx context.Context,
	req *analyticsv1.GetLinkSummaryRequest,
) (*analyticsv1.GetLinkSummaryResponse, error) {
	stats, err := s.uc.GetStats(ctx, req.GetShortCode())
	if err != nil {
		return nil, err
	}

	return &analyticsv1.GetLinkSummaryResponse{
		ShortCode:    req.GetShortCode(),
		TotalClicks:  stats.TotalClicks,
		UniqueClicks: stats.UniqueClicks,
		Browsers:     stats.Browser,
		Os:           stats.OS,
		Devices:      stats.Device,
	}, nil
}

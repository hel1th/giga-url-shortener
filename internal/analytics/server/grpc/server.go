package grpc

import (
	"context"

	analyticsv1 "github.com/hel1th/giga-url-shortener/proto/analytics/v1"
)

type Server struct {
	analyticsv1.UnimplementedAnalyticsServiceServer
	uc AnalyticsUseCase
}
type AnalyticsUseCase interface {
	GetStats(ctx context.Context, shortCode string) (LinkSummary, error)
}

func (s *Server) GetLinkSummary(
	ctx context.Context,
	req *analyticsv1.GetLinkSummaryRequest,
) (*analyticsv1.GetLinkSummaryResponse, error) {
	stats, err := s.uc.GetStats(ctx, req.ShortCode)
}

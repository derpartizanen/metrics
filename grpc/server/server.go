package server

import (
	"context"
	"fmt"
	"net"

	"go.uber.org/zap"
	"google.golang.org/grpc"

	"github.com/derpartizanen/metrics/internal/logger"
	"github.com/derpartizanen/metrics/internal/model"
	"github.com/derpartizanen/metrics/internal/storage"
	"github.com/derpartizanen/metrics/proto"
)

type MetricsServer struct {
	proto.UnimplementedMetricCollectorServer
	server  *grpc.Server
	Service *storage.Storage
}

func NewServer(cfg Config) *MetricsServer {
	config = cfg
	srv := &MetricsServer{Service: cfg.Service}
	srv.server = grpc.NewServer()
	proto.RegisterMetricCollectorServer(srv.server, srv)

	return srv
}

func (s *MetricsServer) Run(ctx context.Context) error {
	listen, err := net.Listen("tcp", config.ServerAddr)
	if err != nil {
		return err
	}
	logger.Log.Info("Running grpc server", zap.String("address", config.ServerAddr))

	if err := s.server.Serve(listen); err != nil {
		return err
	}

	return nil
}

func (s *MetricsServer) Shutdown(ctx context.Context) error {
	s.server.GracefulStop()
	return nil
}

func (s *MetricsServer) Update(ctx context.Context, in *proto.UpdateRequest) (*proto.UpdateResponse, error) {
	metric := model.Metrics{
		ID:    in.Metric.Id,
		MType: in.Metric.Type,
		Value: &in.Metric.Value,
		Delta: &in.Metric.Delta,
	}

	err := s.Service.SaveMetric(metric)
	if err != nil {
		logger.Log.Error("metric update error", zap.Error(err), zap.Any("metric", metric))
		return nil, fmt.Errorf("failed to update metrics: %w", err)
	}

	return &proto.UpdateResponse{}, nil
}

func (s *MetricsServer) Updates(ctx context.Context, in *proto.UpdatesRequest) (*proto.UpdatesResponse, error) {
	metrics := make([]model.Metrics, 0, len(in.Metrics))
	for _, metric := range in.Metrics {
		metrics = append(metrics, model.Metrics{
			ID:    metric.Id,
			MType: metric.Type,
			Delta: &metric.Delta,
			Value: &metric.Value,
		})
	}

	err := s.Service.SetAllMetrics(metrics)
	if err != nil {
		logger.Log.Error("metrics update error", zap.Error(err))
		return nil, fmt.Errorf("failed to update metrics: %w", err)
	}

	return &proto.UpdatesResponse{}, nil
}

func (s *MetricsServer) Value(ctx context.Context, in *proto.ValueRequest) (*proto.ValueResponse, error) {
	metric := model.Metrics{ID: in.Id, MType: in.MetricType}

	err := s.Service.GetMetric(&metric)
	if err != nil {
		return nil, fmt.Errorf("failed to get metric: %w", err)
	}

	m := &proto.Metric{
		Id:   metric.ID,
		Type: metric.MType,
	}
	if metric.Delta != nil {
		m.Delta = *metric.Delta
	}
	if metric.Value != nil {
		m.Value = *metric.Value
	}

	return &proto.ValueResponse{Metric: m}, nil
}

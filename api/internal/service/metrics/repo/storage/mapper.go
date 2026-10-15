package storage

import (
	metricsEntities "dashboard/api/internal/model/metrics"
	pb "dashboard/api/internal/service/metrics/repo/storage/codec"
	"time"

	"google.golang.org/protobuf/types/known/timestamppb"
)

func EncodePoint(point metricsEntities.SeriesPoint) *pb.MetricPoint {

	encoded := &pb.MetricPoint{
		MeasuredAt: timestamppb.New(point.Timestamp.UTC()),
	}

	if intValue, ok := point.Value.Int(); ok {
		encoded.Value = &pb.MetricPoint_IntValue{IntValue: intValue}
		return encoded
	}

	encoded.Value = &pb.MetricPoint_FloatValue{FloatValue: point.Value.Float()}

	return encoded
}

func DecodePoint(message *pb.MetricPoint) metricsEntities.SeriesPoint {

	if message == nil {
		return metricsEntities.SeriesPoint{}
	}

	var value metricsEntities.Number

	switch typed := message.GetValue().(type) {
	case *pb.MetricPoint_IntValue:
		value = metricsEntities.Int64(typed.IntValue)
	case *pb.MetricPoint_FloatValue:
		value = metricsEntities.Float64(typed.FloatValue)
	default:
		value = metricsEntities.Int64(0)
	}

	var measuredAt time.Time

	if message.GetMeasuredAt() != nil {
		measuredAt = message.GetMeasuredAt().AsTime().UTC()
	}

	return metricsEntities.SeriesPoint{
		Timestamp: measuredAt,
		Value:     value,
	}
}

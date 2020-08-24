package core

import (
	"context"
	"encoding/json"
	"errors"
	"github.com/ZhongAnTech/BlockDB/brefactor/core_interface"
	"go.mongodb.org/mongo-driver/bson"
)

type BusinessReader struct {
	storageExecutor core_interface.StorageExecutor
}

func NewBusinessReader(storageExecutor core_interface.StorageExecutor) *BusinessReader {
	s := &BusinessReader{
		storageExecutor: storageExecutor,
	}
	return s
}

func (s *BusinessReader) Close(ctx context.Context) error {
	return s.storageExecutor.Close(ctx)
}

func (s *BusinessReader) Info(ctx context.Context, opHash string) ([]byte, error) {
	filter := bson.D{{"op_hash", opHash}}
	response, err := s.storageExecutor.Select(ctx, "sample_collection", filter, nil, 0, 0)
	if err != nil {
		return nil, err
	}
	if len(response.Content) != 1 {
		return nil, errors.New("miss info")
	}
	return json.Marshal(response.Content[0])
}

func (s *BusinessReader) Actions(ctx context.Context, opHash string) ([]byte, error) {
	filter := bson.D{{"op_hash", opHash}}
	response, err := s.storageExecutor.Select(ctx, "sample_collection", filter, nil, 0, 0)
	if err != nil {
		return nil, err
	}
	return json.Marshal(response.Content)
}

func (s *BusinessReader) Action(ctx context.Context, opHash string, version int) ([]byte, error) {
	filter := bson.D{{"op_hash", opHash}, {"version", version}}
	response, err := s.storageExecutor.Select(ctx, "sample_collection", filter, nil, 0, 0)
	if err != nil {
		return nil, err
	}
	if len(response.Content) != 1 {
		return nil, errors.New("miss action")
	}
	return json.Marshal(response.Content[0])
}

func (s *BusinessReader) Values(ctx context.Context, opHash string) ([]byte, error) {
	filter := bson.D{{"op_hash", opHash}}
	response, err := s.storageExecutor.Select(ctx, "sample_collection", filter, nil, 0, 0)
	if err != nil {
		return nil, err
	}
	return json.Marshal(response.Content)
}

func (s *BusinessReader) Value(ctx context.Context, opHash string, version int) ([]byte, error) {
	filter := bson.D{{"op_hash", opHash}, {"version", version}}
	response, err := s.storageExecutor.Select(ctx, "sample_collection", filter, nil, 0, 0)
	if err != nil {
		return nil, err
	}
	if len(response.Content) != 1 {
		return nil, errors.New("miss value")
	}
	return json.Marshal(response.Content[0])
}

func (s *BusinessReader) CurrentValue(ctx context.Context, opHash string) ([]byte, error) {
	filter := bson.D{{"op_hash", opHash}}
	response, err := s.storageExecutor.Select(ctx, "sample_collection", filter, nil, 0, 0)
	if err != nil {
		return nil, err
	}

	if len(response.Content) != 1 {
		return nil, errors.New("miss info")
	}

	info := response.Content[0]
	version := info["latest_version"]
	return s.Value(ctx, opHash, version.(int))
}

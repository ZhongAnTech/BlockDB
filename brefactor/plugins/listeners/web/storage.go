package web

import (
	"encoding/json"
	"errors"
	"go.mongodb.org/mongo-driver/bson"
)

type Storage struct {
	su storageUtil
}

func (s *Storage) Close() error {
	return s.su.Close()
}

func (s *Storage) Info(opHash string) ([]byte, error) {
	filter := bson.D{{"op_hash", opHash}}
	response, err := s.su.Select("sample_collection", filter, nil, 0, 0)
	if err != nil {
		return nil, err
	}
	if len(response.Content) != 1 {
		return nil, errors.New("miss info")
	}
	return json.Marshal(response.Content[0])
}

func (s *Storage) Actions(opHash string) ([]byte, error) {
	filter := bson.D{{"op_hash", opHash}}
	response, err := s.su.Select("sample_collection", filter, nil, 0, 0)
	if err != nil {
		return nil, err
	}
	return json.Marshal(response.Content)
}

func (s *Storage) Action(opHash string, version int) ([]byte, error) {
	filter := bson.D{{"op_hash", opHash}, {"version", version}}
	response, err := s.su.Select("sample_collection", filter, nil, 0, 0)
	if err != nil {
		return nil, err
	}
	if len(response.Content) != 1 {
		return nil, errors.New("miss action")
	}
	return json.Marshal(response.Content[0])
}

func (s *Storage) Values(opHash string) ([]byte, error) {
	filter := bson.D{{"op_hash", opHash}}
	response, err := s.su.Select("sample_collection", filter, nil, 0, 0)
	if err != nil {
		return nil, err
	}
	return json.Marshal(response.Content)
}

func (s *Storage) Value(opHash string, version int) ([]byte, error) {
	filter := bson.D{{"op_hash", opHash}, {"version", version}}
	response, err := s.su.Select("sample_collection", filter, nil, 0, 0)
	if err != nil {
		return nil, err
	}
	if len(response.Content) != 1 {
		return nil, errors.New("miss value")
	}
	return json.Marshal(response.Content)
}

func (s *Storage) CurrentValue(opHash string) ([]byte, error) {
	filter := bson.D{{"op_hash", opHash}}
	response, err := s.su.Select("sample_collection", filter, nil, 0, 0)
	if err != nil {
		return nil, err
	}

	if len(response.Content) != 1 {
		return nil, errors.New("miss info")
	}

	 info := response.Content[0]
	 version := info["version"]
	 return s.Value(opHash, version.(int))
}


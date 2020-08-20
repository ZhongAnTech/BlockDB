package web

type Storage struct {
	su storageUtil
}

func (s *Storage) Close() error {
	return s.su.Close()
}

//func (s *Storage) Info(opHash string) ([]byte, error) {
//	filter := bson.D{{"op_hash", opHash}}
//	_, err := s.su.Select("sample_collection", filter, nil, 0, 0)
//	if err != nil {
//
//	}
//}
//
//func (s *Storage) Actions(opHash string) ([]byte, error) {
//
//}
//
//func (s *Storage) Action(opHash string, version int) ([]byte, error) {
//
//}
//
//func (s *Storage) Values(opHash string) ([]byte, error) {
//
//}
//
//func (s *Storage) Value(opHash string, version int) ([]byte, error) {
//
//}
//
//func (s *Storage) CurrentValue(opHash string) ([]byte, error) {
//
//}


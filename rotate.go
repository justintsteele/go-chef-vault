package vault

type RotateResponse struct {
	Response
}

func (s *Service) RotateKeys(payload *Payload) (*RotateResponse, error) {
	return &RotateResponse{}, nil
}

func (s *Service) RotateAllKeys() (*RotateResponse, error) {
	return &RotateResponse{}, nil
}

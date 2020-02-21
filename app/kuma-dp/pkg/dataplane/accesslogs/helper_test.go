package accesslogs

type fakeSender []string

func (s *fakeSender) Connect() error {
	return nil
}
func (s *fakeSender) Send(record string) error {
	*s = append(*s, record)
	return nil
}
func (s *fakeSender) Close() error {
	return nil
}

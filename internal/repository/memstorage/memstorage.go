package memstorage

type MemStorage struct {
	gauge   map[string]float64
	counter map[string]int64
}

func New() *MemStorage {
	return &MemStorage{
		gauge:   make(map[string]float64),
		counter: make(map[string]int64),
	}
}

func (s *MemStorage) UpdateGaugeMetric(name string, value float64) error {
	s.gauge[name] = value

	return nil
}

func (s *MemStorage) UpdateCounterMetric(name string, value int64) error {
	s.counter[name] += value

	return nil
}

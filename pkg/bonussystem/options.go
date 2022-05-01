package bonussystem

import "time"

type Option func(system *BonusSystem)

func WithUpdateInterval(t time.Duration) Option {
	return func(s *BonusSystem) {
		s.updateInterval = t
	}
}

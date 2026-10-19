package palworld

import (
	"context"
	"errors"
)

// PlayerCount is a placeholder until REST player monitoring is implemented.
func (*Palworld) PlayerCount(_ context.Context) (int, error) {
	return 0, errors.ErrUnsupported
}

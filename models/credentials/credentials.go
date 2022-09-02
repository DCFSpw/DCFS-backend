package credentials

import "golang.org/x/net/context"

type Credentials interface {
	Authenticate(ctx context.Context) error
}

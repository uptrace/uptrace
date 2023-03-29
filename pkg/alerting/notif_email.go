package alerting

import "context"

func notifyByEmailHandler(ctx context.Context, eventID uint64, recipients []string) error {
	// TODO: support email notif channel
	return nil
}

package wrapper

import "time"

const (
	defaultRetryCount     = 10
	defaultBaseRetryDelay = 100 * time.Millisecond
)

func Bridge(retryCount int, baseRetryDelay time.Duration, bridgeFunc func() error) error {
	var err error

	for i := 0; i < retryCount; i++ {
		err = bridgeFunc()
		if err == nil {
			return nil
		}
		time.Sleep(baseRetryDelay)
		baseRetryDelay *= 2
	}

	return err
}

func MustBridge(retryCount int, baseRetryDelay time.Duration, bridgeFunc func() error) {
	if err := Bridge(retryCount, baseRetryDelay, bridgeFunc); err != nil {
		panic(err)
	}
}

func BridgeWithDefaults(bridgeFunc func() error) error {
	return Bridge(defaultRetryCount, defaultBaseRetryDelay, bridgeFunc)
}

func MustBridgeWithDefaults(bridgeFunc func() error) {
	if err := BridgeWithDefaults(bridgeFunc); err != nil {
		panic(err)
	}
}

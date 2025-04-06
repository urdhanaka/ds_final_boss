package virtualization_utils

import (
	"crypto/rand"
	"fmt"
	"log/slog"
	"math/big"
	"net"
)

func CreateToken() string {
	defaultToken := "loremipsum"
	letterRunes := []rune("abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ")

	res := make([]rune, 16)
	for i := range res {
		index, err := rand.Int(rand.Reader, big.NewInt(int64(len(letterRunes))))
		if err != nil {
			slog.Error("createToken(): could not create token",
				"error", err.Error(),
			)

			return defaultToken
		}

		res[i] = letterRunes[index.Int64()]
	}

	return string(res)
}

func GetRandomPort(bottom, high int) (int, error) {
	portRange := high - bottom

	res, err := rand.Int(rand.Reader, big.NewInt(int64(portRange)))
	if err != nil {
		return 0, err
	}

	return (bottom + int(res.Int64())), nil
}

func IsPortAvailable(port int) bool {
	ln, err := net.Listen("tcp", fmt.Sprintf(":%d", port))
	if err != nil {
		slog.Error("error checking port availability",
			"error", err.Error(),
		)

		return false
	}

	ln.Close()

	return true
}

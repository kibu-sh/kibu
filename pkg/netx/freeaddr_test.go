package netx

import (
	"fmt"
	"github.com/stretchr/testify/require"
	"testing"
)

func TestGetFreeAddr(t *testing.T) {
	addr, err := GetFreeAddr()
	require.NoError(t, err)
	require.NotEqual(t, "0", addr.Port())
	require.Equal(t, fmt.Sprintf("127.0.0.1:%s", addr.Port()), addr.HostPort())
	require.Equal(t, fmt.Sprintf("tcp://127.0.0.1:%s", addr.Port()), addr.String())
}

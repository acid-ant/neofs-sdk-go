package netmap

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestNodeInfo_SetAttribute(t *testing.T) {
	var n NodeInfo

	const key = "some key"
	val := "some value"

	require.Zero(t, n.Attribute(val))

	n.SetAttribute(key, val)
	require.Equal(t, val, n.Attribute(key))

	val = "some other value"

	n.SetAttribute(key, val)
	require.Equal(t, val, n.Attribute(key))
}

func TestNodeInfo_Status(t *testing.T) {
	var n NodeInfo

	require.False(t, n.IsOnline())
	require.False(t, n.IsOffline())
	require.False(t, n.IsMaintenance())

	n.SetOnline()
	require.True(t, n.IsOnline())
	require.False(t, n.IsOffline())
	require.False(t, n.IsMaintenance())

	n.SetOffline()
	require.True(t, n.IsOffline())
	require.False(t, n.IsOnline())
	require.False(t, n.IsMaintenance())

	n.SetMaintenance()
	require.True(t, n.IsMaintenance())
	require.False(t, n.IsOnline())
	require.False(t, n.IsOffline())
}

func TestNodeInfo_ExternalAddr(t *testing.T) {
	var n NodeInfo

	require.Empty(t, n.ExternalAddresses())
	require.Panics(t, func() { n.SetExternalAddresses() })

	addr := []string{"1", "2", "3"}
	n.SetExternalAddresses(addr[0])
	require.Equal(t, addr[:1], n.ExternalAddresses())

	n.SetExternalAddresses(addr[1:]...)
	require.Equal(t, addr[1:], n.ExternalAddresses())
}

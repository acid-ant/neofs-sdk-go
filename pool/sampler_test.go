package pool

import (
	"context"
	"fmt"
	"math/rand"
	"testing"

	"github.com/nspcc-dev/neofs-sdk-go/netmap"
	"github.com/stretchr/testify/require"
)

func TestSamplerStability(t *testing.T) {
	const COUNT = 100000

	cases := []struct {
		probabilities []float64
		expected      []int
	}{
		{
			probabilities: []float64{1, 0},
			expected:      []int{COUNT, 0},
		},
		{
			probabilities: []float64{0.1, 0.2, 0.7},
			expected:      []int{10138, 19813, 70049},
		},
		{
			probabilities: []float64{0.2, 0.2, 0.4, 0.1, 0.1, 0},
			expected:      []int{19824, 20169, 39900, 10243, 9864, 0},
		},
	}

	for _, tc := range cases {
		sampler := newSampler(tc.probabilities, rand.NewSource(0))
		res := make([]int, len(tc.probabilities))
		for i := 0; i < COUNT; i++ {
			res[sampler.Next()]++
		}

		require.Equal(t, tc.expected, res, "probabilities: %v", tc.probabilities)
	}
}

type clientMock struct {
	clientWrapper
	name string
	err  error
}

func (c *clientMock) endpointInfo(context.Context, prmEndpointInfo) (netmap.NodeInfo, error) {
	return netmap.NodeInfo{}, nil
}

func (c *clientMock) networkInfo(context.Context, prmNetworkInfo) (netmap.NetworkInfo, error) {
	return netmap.NetworkInfo{}, nil
}

func newNetmapMock(name string, needErr bool) *clientMock {
	var err error
	if needErr {
		err = fmt.Errorf("not available")
	}
	return &clientMock{
		clientWrapper: clientWrapper{
			clientStatusMonitor: newClientStatusMonitor("", 10),
		},
		name: name,
		err:  err,
	}
}

func TestHealthyReweight(t *testing.T) {
	var (
		weights = []float64{0.9, 0.1}
		names   = []string{"node0", "node1"}
		buffer  = make([]float64, len(weights))
	)

	cache, err := newCache()
	require.NoError(t, err)

	inner := &innerPool{
		sampler: newSampler(weights, rand.NewSource(0)),
		clients: []client{
			newNetmapMock(names[0], true),
			newNetmapMock(names[1], false),
		},
	}
	p := &Pool{
		innerPools:      []*innerPool{inner},
		cache:           cache,
		key:             newPrivateKey(t),
		rebalanceParams: rebalanceParameters{nodesParams: []*nodesParam{{weights: weights}}},
	}

	// check getting first node connection before rebalance happened
	connection0, err := p.connection()
	require.NoError(t, err)
	mock0 := connection0.(*clientMock)
	require.Equal(t, names[0], mock0.name)

	p.updateInnerNodesHealth(context.TODO(), 0, buffer)

	connection1, err := p.connection()
	require.NoError(t, err)
	mock1 := connection1.(*clientMock)
	require.Equal(t, names[1], mock1.name)

	// enabled first node again
	inner.lock.Lock()
	inner.clients[0] = newNetmapMock(names[0], false)
	inner.lock.Unlock()

	p.updateInnerNodesHealth(context.TODO(), 0, buffer)
	inner.sampler = newSampler(weights, rand.NewSource(0))

	connection0, err = p.connection()
	require.NoError(t, err)
	mock0 = connection0.(*clientMock)
	require.Equal(t, names[0], mock0.name)
}

func TestHealthyNoReweight(t *testing.T) {
	var (
		weights = []float64{0.9, 0.1}
		names   = []string{"node0", "node1"}
		buffer  = make([]float64, len(weights))
	)

	sampl := newSampler(weights, rand.NewSource(0))
	inner := &innerPool{
		sampler: sampl,
		clients: []client{
			newNetmapMock(names[0], false),
			newNetmapMock(names[1], false),
		},
	}
	p := &Pool{
		innerPools:      []*innerPool{inner},
		rebalanceParams: rebalanceParameters{nodesParams: []*nodesParam{{weights: weights}}},
	}

	p.updateInnerNodesHealth(context.TODO(), 0, buffer)

	inner.lock.RLock()
	defer inner.lock.RUnlock()
	require.Equal(t, inner.sampler, sampl)
}

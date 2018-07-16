package pooling

import (
	"errors"

	"github.com/dereklstinson/GoCuNets/gocudnn/tensor"
	gocudnn "github.com/dereklstinson/GoCudnn"
)

type Pooling struct {
	desc *gocudnn.PoolingD
}

func BuildPooling(mode gocudnn.PoolingMode, nan gocudnn.PropagationNAN, input tensor.Tensor, window, padding, stride []int32) (*Pooling, error) {
	_, _, dims, err := input.Properties()
	if err != nil {
		return nil, err
	}
	if len(dims) > 4 {
		pooldim := int32(len(window))
		desc, err := gocudnn.Pooling{}.CreatePoolingNdDescriptor(mode, nan, pooldim, window, padding, stride)
		if err != nil {
			return nil, err
		}
		return &Pooling{
			desc: desc,
		}, nil
	}
	if len(dims) < 4 {
		return nil, errors.New("Dims should be 4 or more")
	}
	desc, err := gocudnn.Pooling{}.NewPooling2dDescriptor(mode, nan, window, padding, stride)
	if err != nil {
		return nil, err
	}
	return &Pooling{
		desc: desc,
	}, nil
}
func 
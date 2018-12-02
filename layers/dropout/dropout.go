package dropout

import (
	"github.com/dereklstinson/GoCuNets/cudnn"
	"github.com/dereklstinson/GoCuNets/cudnn/dropout"
	"github.com/dereklstinson/GoCuNets/layers"
)

//Layer holds the op for the dropout
type Layer struct {
	op      *dropout.Ops
	dropout float32
	seed    uint64
	managed bool
}

//Settings is the settings for drop out layer
type Settings struct {
	Dropout float32 `json:"dropout,omitempty"`
	Seed    uint64  `json:"seed,omitempty"`
	Managed bool    `json:"managed,omitempty"`
}

//Setup sets up the layer
func Setup(handle *cudnn.Handler, x *layers.IO, drpout float32, seed uint64, managed bool) (*Layer, error) {
	op, err := dropout.Stage(handle, x.T(), drpout, seed, managed)
	return &Layer{
		op: op,
	}, err
}

//Preset presets the layer, but doesn't build it. Useful if you want to set up the network before the tensor descriptors for the input
//are made. handle can be nil.  I just wanted to keep it consistant.
func Preset(handle *cudnn.Handler, dropout float32, seed uint64, managed bool) (*Layer, error) {

	return &Layer{
		dropout: dropout,
		seed:    seed,
		managed: managed,
	}, nil
}

//BuildFromPreset will construct the dropout layer if preset was made. Both handle and input are needed.
//This can also be called again if input size has changed
func (l *Layer) BuildFromPreset(handle *cudnn.Handler, input *layers.IO) error {
	var err error
	/*	if l.op != nil {
			err = l.op.Destroy()
			if err != nil {
				return err
			}
		}
	*/
	l.op, err = dropout.Stage(handle, input.T(), l.dropout, l.seed, l.managed)
	return err
}

//ForwardProp does the forward propagation
func (l *Layer) ForwardProp(handle *cudnn.Handler, x, y *layers.IO) error {
	return l.op.ForwardProp(handle, x.T(), y.T())
}

//BackProp does the back propagation
func (l *Layer) BackProp(handle *cudnn.Handler, x, y *layers.IO) error {
	return l.op.BackProp(handle, x.DeltaT(), y.DeltaT())
}

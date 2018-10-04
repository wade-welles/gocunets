//Package trainer is a package that is used for training networks.  There is not much support for this yet.
//It will have vanilla and momentum on using a device. Its hard to build any kind of trainer using cudnn.
//the optensor is kind of limited.
package trainer

import (
	"github.com/dereklstinson/GoCuNets/layers"
	"github.com/dereklstinson/GoCudnn"
)

//Trainer will be used for updating weights.  Only momentum and adam are available right now
type Trainer interface {
	UpdateWeights(ctx gocudnn.Handler, weights *layers.IO) error
	L1L2Loss() (float32, float32, error)
}

func CreateTrainingMem(handle gocudnn.Handler, trainer Trainer, weights *layers.IO) error {

	switch x := trainer.(type) {
	case *Adam:

		return x.SetTrainingMem(handle, weights)
	case *Momentum:

		return x.SetTrainingMem(handle, weights)
	}

	return nil
}

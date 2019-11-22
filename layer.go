package gocunets

import (
	"errors"
	"fmt"

	"github.com/dereklstinson/GoCuNets/devices/gpu/nvidia/cudnn"
	"github.com/dereklstinson/GoCuNets/layers"
	"github.com/dereklstinson/GoCuNets/layers/activation"
	"github.com/dereklstinson/GoCuNets/layers/batchnorm"
	"github.com/dereklstinson/GoCuNets/layers/cnn"
	"github.com/dereklstinson/GoCuNets/layers/cnntranspose"
	"github.com/dereklstinson/GoCuNets/layers/dropout"
	"github.com/dereklstinson/GoCuNets/layers/pooling"
	"github.com/dereklstinson/GoCuNets/layers/reshape"
	"github.com/dereklstinson/GoCuNets/layers/softmax"
	"github.com/dereklstinson/GoCuNets/trainer"
)

type layer struct {
	name                          string
	activation                    *activation.Layer
	cnn                           *cnn.Layer
	softmax                       *softmax.Layer
	pool                          *pooling.Layer
	drop                          *dropout.Layer
	batch                         *batchnorm.Layer
	reshape                       *reshape.Layer
	cnntranspose                  *cnntranspose.Layer
	scalarnumalpha, scalarnumbeta int
}

func (l *layer) loadtrainer(handle *cudnn.Handler, trainers ...trainer.Trainer) error {
	if l.cnn != nil {
		if len(trainers) != 2 {
			fmt.Println(len(trainers))
			return fmt.Errorf("l.cnn got %d should get %d", len(trainers), 2)
		}
		return l.cnn.LoadTrainer(handle, trainers[0], trainers[1])
	}

	if l.batch != nil {
		if len(trainers) != 2 {
			return fmt.Errorf("l.batch got %d should get %d", len(trainers), 2)
		}
		return l.batch.LoadTrainer(handle, trainers[0], trainers[1])
	}
	if l.cnntranspose != nil {
		if len(trainers) != 2 {

			return fmt.Errorf("l.cnntranspose got %d should get %d", len(trainers), 2)

		}
		return l.cnntranspose.LoadTrainer(handle, trainers[0], trainers[1])
	}
	if l.activation != nil {
		tneed := l.activation.TrainersNeeded()
		if tneed > 0 {

			if len(trainers) != tneed {

				return fmt.Errorf("l.activation got %d should get %d", len(trainers), tneed)
			}
		}
		return l.activation.LoadTrainer(handle, trainers)
	}

	return errors.New("inbedded error doesn't support trainers")
}

func (l *layer) trainersneeded() int {
	if l.cnn != nil {
		return 2
	}
	if l.cnntranspose != nil {
		return 2
	}
	if l.batch != nil {
		return 2
	}
	if l.activation != nil {
		return l.activation.TrainersNeeded()

	}
	return 0

}

func wraplayer(input interface{}) (hidden *layer, ios int) {
	switch l := input.(type) {

	case *activation.Layer:
		if l.TrainersNeeded() > 0 {
			return &layer{
				activation: l,
				name:       "Activation",
			}, 1 + l.TrainersNeeded()
		}
		return &layer{
			activation: l,
			name:       "Activation",
		}, 1

	case *cnn.Layer:
		return &layer{
			cnn:  l,
			name: "CNN",
		}, 2

	case *softmax.Layer:
		return &layer{
			softmax: l,
			name:    "SoftMax",
		}, 1
	case *pooling.Layer:
		return &layer{
			pool: l,
			name: "Pooling",
		}, 1
	case *dropout.Layer:
		return &layer{
			drop: l,
			name: "DropOut",
		}, 1
	case *batchnorm.Layer:
		return &layer{
			batch: l,
			name:  "BatchNorm",
		}, 1
	case *reshape.Layer:
		return &layer{
			reshape: l,
			name:    "Reshape",
		}, 1
	case *cnntranspose.Layer:
		return &layer{
			cnntranspose: l,
			name:         "CNN-Transpose",
		}, 2

	default:
		return nil, -1
	}
}

func (l *layer) initalphascalarsamount() int {

	if l.cnn != nil {
		l.scalarnumalpha = l.cnn.NumAlphaScalars()
		return l.scalarnumalpha
	}

	if l.pool != nil {
		l.scalarnumalpha = l.pool.NumAlphaScalars()
		return l.scalarnumalpha

	}
	if l.drop != nil {

		return 0
	}
	if l.activation != nil {
		l.scalarnumalpha = l.activation.NumAlphaScalars()
		return l.scalarnumalpha

	}
	if l.batch != nil {
		l.scalarnumalpha = l.batch.NumAlphaScalars()
		return l.scalarnumalpha
	}

	if l.softmax != nil {
		l.scalarnumalpha = l.softmax.NumAlphaScalars()
		return l.scalarnumalpha

	}
	if l.reshape != nil {
		return 0
	}
	if l.cnntranspose != nil {
		l.scalarnumalpha = l.cnntranspose.NumAlphaScalars()
		return l.scalarnumalpha

	}
	return 0

}
func (l *layer) initbetascalarsamount() int {

	if l.cnn != nil {
		l.scalarnumbeta = l.cnn.NumBetaScalars()
		return l.scalarnumbeta
	}

	if l.pool != nil {
		l.scalarnumbeta = l.pool.NumBetaScalars()
		return l.scalarnumbeta

	}
	if l.drop != nil {

		return 0
	}
	if l.activation != nil {
		l.scalarnumbeta = l.activation.NumBetaScalars()
		return l.scalarnumbeta

	}
	if l.batch != nil {
		l.scalarnumbeta = l.batch.NumBetaScalars()
		return l.scalarnumbeta
	}

	if l.softmax != nil {
		l.scalarnumbeta = l.softmax.NumBetaScalars()
		return l.scalarnumbeta

	}
	if l.reshape != nil {
		return 0
	}
	if l.cnntranspose != nil {
		l.scalarnumbeta = l.cnntranspose.NumBetaScalars()
		return l.scalarnumbeta

	}
	return 0

}
func (l *layer) updateabetascalar(scalars []float64) (offset []float64) {
	if l.cnn != nil {

		l.cnn.SetBetaScalars(scalars[:l.scalarnumbeta])
		return scalars[l.scalarnumbeta:]
	}

	if l.pool != nil {
		l.pool.SetBetaScalars(scalars[:l.scalarnumbeta])
		return scalars[l.scalarnumbeta:]

	}
	if l.drop != nil {

		return scalars
	}
	if l.activation != nil {
		l.activation.SetBetaScalars(scalars[:l.scalarnumbeta])
		return scalars[l.scalarnumbeta:]

	}
	if l.batch != nil {
		l.batch.SetBetaScalars(scalars[:l.scalarnumbeta])
		return scalars[l.scalarnumbeta:]
	}

	if l.softmax != nil {
		l.softmax.SetBetaScalars(scalars[:l.scalarnumbeta])
		return scalars[l.scalarnumbeta:]

	}
	if l.reshape != nil {
		return scalars
	}
	if l.cnntranspose != nil {
		l.cnntranspose.SetBetaScalars(scalars[:l.scalarnumbeta])
		return scalars[l.scalarnumbeta:]

	}
	return scalars
}
func (l *layer) updatealphascalar(scalars []float64) (offset []float64) {
	if l.cnn != nil {

		l.cnn.SetAlphaScalars(scalars[:l.scalarnumalpha])
		return scalars[l.scalarnumalpha:]
	}

	if l.pool != nil {
		l.pool.SetAlphaScalars(scalars[:l.scalarnumalpha])
		return scalars[l.scalarnumalpha:]

	}
	if l.drop != nil {

		return scalars
	}
	if l.activation != nil {
		l.activation.SetAlphaScalars(scalars[:l.scalarnumalpha])
		return scalars[l.scalarnumalpha:]

	}
	if l.batch != nil {
		l.batch.SetAlphaScalars(scalars[:l.scalarnumalpha])
		return scalars[l.scalarnumalpha:]

	}

	if l.softmax != nil {
		l.softmax.SetAlphaScalars(scalars[:l.scalarnumalpha])
		return scalars[l.scalarnumalpha:]

	}
	if l.reshape != nil {
		return scalars
	}
	if l.cnntranspose != nil {
		l.cnntranspose.SetAlphaScalars(scalars[:l.scalarnumalpha])
		return scalars[l.scalarnumalpha:]

	}
	return scalars
}
func (l *layer) getoutputwithname(handle *cudnn.Handler, input *layers.IO) (*layers.IO, string, error) {

	if l.cnn != nil {
		x, err := l.cnn.MakeOutputTensor(handle, input)
		return x, "CNN-Output", err
	}

	if l.pool != nil {
		x, err := l.pool.MakeOutputLayer(handle, input)
		return x, "Pooling-Output", err
	}
	if l.drop != nil {

		err := l.drop.BuildFromPreset(handle, input)
		if err != nil {

			return nil, "", err
		}
		x, err := input.ZeroClone(handle)
		return x, "DropOut-Output", err
	}
	if l.activation != nil {
		x, err := input.ZeroClone(handle)
		return x, "Activation-Output", err
	}
	if l.batch != nil {
		err := l.batch.SetupPreset(handle, input)
		if err != nil {
			return nil, "", err
		}
		x, err := input.ZeroClone(handle)
		return x, "BatchNorm-Output", err
	}

	if l.softmax != nil {
		x, err := input.ZeroClone(handle)
		return x, "SoftMax-Output", err
	}
	if l.reshape != nil {
		x, err := l.reshape.MakeOutputTensor(handle, input)
		return x, "Reshape-Output", err
	}
	if l.cnntranspose != nil {
		x, err := l.cnntranspose.MakeOutputTensor(handle, input)
		return x, "CnnTranspose-Output", err
	}
	return nil, "", errors.New("Layer Needs Support")
}

/*
func (l *layer) getoutputdims(handle *cudnn.Handler, input *layers.IO) ([]int32, error) {

}
*/
func (l *layer) getoutput(handle *cudnn.Handler, input *layers.IO) (io *layers.IO, err error) {

	if l.cnn != nil {
		io, err = l.cnn.MakeOutputTensor(handle, input)
		if io == nil {
			fmt.Println("input is", input.T().Dims())

		}
		if err != nil {
			fmt.Println("Error in CNN Make Output Tensor input is:", input)
		}
		return io, err
	}
	if l.pool != nil {
		io, err = l.pool.MakeOutputLayer(handle, input)
		if io == nil {
			panic("IO IS NILL")
		}
		return io, err
	}
	if l.drop != nil {
		err = l.drop.BuildFromPreset(handle, input)
		if err != nil {
			return nil, err
		}
		io, err = input.ZeroClone(handle)
		if io == nil {
			panic("IO IS NILL")
		}
		return io, err

	}
	if l.activation != nil {
		io, err = input.ZeroClone(handle)
		if err != nil {
			fmt.Println("Error in activation Make Output Tensor input is:", input)
		}
		if io == nil {
			panic("IO IS NILL")
		}

		return io, err
	}
	if l.batch != nil {
		err := l.batch.SetupPreset(handle, input)
		if err != nil {
			fmt.Println("error in batch initialization")
			return nil, err
		}

		io, err = input.ZeroClone(handle)
		if err != nil {
			fmt.Println("Error in batch Make Output Tensor input is:", input)
		}
		if io == nil {
			panic("IO IS NILL")
		}

		return io, err
	}

	if l.softmax != nil {
		io, err = input.ZeroClone(handle)
		if io == nil {
			panic("IO IS NILL")
		}
		return io, err

	}
	if l.reshape != nil {
		io, err = l.reshape.MakeOutputTensor(handle, input)
		if io == nil {
			panic("IO IS NILL")
		}
		return io, err
	}
	if l.cnntranspose != nil {
		io, err = l.cnntranspose.MakeOutputTensor(handle, input)
		if err != nil {
			fmt.Println("DIMS Reverse", io.T().Dims())
			fmt.Println("Error in cnntranspose Make Output Tensor input is:", input)
		}
		if io == nil {
			panic("IO IS NILL")
		}
		return io, err
	}
	return nil, errors.New("Layer Needs Support")
}

//UpdateWeights updates the weights of layer
func (l *layer) updateWeights(handle *cudnn.Handler, batch int) error {

	if l.cnn != nil {
		return l.cnn.UpdateWeights(handle, batch)
	}

	if l.cnntranspose != nil {
		return l.cnntranspose.UpdateWeights(handle, batch)

	}
	if l.batch != nil {
		return l.batch.UpdateWeights(handle, batch)
	}
	if l.activation != nil {
		if l.activation.TrainersNeeded() > 0 {
			return l.activation.UpdateWeights(handle, batch)

		}

	}
	return nil
}

func (l *layer) l1l2loss() (l1, l2 float32) {

	if l.cnn != nil {
		return l.cnn.L1L2Loss()
	}

	if l.cnntranspose != nil {
		return l.cnntranspose.L1L2Loss()

	}
	return -123, -123
}

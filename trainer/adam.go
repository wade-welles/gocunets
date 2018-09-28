package trainer

import (
	"errors"

	"github.com/dereklstinson/GoCuNets/layers"
	gocudnn "github.com/dereklstinson/GoCudnn"
)

type Adam struct {
	loss1     float32
	loss2     float32
	gpuloss1  gocudnn.Memer
	gpuloss2  gocudnn.Memer
	gsum      gocudnn.Memer
	xsum      gocudnn.Memer
	trainer   *gocudnn.TrainerD
	params    gocudnn.TrainingParams
	regparams gocudnn.RegParams
}

const defaultadambeta1 = 0.9
const defaultadambeta2 = 0.999
const defaultadameps = float32(1e-8)
const defaultadamrate = .001

func (a *Adam) SetTrainingMem(ctx gocudnn.Contexter, weights *layers.IO) error {
	_, err := ctx.GetTrainHandle()
	if err != nil {
		return err
	}
	_, dtype, dims, err := weights.Properties()
	if err != nil {
		return err
	}
	DeFault := gocudnn.MemcpyKindFlag{}.Default()
	Global := gocudnn.ManagedMemFlag{}.Global()

	switch dtype {

	case gocudnn.DataTypeFlag{}.Float():

		asize := dimsize(dims)
		x := make([]float32, asize)
		sizet, err := gocudnn.FindSizeT(x)
		if err != nil {
			return err
		}
		xp, err := gocudnn.MakeGoPointer(x)
		if err != nil {
			return err
		}
		a.gsum, err = gocudnn.MallocManaged(sizet, Global)
		if err != nil {
			return err
		}
		a.xsum, err = gocudnn.MallocManaged(sizet, Global)
		if err != nil {
			return err
		}
		err = gocudnn.CudaMemCopy(a.gsum, xp, sizet, DeFault)
		if err != nil {
			return err
		}
		err = gocudnn.CudaMemCopy(a.xsum, xp, sizet, DeFault)
		if err != nil {
			return err
		}
		a.gpuloss1, err = gocudnn.MallocManaged(gocudnn.SizeT(4), Global)
		if err != nil {
			return err
		}
		a.gpuloss2, err = gocudnn.MallocManaged(gocudnn.SizeT(4), Global)
		if err != nil {
			return err
		}

	default:

		return errors.New("Only Float datatype supported at the moment")
	}
	return nil
}

func (a *Adam) UpdateWeights(ctx gocudnn.Contexter, weights *layers.IO) error {
	blocksize := uint32(32)
	tctx, err := ctx.GetTrainHandle()
	if err != nil {
		return err
	}
	err = a.trainer.L1L2Regularization(tctx, blocksize, weights.DeltaT().Memer(), weights.T().Memer(), a.gpuloss1, a.gpuloss2, a.regparams)
	if err != nil {
		return err
	}
	return a.trainer.TrainValues(tctx, blocksize, weights.DeltaT().Memer(), weights.T().Memer(), a.gsum, a.xsum, a.params)
}
func (a *Adam) L1L2Loss() (float32, float32, error) {
	kind := gocudnn.MemcpyKindFlag{}.Default()
	size := gocudnn.SizeT(4)
	l1, err := gocudnn.MakeGoPointer(a.loss1)
	if err != nil {
		return 0, 0, err
	}
	l2, err := gocudnn.MakeGoPointer(a.loss2)
	if err != nil {
		return 0, 0, err
	}
	err = gocudnn.CudaMemCopy(l1, a.gpuloss1, size, kind)
	err = gocudnn.CudaMemCopy(l2, a.gpuloss2, size, kind)
	return a.loss1, a.loss2, nil
}
func dimsize(dims []int32) int32 {
	x := int32(1)
	for i := 0; i < len(dims); i++ {
		x *= dims[i]
	}
	return x
}
func SetupAdam(ctx gocudnn.Contexter, decay1, decay2 float32, batch int) (*Adam, error) {
	tctx, err := ctx.GetTrainHandle()
	if err != nil {
		return nil, err
	}
	l1l2 := gocudnn.RegularizationFlag{}.L1L2()
	adam := gocudnn.TrainingModeFlag{}.Adam()
	t, err := gocudnn.Xtra{}.NewTrainingDescriptor(tctx, adam, gocudnn.DataTypeFlag{}.Float(), l1l2)
	if err != nil {
		return nil, err
	}
	reg := gocudnn.Xtra{}.CreateRegParamsFloat32(decay1, decay2, float32(batch))
	x := gocudnn.Xtra{}.CreateParamsFloat32(defaultadameps, defaultadamrate, defaultadambeta1, defaultadambeta2)

	return &Adam{
		trainer:   t,
		params:    x,
		regparams: reg,
	}, nil
}

//SetDecay1 sets decay1
func (a *Adam) SetDecay1(decay1 float32) {
	a.regparams.SetDecay1(decay1)
}

//SetDecay2 sets decay 2
func (a *Adam) SetDecay2(decay2 float32) {
	a.regparams.SetDecay2(decay2)

}

//SetBeta1 sets beta1
func (a *Adam) SetBeta1(beta1 float32) {
	a.params.SetBeta1(beta1)
}

//SetBeta2 sets beta2
func (a *Adam) SetBeta2(beta2 float32) {
	a.params.SetBeta2(beta2)

}

//SetRate sets rate
func (a *Adam) SetRate(rate float32) {
	a.params.SetRate(rate)

}
func (a *Adam) SetBatch(batch float32) {
	a.regparams.SetBatch(batch)
}

//SetEps sets eps
func (a *Adam) SetEps(eps float32) {
	a.params.SetEps(eps)

}

/*
func (a *Adam)Loss1()float32{
	return a.loss1
}

*/

//CreateAdamHandle creates a handle for adam
func CreateAdamHandle(dev *gocudnn.Device, kerneldir string) (*gocudnn.TrainHandle, error) {
	var x gocudnn.Xtra
	return x.MakeTrainingHandle(kerneldir, dev)

}
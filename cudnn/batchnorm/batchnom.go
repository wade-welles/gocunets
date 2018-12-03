package batchnorm

import (
	"errors"

	"github.com/dereklstinson/GoCuNets/cudnn"
	"github.com/dereklstinson/GoCuNets/cudnn/tensor"
	gocudnn "github.com/dereklstinson/GoCudnn"
)

//Ops contains the operation of batchnorm.
type Ops struct {
	epsilon float64
	mode    gocudnn.BatchNormMode
	bnsbmvd *gocudnn.TensorD
	scale   *gocudnn.Malloced
	bias    *gocudnn.Malloced
	rrm     *gocudnn.Malloced
	rrv     *gocudnn.Malloced
	rsm     *gocudnn.Malloced
	rsv     *gocudnn.Malloced
	rbnscd  *gocudnn.Malloced
	rbnbd   *gocudnn.Malloced
}

func buildfromdesc(handle *cudnn.Handler, desc *gocudnn.TensorD, managed bool) (*gocudnn.Malloced, error) {
	var tfuncs gocudnn.Tensor
	dtype, _, _, err := desc.GetDescrptor()
	if err != nil {
		return nil, err
	}
	sizet, err := desc.GetSizeInBytes()
	if err != nil {
		return nil, err
	}
	if managed == true {
		gpumem, err := gocudnn.UnifiedMangedGlobal(sizet)
		if err != nil {
			return nil, err
		}
		zero := gocudnn.CScalarByDataType(dtype, 0.0)

		err = tfuncs.SetTensor(handle.Cudnn(), desc, gpumem, zero)
		if err != nil {
			gpumem.Free()
			return nil, err
		}
		return gpumem, err

	}
	gpumem, err := gocudnn.Malloc(sizet)
	if err != nil {
		return nil, err
	}
	zero := gocudnn.CScalarByDataType(dtype, 0.0)
	err = tfuncs.SetTensor(handle.Cudnn(), desc, gpumem, zero)
	if err != nil {
		gpumem.Free()
		return nil, err
	}
	return gpumem, err

}
func errorstacker(original, newerr error) error {
	x := newerr.Error()
	y := original.Error()
	return errors.New(x + "..." + y)
}

//Free Frees the mem
func (o *Ops) Free() error {
	var err error
	var errstack error
	err = o.rbnbd.Free()
	if err != nil {
		errstack = errorstacker(errstack, err)
	}
	err = o.rbnscd.Free()
	if err != nil {
		errstack = errorstacker(errstack, err)
	}
	err = o.rrm.Free()
	if err != nil {
		errstack = errorstacker(errstack, err)
	}
	err = o.rrv.Free()
	if err != nil {
		errstack = errorstacker(errstack, err)
	}
	err = o.rsm.Free()
	if err != nil {
		errstack = errorstacker(errstack, err)
	}
	err = o.rsv.Free()
	if err != nil {
		errstack = errorstacker(errstack, err)
	}
	err = o.bnsbmvd.DestroyDescriptor()
	if err != nil {
		errstack = errorstacker(errstack, err)
	}
	return errstack
}

//PreStageSpatial Normalization is performed over N+spatial dimensions.
//This mode is intended for use after convolutional layers (where spatial invariance is desired).
//In this mode the bnBias, bnScale tensor dimensions are 1xCx1x1.
func PreStageSpatial(handle *cudnn.Handler) (*Ops, error) {
	var x gocudnn.BatchNormModeFlag
	return &Ops{
		mode: x.Spatial(),
	}, nil
}

/*PreStageSpatialPersistant - similar to spatial but can be faster

An optimized path may be selected for CUDNN_DATA_FLOAT and CUDNN_DATA_HALF types, compute capability 6.0 or higher for the following two batch normalization API calls: cudnnBatchNormalizationForwardTraining(), and cudnnBatchNormalizationBackward().
In the case of cudnnBatchNormalizationBackward(), the savedMean and savedInvVariance arguments should not be NULL.

The rest of this section applies for NCHW mode only:

This mode may use a scaled atomic integer reduction that is deterministic but imposes more restrictions on the input data range.
When a numerical overflow occurs the algorithm may produce NaN-s or Inf-s (infinity) in output buffers.

When Inf-s/NaN-s are present in the input data, the output in this mode is the same as from a pure floating-point implementation.

For finite but very large input values, the algorithm may encounter overflows more frequently due to a lower dynamic range and emit Inf-s/NaN-s while CUDNN_BATCHNORM_SPATIAL will produce finite results.
The user can invoke cudnnQueryRuntimeError() to check if a numerical overflow occurred in this mode.
*/
func PreStageSpatialPersistant(handle *cudnn.Handler) (*Ops, error) {
	var x gocudnn.BatchNormModeFlag
	return &Ops{
		mode: x.SpatialPersistent(),
	}, nil
}

//PreStagePerActivation Normalization is performed per-activation. This mode is intended to be used after non-convolutional network layers.
//In this mode the tensor dimensions of bnBias and bnScale, the parameters used in the cudnnBatchNormalization* functions, are 1xCxHxW.
func PreStagePerActivation(handle *cudnn.Handler) (*Ops, error) {
	var x gocudnn.BatchNormModeFlag
	return &Ops{
		mode: x.PerActivation(),
	}, nil
}

//Stage stages the bachnorm op. It also builds the memory for it so you don't have to worry about it.
func Stage(handle *cudnn.Handler, x *tensor.Volume, mode gocudnn.BatchNormMode, managed bool) (*Ops, error) {

	bnd, err := gocudnn.BatchNorm{}.DeriveBNTensorDescriptor(x.TD(), mode)
	if err != nil {
		return nil, err
	}

	rrm, err := buildfromdesc(handle, bnd, managed)
	if err != nil {
		return nil, err
	}
	rrv, err := buildfromdesc(handle, bnd, managed)
	if err != nil {
		rrm.Free()
		return nil, err
	}
	rsm, err := buildfromdesc(handle, bnd, managed)
	if err != nil {
		rrv.Free()
		rrm.Free()
		return nil, err
	}
	rsv, err := buildfromdesc(handle, bnd, managed)
	if err != nil {
		rsm.Free()
		rrv.Free()
		rrm.Free()
		return nil, err
	}
	rbnscd, err := buildfromdesc(handle, bnd, managed)
	if err != nil {
		rsm.Free()
		rsv.Free()
		rrv.Free()
		rrm.Free()
		return nil, err
	}
	rbnbd, err := buildfromdesc(handle, bnd, managed)
	if err != nil {
		rbnscd.Free()
		rsm.Free()
		rsv.Free()
		rrv.Free()
		rrm.Free()
		return nil, err
	}
	bias, err := buildfromdesc(handle, bnd, managed)
	if err != nil {
		rbnscd.Free()
		rsm.Free()
		rsv.Free()
		rrv.Free()
		rrm.Free()
		rbnbd.Free()
		return nil, err
	}
	scale, err := buildfromdesc(handle, bnd, managed)
	if err != nil {
		rbnscd.Free()
		rsm.Free()
		rsv.Free()
		rrv.Free()
		rrm.Free()
		rbnbd.Free()
		bias.Free()
		return nil, err
	}

	return &Ops{
		bnsbmvd: bnd,
		mode:    mode,
		rrm:     rrm,
		rrv:     rrv,
		rsm:     rsm,
		rsv:     rsv,
		rbnbd:   rbnbd,
		rbnscd:  rbnscd,
		scale:   scale,
		bias:    bias,
	}, nil

}

//ForwardTraining is used for the forward training
func (o *Ops) ForwardTraining(handle *cudnn.Handler, alpha, beta, averagingfactor, epsilon float64, x, y *tensor.Volume) error {
	_, dtype, _, err := x.Properties()
	if err != nil {
		return err
	}
	a := gocudnn.CScalarByDataType(dtype.Cu(), alpha)
	b := gocudnn.CScalarByDataType(dtype.Cu(), beta)
	return gocudnn.BatchNorm{}.Funcs.BatchNormalizationForwardTraining(handle.Cudnn(),
		o.mode,
		a,
		b,
		x.TD(),
		x.Memer(),
		y.TD(),
		y.Memer(),
		o.bnsbmvd,
		o.scale,
		o.bias,
		averagingfactor,
		o.rrm,
		o.rrv,
		epsilon,
		o.rsm,
		o.rsv,
	)

}

//BackwardProp is used for the forward training
func (o *Ops) BackwardProp(handle *cudnn.Handler, alphaparam, betaparam, alphadata, betadata, epsilon float64, x, dx, dy *tensor.Volume) error {
	_, dtype, _, err := x.Properties()
	if err != nil {
		return err
	}
	ad := gocudnn.CScalarByDataType(dtype.Cu(), alphadata)
	bd := gocudnn.CScalarByDataType(dtype.Cu(), betadata)
	ap := gocudnn.CScalarByDataType(dtype.Cu(), alphaparam)
	bp := gocudnn.CScalarByDataType(dtype.Cu(), betaparam)
	return gocudnn.BatchNorm{}.Funcs.BatchNormalizationBackward(handle.Cudnn(),
		o.mode,
		ad,
		bd,
		ap,
		bp,
		x.TD(),
		x.Memer(),
		dx.TD(),
		dx.Memer(),
		dy.TD(),
		dy.Memer(),
		o.bnsbmvd,
		o.scale,
		o.rbnscd,
		o.rbnbd,
		epsilon,
		o.rsm,
		o.rsv,
	)

}

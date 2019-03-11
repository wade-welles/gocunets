package xloss

import (
	"github.com/dereklstinson/GoCuNets/devices/gpu/nvidia/cudnn/tensor"
	"github.com/dereklstinson/GoCudnn/xtra"
)

//Ops contains the operation to do the loss function
type Ops struct {
	desc *xtra.XLossD
	loss float32
}

//Stage stages the loss operation
func Stage(handle *xtra.Handle, mode xtra.XLossMode, managed bool) (*Ops, error) {

	xloss, err := xtra.NewLossDescriptor(handle, mode)
	if err != nil {
		return nil, err
	}
	return &Ops{
		desc: xloss,
	}, nil
}
func (o *Ops) Error(handle *xtra.Handle, dx, y, dy *tensor.Volume) error {
	var err error
	o.loss, err = o.desc.CalculateErrorAndLoss(handle, dx.TD(), dx.Memer(), y.TD(), y.Memer(), dy.TD(), dy.Memer())
	return err
}

//Loss returns the loss function found. Should be called after the error is found
func (o *Ops) Loss() float32 {
	return o.loss
}
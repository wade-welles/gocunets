package gocunets

import (
	"errors"
	"fmt"

	"github.com/dereklstinson/GoCuNets/devices/gpu/nvidia/cudnn/tensor"
)

//Module is a wrapper around a neural network set of operations
type Module interface {
	ID() int64
	Forward() error
	Backward() error
	Update(counter int) error //counter can count updates or it can count epochs.  I found updates to work best.
	FindOutputDims() ([]int32, error)
	Inference() error
	InitHiddenLayers(decay1, decay2 float32) (err error)
	InitWorkspace() (err error)
	GetTensorX() (x *Tensor)
	GetTensorDX() (dx *Tensor)
	GetTensorY() (y *Tensor)
	GetTensorDY() (dy *Tensor)
	SetTensorX(x *Tensor)
	SetTensorDX(dx *Tensor)
	SetTensorY(y *Tensor)
	SetTensorDY(dy *Tensor)
	//GetWeights()([]*Tensor)
	//GetTrainers()([]*Tensor)
}

var moduleforwarddebugging bool
var modulebackwarddatadebugging bool
var modulebackwardfilterdebugging bool
var moduleconcatdebugging bool
var moduleactivationdebugging bool

//ModuleActivationDebug is for debugging
func ModuleActivationDebug() {
	moduleactivationdebugging = true
}

//ModuleConcatDebug is for debugging
func ModuleConcatDebug() {
	moduleconcatdebugging = true
}

//ModuleForwardDebug is for debugging
func ModuleForwardDebug() {
	moduleforwarddebugging = true
}

//ModuleBackwardDataDebug is for debugging
func ModuleBackwardDataDebug() {
	modulebackwarddatadebugging = true
}

//ModuleBackwardFilterDebug is for debugging
func ModuleBackwardFilterDebug() {
	modulebackwardfilterdebugging = true
}

//ModuleDebugMode sets a flag that prints outputs of inner module outputs

//This func is way to complicated.
//But if the filterdim is odd.  It is suited for odd in odd out.
//With an odd input being 2^n + 1 output will be 2^(n-1) +1.
//If the filter dim is even then it is suited for even in even out. With an even input being 2^n. output will be 2^(n-1)
//it only does stride 2 for right now.  It will do a stride of 1 in a little while when I get my spreadsheet out.

func recommendedpaddilation(filterdim, index, stride, offset int32) (dilation, pad int32, err error) {
	if filterdim%2 == 0 {
		if stride == 1 {
			dilation = 2 * (index + 1)
			if (((filterdim-1)*dilation + 1 + offset) % 2) != 0 {
				return -1, -1, fmt.Errorf("(((filterdim-1)*dilation +1 + offset) modual 2) != 0, (((%v-1)*%v)+1+%v)", filterdim, dilation, offset)
			}
			pad = ((filterdim-1)*dilation + 1 + offset) / 2

		} else {
			dilation = 2*index + 1
			pad = ((filterdim-1)*dilation + 1 + offset) / 2
		}

	} else {
		dilation = index + 1
		pad = ((filterdim-1)*dilation + 1 + offset) / 2
	}

	if pad < 0 {
		return -1, -1, errors.New("recommendedpaddilation params givin give pad< 0")
	}
	return dilation, pad, nil

}
func checkparamsconv(i, f, p, s, d int32) (isgood bool) {
	if (i+2*p-(((f-1)*d)+1))%s == 0 {
		isgood = true
	}
	return isgood
}

func dimoutput(i, f, p, s, d int32) (o int32) {
	////if p = (((f - 1) * d) + 1+offset)/2  &&  s=2 && f=2
	o = 1 + (i+2*p-(((f-1)*d)+1))/s
	return o
}
func dimoutputreverse(i, f, p, s, d int32) (o int32) {
	o = (i-1)*s - 2*p + ((f - 1) * d) + 1
	//if p = (((f - 1) * d) + 1+offset)/2  &&  s=2 && f=2
	//then p = (d+1 + offset)/2
	//then o = 2(i-1) - (d+1 + offset)/2 + d+1
	//
	return o
}

//SimpleModuleNetwork is a simple module network
type SimpleModuleNetwork struct {
	id             int64
	decay1, decay2 float32
	C              *Concat
	Modules        []Module
	Output         *OutputModule
	Classifier     *ClassifierModule
	b              *Builder

	//	x, dx, y, dy        *Tensor
	//	firstinithiddenfirstinithidden    bool
	//	firstinitworkspace bool
	//	firstfindoutputdims bool
}

//CreateSimpleModuleNetwork a simple module network
func CreateSimpleModuleNetwork(id int64, b *Builder) (smn *SimpleModuleNetwork) {
	smn = new(SimpleModuleNetwork)
	smn.b = b
	smn.id = id

	return smn
}

//SetMSEClassifier needs to be made
func (m *SimpleModuleNetwork) SetMSEClassifier() (err error) {
	return errors.New("(m *SimpleModuleNetwork) SetMSEClassifier() needs to be made")
}

//SetSoftMaxClassifier sets the classifier module it should be added last.
//Should be ran after OutputModule is set
func (m *SimpleModuleNetwork) SetSoftMaxClassifier() (err error) { //(y, dy *Tensor, err error) {

	lastmod := m.Output
	if lastmod.GetTensorDX() == nil {
		lastmod.SetTensorDX(m.Modules[len(m.Modules)-1].GetTensorDY())
	}
	if lastmod.GetTensorX() == nil {
		lastmod.SetTensorX(m.Modules[len(m.Modules)-1].GetTensorY())
	}
	if lastmod.GetTensorDY() == nil {
		lmoutputdims, err := lastmod.FindOutputDims()
		if err != nil {
			return err
		}
		lmdy, err := m.b.CreateTensor(lmoutputdims)
		if err != nil {
			return err
		}
		lastmod.SetTensorDY(lmdy)
	}
	if lastmod.GetTensorY() == nil {
		lmoutputdims, err := lastmod.FindOutputDims()
		if err != nil {
			return err
		}
		lmy, err := m.b.CreateTensor(lmoutputdims)
		if err != nil {
			return err
		}
		lastmod.SetTensorY(lmy)
	}
	lmoutputdims, err := lastmod.FindOutputDims()
	if err != nil {
		return err
	}
	y, err := m.b.CreateTensor(lmoutputdims)
	if err != nil {
		return err
	}
	dy, err := m.b.CreateTensor(lmoutputdims)
	if err != nil {
		return err
	}

	m.Classifier, err = CreateSoftMaxClassifier(lastmod.ID()+1, m.b, lastmod.GetTensorY(), lastmod.GetTensorDY(), y, dy)
	if err != nil {
		return err
	}
	return nil
}

//SetModules sets modules
func (m *SimpleModuleNetwork) SetModules(modules []Module) {
	m.Modules = modules
}

//ID satisfies Module interface
func (m *SimpleModuleNetwork) ID() int64 {
	return m.id
}

//GetTensorX Gets x tensor
func (m *SimpleModuleNetwork) GetTensorX() *Tensor {
	if m.Modules[0] != nil {
		return m.Modules[0].GetTensorX()
	}
	return nil
}

//GetTensorDX Gets dx tensor
func (m *SimpleModuleNetwork) GetTensorDX() *Tensor {
	if m.Modules[0] != nil {
		return m.Modules[0].GetTensorDX()

	}
	return nil
}

//GetTensorY Gets y tensor
func (m *SimpleModuleNetwork) GetTensorY() *Tensor {
	if m.Classifier != nil {
		return m.Classifier.GetTensorY()
		//	return m.y
	} else if m.Output != nil {
		return m.Output.GetTensorY()
	}
	if m.Modules != nil {
		return m.Modules[len(m.Modules)-1].GetTensorY()
	}

	return nil
}

//GetTensorDY Gets dy tensor
func (m *SimpleModuleNetwork) GetTensorDY() *Tensor {
	if m.Classifier != nil {
		return m.Classifier.GetTensorDY()
		//	return m.dy
	} else if m.Output != nil {
		return m.Output.GetTensorDY()
	} else if m.Modules != nil {
		return m.Modules[len(m.Modules)-1].GetTensorDY()
	}
	return nil
	//return m.dy
}

//SetTensorX sets x tensor
func (m *SimpleModuleNetwork) SetTensorX(x *Tensor) {
	//	m.x = x
	if m.Modules != nil {
		m.Modules[0].SetTensorX(x)
	}

}

//SetTensorDX sets dx tensor
func (m *SimpleModuleNetwork) SetTensorDX(dx *Tensor) {
	//	m.dx = dx
	if m.Modules != nil {
		m.Modules[0].SetTensorDX(dx)
	}
}

//SetTensorY sets y tensor
func (m *SimpleModuleNetwork) SetTensorY(y *Tensor) {
	//	m.y = y
	if m.Classifier != nil {
		m.Classifier.SetTensorY(y)
	} else if m.Output != nil {
		m.Output.SetTensorY(y)
	} else if len(m.Modules) > 0 {
		m.Modules[len(m.Modules)-1].SetTensorY(y)
	}

}

//SetTensorDY sets dy tensor
func (m *SimpleModuleNetwork) SetTensorDY(dy *Tensor) {
	//	m.dy = dy
	if m.Classifier != nil {
		m.Classifier.SetTensorDY(dy)
	} else if m.Output != nil {
		m.Output.SetTensorDY(dy)
	} else if len(m.Modules) > 0 {
		m.Modules[len(m.Modules)-1].SetTensorDY(dy)
	}

}

//InitHiddenLayers satisfies the Module interface
func (m *SimpleModuleNetwork) InitHiddenLayers(decay1, decay2 float32) (err error) {
	m.decay1, m.decay2 = decay1, decay2
	if m.Modules == nil {
		return fmt.Errorf("(m *SimpleModuleNetwork) InitHiddenLayers: %s", "Modules are nil")
	}
	//if m.x == nil {
	//	return fmt.Errorf("(m *SimpleModuleNetwork) InitHiddenLayers: %s", "TensorX is nil")
	//}

	if m.Modules[0].GetTensorY() == nil {
		_, err = m.FindOutputDims() //m.FindOutputDims creates connections between Modules
		if err != nil {
			return fmt.Errorf("(m *SimpleModuleNetwork) InitHiddenLayers: %v", err)
		}

	}
	for i, mod := range m.Modules {
		err = mod.InitHiddenLayers(decay1, decay2)
		if err != nil {
			return fmt.Errorf("(m *SimpleModuleNetwork) InitHiddenLayers: index %v\n %v", i, err)
		}
	}

	err = m.Output.InitHiddenLayers(decay1, decay2)
	if err != nil {
		return fmt.Errorf("(m *SimpleModuleNetwork) InitHiddenLayers: m.Output: %v", err)
	}
	//m.firstinithidden = true
	return nil
}

//InitWorkspace inits workspace
func (m *SimpleModuleNetwork) InitWorkspace() (err error) {
	for i, mod := range m.Modules {
		err = mod.InitWorkspace()
		if err != nil {
			return fmt.Errorf("(m *SimpleModuleNetwork) InitWorkspace: index: %v\n err: %v", i, err)
		}
	}
	err = m.Output.InitWorkspace()
	if err != nil {
		return fmt.Errorf("(m *SimpleModuleNetwork) InitWorkspace: m.Output: %v", err)
	}
	//	m.firstinitworkspace = true
	return nil
}

//FindOutputDims satisifis the Module interface
//
//Have to run (m *SimpleModuleNetwork)SetTensorX().  If module network requres backpropdata to go to another module network.
//Then also run (m *SimpleModuleNetwork)SetTensorDX()
func (m *SimpleModuleNetwork) FindOutputDims() (dims []int32, err error) {
	//	if m.x == nil {
	//		return nil, errors.New("(m *SimpleModuleNetwork) FindOutputDims: TensorX hasn't been set")
	//	}
	if m.Modules == nil {
		return nil, errors.New("(m *SimpleModuleNetwork) FindOutputDims: No Modules have been set")
	}
	//if m.Output != nil {
	//	return m.Output.FindOutputDims()
	//}
	var px = m.GetTensorX()
	if px == nil {
		return nil, errors.New("(m *SimpleModuleNetwork) FindOutputDims: First Module's input not set")
	}
	var pdx = m.GetTensorDX()
	for _, mod := range m.Modules {
		if mod.GetTensorX() == nil {
			mod.SetTensorX(px)
		} else {
			if mod.GetTensorX() != px {
				panic("SHould be the same")
			}
		}
		if mod.GetTensorDX() == nil {
			mod.SetTensorDX(pdx)
		} else {
			if mod.GetTensorDX() != pdx {
				panic("SHould be the same")
			}
		}

		outputdims, err := mod.FindOutputDims()
		if err != nil {
			return nil, err
		}
		if mod.GetTensorY() == nil {
			px, err = m.b.CreateTensor(outputdims)
			if err != nil {
				return nil, err
			}
			mod.SetTensorY(px)
		} else {
			px = mod.GetTensorY()
		}
		if mod.GetTensorDY() == nil {
			pdx, err = m.b.CreateTensor(outputdims)
			if err != nil {
				return nil, err
			}
			mod.SetTensorDY(pdx)
		} else {
			pdx = mod.GetTensorDY()
		}

	}

	outputdims, err := m.Modules[len(m.Modules)-1].FindOutputDims()
	if m.Output == nil {
		return outputdims, err
	}
	if m.Output.GetTensorX() == nil {
		m.Output.SetTensorX(px)
	} else {
		if m.Output.GetTensorX() != px {
			panic("SHould be the same")
		}
	}
	if m.Output.GetTensorDX() == nil {
		m.Output.SetTensorDX(pdx)
	} else {
		if m.Output.GetTensorDX() != pdx {
			panic("SHould be the same")
		}
	}
	if m.Output.GetTensorY() == nil {
		px, err = m.b.CreateTensor(outputdims)
		if err != nil {
			return nil, err
		}
		m.Output.SetTensorY(px)
	} else {
		px = m.Output.GetTensorY()
	}
	if m.Output.GetTensorDY() == nil {
		pdx, err = m.b.CreateTensor(outputdims)
		if err != nil {
			return nil, err
		}
		m.Output.SetTensorDY(pdx)
	} else {
		pdx = m.Output.GetTensorDY()
	}
	outputdims, err = m.Output.FindOutputDims()
	if m.Classifier == nil {
		return outputdims, nil
	}
	if m.Classifier.GetTensorX() == nil {
		m.Classifier.SetTensorX(px)
	} else {
		if px != m.Classifier.GetTensorX() {
			panic("Should be the same")
		}
	}
	if m.Classifier.GetTensorDX() == nil {
		m.Classifier.SetTensorDX(pdx)
	} else {
		if pdx != m.Classifier.GetTensorDX() {
			panic("Should be the same")
		}
	}

	return outputdims, nil
}

//Forward does a forward without a concat
func (m *SimpleModuleNetwork) Forward() (err error) {
	for i := range m.Modules {
		err = m.Modules[i].Forward()
		if err != nil {
			return err
		}
	}
	err = m.Output.Forward()
	if err != nil {
		return err
	}
	if m.Classifier == nil {
		return nil
	}
	return m.Classifier.PerformError()
}

//Update updates the hidden weights
//Update can count epochs or updates.  I found counting updates works the best.
func (m *SimpleModuleNetwork) Update(counter int) (err error) {
	err = m.Output.Update(counter)
	if err != nil {
		return err
	}
	for i := range m.Modules {
		if i == 0 {
			//	trainer.DebuggingAdam()
		}
		err = m.Modules[i].Update(counter)
		if err != nil {
			return err
		}
	}
	return nil
}

//BackPropForSharedInputForModuleNetworks is a hack to make up if two module networks share the same input.
//It will zero out the dx values for the module and then run back propagation
func BackPropForSharedInputForModuleNetworks(m []*SimpleModuleNetwork) (err error) {
	err = m[0].GetTensorDX().SetValues(m[0].b.h.Handler, 0)
	if err != nil {
		return err
	}
	for i := range m {
		err = m[i].Backward()
		if err != nil {
			return err
		}

	}
	return nil
}

//GetLoss returns the loss found.
func (m *SimpleModuleNetwork) GetLoss() float32 {
	return m.Classifier.GetAverageBatchLoss()
}

//Backward does a forward without a concat
func (m *SimpleModuleNetwork) Backward() (err error) {

	err = m.Output.Backward()
	if err != nil {
		return err
	}
	for i := len(m.Modules) - 1; i >= 0; i-- {

		err = m.Modules[i].Backward()
		if err != nil {
			return err
		}
	}

	return nil
}

//Inference does a forward without a concat
func (m *SimpleModuleNetwork) Inference() (err error) {
	for i := range m.Modules {
		err = m.Modules[i].Inference()
		if err != nil {
			return err
		}
	}
	if m.Output != nil {
		m.Output.Inference()
	}
	if m.Classifier == nil {
		return nil
	}
	return m.Classifier.Inference()
}

//TestForward does the forward prop but it still calculates loss for testing
func (m *SimpleModuleNetwork) TestForward() (err error) {
	for i := range m.Modules {
		err = m.Modules[i].Inference()
		if err != nil {
			return err
		}
	}
	if m.Output != nil {
		return m.Output.Inference()
	}
	if m.Classifier != nil {
		return m.Classifier.TestForward()
	}
	return nil
}

//ForwardCustom does a custom forward function
func (m *SimpleModuleNetwork) ForwardCustom(forward func() error) (err error) {
	return forward()
}

//BackwardCustom does a custom backward function
func (m *SimpleModuleNetwork) BackwardCustom(backward func() error) (err error) {
	return backward()
}

//Forwarder does the forward operation
type Forwarder interface {
	Forward() error
}

//Backwarder does the backward operation
type Backwarder interface {
	Backward() error
}

//Updater does the update interface
type Updater interface {
	Update() error
}

//ReverseConcat is just a simple solution to split a source into multiple dests. If the source channel is not divisible by
//the number of dests then the remainder of the sources channels will be set into the last dest.
type ReverseConcat struct {
	c          *tensor.Concat
	h          *Handle
	dests      []*tensor.Volume
	deltadests []*tensor.Volume
	src        *tensor.Volume
	deltasrc   *tensor.Volume
}

//CreateReverseConcat creates a reverse concat
func CreateReverseConcat(h *Handle) (c *ReverseConcat, err error) {
	c = new(ReverseConcat)
	c.c, err = tensor.CreateConcat(h.Handler)
	c.h = h
	return c, err
}

//SetOutputDeltaDests sets the delta dests for back propagation
func (c *ReverseConcat) SetOutputDeltaDests(deltadests []*Tensor) {
	c.deltadests = make([]*tensor.Volume, len(deltadests))
	for i := range deltadests {
		c.deltadests[i] = deltadests[i].Volume
	}
	return
}

//SetOutputDests sets the output dests
func (c *ReverseConcat) SetOutputDests(dests []*Tensor) {
	c.dests = make([]*tensor.Volume, len(dests))
	for i := range dests {
		c.dests[i] = dests[i].Volume
	}
	return
}

//SetInputSource sets the input source
func (c *ReverseConcat) SetInputSource(src *Tensor) {
	c.src = src.Volume
}

//SetInputDeltaSource sets the delta src for back propagation
func (c *ReverseConcat) SetInputDeltaSource(deltasrc *Tensor) {
	c.deltasrc = deltasrc.Volume
}

//FindOutputDimsfromInputDims finds the input dims from output dims. Last dest will get + the remainder for overflow or just the remainder of underflow
func (c *ReverseConcat) FindOutputDimsfromInputDims(src []int32, ndests int32, frmt TensorFormat) (destdims [][]int32, err error) {
	fflg := frmt

	switch frmt {
	case fflg.NCHW():
		channels := src[1]
		destchansize := channels / ndests
		remainder := channels % ndests
		var underflow bool
		if destchansize*ndests > channels {
			underflow = true
		}
		destdims := make([][]int32, ndests)
		for i := int32(0); i < ndests; i++ {

			destdims[i] = make([]int32, len(src))
			for j := range destdims[i] {
				destdims[i][j] = src[j]
			}

			if i == int32(len(destdims)-1) {
				if underflow {
					destdims[i][1] = remainder
				}
				destdims[i][1] = destchansize + remainder
			} else {
				destdims[i][1] = destchansize
			}
		}
		return destdims, nil
	case fflg.NHWC():
		channels := src[len(src)-1]
		destchansize := channels / ndests
		remainder := channels % ndests
		var underflow bool
		if destchansize*ndests > channels {
			underflow = true
		}
		for i := int32(0); i < ndests; i++ {

			destdims[i] = make([]int32, len(src))
			for j := range destdims[i] {
				destdims[i][j] = src[j]
			}

			if i == int32(len(destdims)-1) {
				if underflow {
					destdims[i][len(src)-1] = remainder
				}
				destdims[i][len(src)-1] = destchansize + remainder
			} else {
				destdims[i][len(src)-1] = destchansize
			}
		}
		return destdims, nil
	default:
		return nil, errors.New("(c *ReverseConcat) FindOutputDimsfromInputDims: unsupported format")
	}

}

//FindOutputDims finds the output dims for the dests
func (c *ReverseConcat) FindOutputDims(Source *Tensor, ndests int32) (outputdims [][]int32, err error) {

	var tf TensorFormat
	tf.TensorFormat = Source.Format()
	return c.FindOutputDimsfromInputDims(Source.Dims(), ndests, tf)

}

//Forward Does forward with data flowing srcs to dest
func (c *ReverseConcat) Forward() error {

	return c.c.Backward(c.h.Handler, c.dests, c.src)
}

//Backward Does backward with data flowing dest to srcs
func (c *ReverseConcat) Backward() error {
	return c.c.Forward(c.h.Handler, c.deltadests, c.deltasrc)
}

//Concat does the concat operation
type Concat struct {
	c         *tensor.Concat
	h         *Handle
	srcs      []*tensor.Volume
	deltasrcs []*tensor.Volume
	dest      *tensor.Volume
	deltadest *tensor.Volume
}

//CreateConcat creates a concat operation handler
func CreateConcat(h *Handle) (c *Concat, err error) {
	c = new(Concat)
	c.c, err = tensor.CreateConcat(h.Handler)
	c.h = h
	return c, err

}

//SetInputDeltaSrcs sets the delta srcs for back propagation
func (c *Concat) SetInputDeltaSrcs(deltasrcs []*Tensor) {
	c.deltasrcs = make([]*tensor.Volume, len(deltasrcs))
	for i := range deltasrcs {
		c.deltasrcs[i] = deltasrcs[i].Volume
	}
	return
}

//SetInputSrcs sets the input srcs
func (c *Concat) SetInputSrcs(srcs []*Tensor) {
	c.srcs = make([]*tensor.Volume, len(srcs))
	for i := range srcs {
		c.srcs[i] = srcs[i].Volume
	}
	return
}

//SetDest sets the output dest
func (c *Concat) SetDest(dest *Tensor) {
	c.dest = dest.Volume
}

//SetDeltaDest sets the delta dest for back propagation
func (c *Concat) SetDeltaDest(deltadest *Tensor) {
	c.deltadest = deltadest.Volume
}

//FindOutputDimsfromInputDims finds the input dims from output dims
func (c *Concat) FindOutputDimsfromInputDims(srcs [][]int32, frmt TensorFormat) (outputdims []int32, err error) {

	return c.c.GetOutputDimsfromInputDims(srcs, frmt.TensorFormat)
}

//FindOutputDims finds the output dims
func (c *Concat) FindOutputDims(srcs []*Tensor) (outputdims []int32, err error) {
	vols := make([]*tensor.Volume, len(srcs))
	for i := range vols {
		vols[i] = srcs[i].Volume
	}
	outputdims, err = c.c.GetOutputdims(vols)
	return outputdims, err
}

//Forward Does forward with data flowing srcs to dest
func (c *Concat) Forward() error {

	return c.c.Forward(c.h.Handler, c.srcs, c.dest)
}

//Backward Does backward with data flowing dest to srcs
func (c *Concat) Backward() error {
	return c.c.Backward(c.h.Handler, c.deltasrcs, c.deltadest)
}
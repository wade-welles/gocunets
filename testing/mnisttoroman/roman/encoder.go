package roman

import (
	"math/rand"

	gocunets "github.com/dereklstinson/GoCuNets"
	"github.com/dereklstinson/GoCuNets/cudnn"
	"github.com/dereklstinson/GoCuNets/layers/activation"
	"github.com/dereklstinson/GoCuNets/layers/cnn"
	"github.com/dereklstinson/GoCuNets/layers/cnntranspose"
	"github.com/dereklstinson/GoCuNets/layers/dropout"
	"github.com/dereklstinson/GoCuNets/trainer"
	"github.com/dereklstinson/GoCuNets/utils"
	gocudnn "github.com/dereklstinson/GoCudnn"
)

//RomanDecoder using regular method of increasing size of convolution...by just increasing the outer padding
func RomanDecoder(handle *cudnn.Handler,
	frmt cudnn.TensorFormat,
	dtype cudnn.DataType,
	CMode gocudnn.ConvolutionMode,
	memmanaged bool,
	batchsize int32,
	metabatchsize int32,
	learningrates float32,
	codingvector int32,
	numofneurons int32,
	l1regularization float32,
	l2regularization float32) *gocunets.Network {

	filter := utils.Dims
	padding := utils.Dims
	stride := utils.Dims
	dilation := utils.Dims
	//var tmdf gocudnn.TrainingModeFlag
	//tmode := tmdf.Adam()
	//var aflg gocudnn.ActivationModeFlag
	var cflg gocudnn.ConvolutionModeFlag
	reversecmode := cflg.Convolution()
	network := gocunets.CreateNetwork()

	//Setting Up Network

	/*
		Convoultion Layer D1       8
	*/
	network.AddLayer( //in(batchsize, 3, 7, 7)
		cnntranspose.ReverseBuild(handle, frmt, dtype, filter(codingvector, numofneurons, 7, 7), reversecmode, padding(0, 0), stride(1, 1), dilation(1, 1), true, memmanaged),
	) //7
	/*
		Activation Layer D2       9
	*/
	network.AddLayer(
		activation.Leaky(handle),
		//activation.AdvancedThreshRandRelu(handle, dtype, []int32{batchsize, numofneurons, 14, 14}, true),
	)
	/*
		Convoultion Layer D3      10
	*/
	//network.AddLayer(
	//	dropout.Preset(handle, 50, uint64(rand.Int()), memmanaged),
	//)
	network.AddLayer( //in(batchsize, numofneurons, 14, 14)
		cnntranspose.ReverseBuild(handle, frmt, dtype, filter(numofneurons, numofneurons, 8, 8), reversecmode, padding(0, 0), stride(1, 1), dilation(1, 1), false, memmanaged),
	) //7-8+(14)+1 =14
	/*
		Activation Layer D4        11
	*/
	network.AddLayer(
		activation.Leaky(handle),
		//activation.AdvancedThreshRandRelu(handle, dtype, []int32{batchsize, numofneurons, 21, 21}, true),
	)

	/*
		Convoultion Layer D5       12
	*/
	network.AddLayer( //in(batchsize, numofneurons, 21, 21),
		cnntranspose.ReverseBuild(handle, frmt, dtype, filter(numofneurons, numofneurons, 8, 8), reversecmode, padding(0, 0), stride(1, 1), dilation(1, 1), false, memmanaged),
	) //14-8 +14 +1 =2layer1layer
	//network.AddLayer(layer
	//	dropout.Preset(handle, 50, uint64(rand.Int()), memmanaglayered),
	//	)
	/*
		Activation Layer D6       13
	*/
	network.AddLayer(
		//activation.AdvancedThreshRandRelu(handle, dtype, []int32{batchsize, numofneurons, 28, 28}, true),
		activation.Leaky(handle),
	)

	/*
		Convoultion Layer D7         14
	*/
	network.AddLayer( //in(batchsize, numofneurons, 28, 28),
		cnntranspose.ReverseBuild(handle, frmt, dtype, filter(numofneurons, 1, 8, 8), reversecmode, padding(0, 0), stride(1, 1), dilation(1, 1), false, memmanaged),
	) //28

	//var err error
	numoftrainers := network.TrainersNeeded()

	trainersbatch := make([]trainer.Trainer, numoftrainers) //If these were returned then you can do some training parameter adjustements on the fly
	trainerbias := make([]trainer.Trainer, numoftrainers)   //If these were returned then you can do some training parameter adjustements on the fly
	for i := 0; i < numoftrainers; i++ {
		a, b, err := trainer.SetupAdamWandB(handle.XHandle(), l1regularization, l2regularization, metabatchsize)
		a.SetRate(learningrates) //This is here to change the rate if you so want to
		b.SetRate(learningrates)

		trainersbatch[i], trainerbias[i] = a, b

		if err != nil {
			panic(err)
		}

	}
	network.LoadTrainers(handle, trainersbatch, trainerbias) //Load the trainers in the order they are needed
	return network
}

//ArabicEncoder encodes the arabic
func ArabicEncoder(handle *cudnn.Handler,
	frmt cudnn.TensorFormat,
	dtype cudnn.DataType,
	CMode gocudnn.ConvolutionMode,
	memmanaged bool,
	batchsize int32,
	metabatchsize int32,
	learningrates float32,
	codingvector int32,
	numofneurons int32,
	l1regularization float32,
	l2regularization float32) *gocunets.Network {

	filter := utils.Dims
	padding := utils.Dims
	stride := utils.Dims
	dilation := utils.Dims
	//var tmdf gocudnn.TrainingModeFlag
	//tmode := tmdf.Adam()
	//var aflg gocudnn.ActivationModeFlag

	network := gocunets.CreateNetwork()
	//Setting Up Network

	/*
		Convoultion Layer E1  0
	*/

	network.AddLayer( //in(batchsize, 1, 28, 28),
		cnn.SetupDynamic(handle, frmt, dtype, filter(numofneurons, 1, 8, 8), CMode, padding(0, 0), stride(1, 1), dilation(1, 1), memmanaged),
	) //28-8+1 = 21
	/*
		Activation Layer E2    1
	*/
	network.AddLayer(
		activation.Leaky(handle),
		//activation.AdvancedThreshRandRelu(handle, dtype, []int32{batchsize, numofneurons, 21, 21}, true),
	)
	network.AddLayer(
		dropout.Preset(handle, 50, uint64(rand.Int()), memmanaged),
	)
	/*
		Convoultion Layer E3    2
	*/
	network.AddLayer( //in(batchsize, numofneurons, 21, 21),
		cnn.SetupDynamic(handle, frmt, dtype, filter(numofneurons, numofneurons, 8, 8), CMode, padding(0, 0), stride(1, 1), dilation(1, 1), memmanaged),
	) //21-8+1 =14
	/*
		Activation Layer E4    3
	*/
	network.AddLayer(
		activation.Leaky(handle),
		//activation.AdvancedThreshRandRelu(handle, dtype, []int32{batchsize, numofneurons, 14, 14}, true),
	)

	/*
		Convoultion Layer E5    4
	*/
	network.AddLayer( // in(batchsize, numofneurons, 14, 14),
		cnn.SetupDynamic(handle, frmt, dtype, filter(numofneurons, numofneurons, 8, 8), CMode, padding(0, 0), stride(1, 1), dilation(1, 1), memmanaged),
	) // 14-8+1=7
	//network.AddLayer(
	//	dropout.Preset(handle, 50, uint64(rand.Int()), memmanaged),
	//)
	/*
		Activation Layer E6    5
	*/
	network.AddLayer(
		activation.Leaky(handle),
	//	activation.AdvancedThreshRandRelu(handle, dtype, []int32{batchsize, numofneurons, 7, 7}, true),
	)
	/*
		Convoultion Layer E7    6
	*/
	network.AddLayer( // in(batchsize, numofneurons, 7, 7),
		cnn.SetupDynamic(handle, frmt, dtype, filter(codingvector, numofneurons, 7, 7), CMode, padding(0, 0), stride(1, 1), dilation(1, 1), memmanaged),
	) // 1

	/*
		Activation Layer MIDDLE    7
	*/
	network.AddLayer(

		activation.Leaky(handle),

	//	activation.AdvancedThreshRandRelu(handle, dtype, []int32{batchsize, numofneurons, 1, 1}, true),
	)

	//var err error
	numoftrainers := network.TrainersNeeded()

	trainersbatch := make([]trainer.Trainer, numoftrainers) //If these were returned then you can do some training parameter adjustements on the fly
	trainerbias := make([]trainer.Trainer, numoftrainers)   //If these were returned then you can do some training parameter adjustements on the fly
	for i := 0; i < numoftrainers; i++ {
		a, b, err := trainer.SetupAdamWandB(handle.XHandle(), l1regularization, l2regularization, metabatchsize)
		a.SetRate(learningrates) //This is here to change the rate if you so want to
		b.SetRate(learningrates)

		trainersbatch[i], trainerbias[i] = a, b

		if err != nil {
			panic(err)
		}

	}
	network.LoadTrainers(handle, trainersbatch, trainerbias) //Load the trainers in the order they are needed
	return network
}

//ArabicDecoder using regular method of increasing size of convolution...by just increasing the outer padding
func ArabicDecoder(handle *cudnn.Handler,
	frmt cudnn.TensorFormat,
	dtype cudnn.DataType,
	CMode gocudnn.ConvolutionMode,
	memmanaged bool,
	batchsize int32,
	metabatchsize int32,
	learningrates float32,
	codingvector int32,
	numofneurons int32,
	l1regularization float32,
	l2regularization float32) *gocunets.Network {

	filter := utils.Dims
	padding := utils.Dims
	stride := utils.Dims
	dilation := utils.Dims
	//var tmdf gocudnn.TrainingModeFlag
	//tmode := tmdf.Adam()
	//var aflg gocudnn.ActivationModeFlag
	var cflg gocudnn.ConvolutionModeFlag
	reversecmode := cflg.Convolution()
	network := gocunets.CreateNetwork()
	//Setting Up Network

	/*
		Convoultion Layer D1
	*/

	network.AddLayer( // in(batchsize, 3, 7, 7),
		cnntranspose.ReverseBuild(handle, frmt, dtype, filter(codingvector, numofneurons, 7, 7), reversecmode, padding(0, 0), stride(1, 1), dilation(1, 1), false, memmanaged),
	) //7
	/*
		Activation Layer D2
	*/
	network.AddLayer(
		activation.Leaky(handle),
		//activation.AdvancedThreshRandRelu(handle, dtype, []int32{batchsize, numofneurons, 14, 14}, true),
	)
	//network.AddLayer(
	//	dropout.Preset(handle, 50, uint64(rand.Int()), memmanaged),
	//)
	/*
		Convoultion Layer D3
	*/
	network.AddLayer( // in(batchsize, numofneurons, 14, 14),
		cnntranspose.ReverseBuild(handle, frmt, dtype, filter(numofneurons, numofneurons, 8, 8), reversecmode, padding(0, 0), stride(1, 1), dilation(1, 1), false, memmanaged),
	) //7-8+(14)+1 =14
	/*
		Activation Layer D4
	*/
	network.AddLayer(
		activation.Leaky(handle),
		//activation.AdvancedThreshRandRelu(handle, dtype, []int32{batchsize, numofneurons, 21, 21}, true),
	)

	/*
		Convoultion Layer D5
	*/
	network.AddLayer( //in(batchsize, numofneurons, 21, 21),
		cnntranspose.ReverseBuild(handle, frmt, dtype, filter(numofneurons, numofneurons, 8, 8), reversecmode, padding(0, 0), stride(1, 1), dilation(1, 1), false, memmanaged),
	) //14-8 +14 +1 =21
	//network.AddLayer(
	//	dropout.Preset(handle, 50, uint64(rand.Int()), memmanaged),
	//	)
	/*
		Activation Layer D6
	*/
	network.AddLayer(
		activation.Leaky(handle),
	//	activation.AdvancedThreshRandRelu(handle, dtype, []int32{batchsize, numofneurons, 28, 28}, true),
	)

	/*
		Convoultion Layer D7
	*/
	network.AddLayer( //in(batchsize, numofneurons, 28, 28),
		cnntranspose.ReverseBuild(handle, frmt, dtype, filter(numofneurons, 1, 8, 8), reversecmode, padding(0, 0), stride(1, 1), dilation(1, 1), false, memmanaged),
	) //28

	//var err error
	numoftrainers := network.TrainersNeeded()

	trainersbatch := make([]trainer.Trainer, numoftrainers) //If these were returned then you can do some training parameter adjustements on the fly
	trainerbias := make([]trainer.Trainer, numoftrainers)   //If these were returned then you can do some training parameter adjustements on the fly
	for i := 0; i < numoftrainers; i++ {
		a, b, err := trainer.SetupAdamWandB(handle.XHandle(), l1regularization, l2regularization, metabatchsize)
		a.SetRate(learningrates) //This is here to change the rate if you so want to
		b.SetRate(learningrates)

		trainersbatch[i], trainerbias[i] = a, b

		if err != nil {
			panic(err)
		}

	}
	network.LoadTrainers(handle, trainersbatch, trainerbias) //Load the trainers in the order they are needed
	return network
}

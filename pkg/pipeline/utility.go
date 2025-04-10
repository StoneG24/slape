package pipeline

import (
	"log/slog"

	"github.com/StoneG24/slape/cmd/prompt"
	"github.com/StoneG24/slape/internal/vars"
	"github.com/jaypipes/ghw"
	"github.com/jaypipes/ghw/pkg/gpu"
)

func processPrompt(mode string) (string, int64) {
	var promptChoice string
	var maxtokens int64

	switch mode {
	case "simple":
		promptChoice = prompt.SimplePrompt
		maxtokens = 1024
	case "cot":
		promptChoice = prompt.CoTPrompt
		maxtokens = 4096
	case "tot":
		promptChoice = prompt.ToTPrompt
		maxtokens = 32768
	case "got":
		promptChoice = prompt.GoTPrompt
		maxtokens = 32768
	case "moe":
		promptChoice = prompt.MoEPrompt
		maxtokens = 32768
	case "thinkinghats":
		promptChoice = prompt.SixThinkingHats
		maxtokens = 32768
	case "goe":
		promptChoice = prompt.GoEPrompt
		maxtokens = 32768
	default:
		promptChoice = prompt.SimplePrompt
		maxtokens = 100
	}

	return promptChoice, maxtokens
}

func PickImage() string {
	gpuTrue := IsGPU()
	if gpuTrue {
		gpus, err := GatherGPUs()
		if err != nil {
			return vars.CpuImage
		}
		// After reading upstream, he reads the devices mounted
		// with $ ll /sys/class/drm/
		for i, gpu := range gpus {
			// TODO(v) this behavior is mostly for laptops and needs to get looked at again later.
			// onboard graphics card usually is index 0.
			if i == 0 {
				continue
			}
			switch gpu.DeviceInfo.Vendor.Name {
			case "NVIDIA Corporation":
				return vars.CudagpuImage
			case "Advanced Micro Devices, Inc. [AMD/ATI]":
				return vars.RocmgpuImage
			}
		}
	}
	return vars.CpuImage
}

func IsGPU() bool {
	gpuInfo, err := ghw.GPU()
	// if there is an error continue without using a GPU
	if err != nil {
		slog.Error("%s", err)
		slog.Warn("Continuing without GPU...")
	}

	// for debugging
	//slog.Debug("Debug", "Debug", gpuInfo.GraphicsCards)

	if len(gpuInfo.GraphicsCards) == 0 {
		slog.Warn("No GPUs to use, switching to cpu only")
		return false
	} else {
		return true
	}

	// This guy nil derefernce panics when the gpu isn't actually a graphics card
	// fmt.Println(gpuInfo.GraphicsCards[0].Node.Memory)
}

// CheckMemoryUsage is used to check the availble memory of a machine.
func GetAmountofMemory() (int64, error) {
	memory, err := ghw.Memory()
	if err != nil {
		return 0, err
	}
	return memory.TotalUsableBytes, nil
}

// If thread count is low we probably should run only a small amount of models
func GetNumThreads() (uint32, error) {
	cpu, err := ghw.CPU()
	if err != nil {
		return 0, err
	}
	return cpu.TotalThreads, nil
}

// Gather the gpus on the system a select their vendors.
// Normally one would only have two but this interface assumes
// that the internal graphics module on the CPU is a graphics card.
func GatherGPUs() ([]*gpu.GraphicsCard, error) {
	gpuInfo, err := ghw.GPU()
	if err != nil {
		return nil, err
	}

	return gpuInfo.GraphicsCards, nil
}

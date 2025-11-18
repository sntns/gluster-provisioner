package application

import (
	"context"
	"errors"

	"github.com/sntns/gluster-provisioner/pkg/application/layer"
	"github.com/sntns/gluster-provisioner/pkg/model"
)

var (
	ErrInvalidTransition = errors.New("invalid layer transition")
)

// Layer defines the interface for FSM layers.
type Layer interface {
	Up(ctx context.Context, state *layer.State) error   // Move from previous to this layer
	Down(ctx context.Context, state *layer.State) error // Move from this layer to previous
	String() string                                     // Returns the name of the layer
	Dependencies() []string                             // Returns the names of dependent layers
}

// DeviceStack handles the state and transitions for a device.
type DeviceStack struct {
	Current int
	Layers  []Layer
	State   layer.State
}

// NewDeviceStack creates a new Stack for a detected device.
func NewDeviceStack(application *Application, device model.DeviceInfo) *DeviceStack {
	layers := []Layer{
		layer.NewDiscovered(device),
		layer.NewReady(application.fetcher),
		layer.NewPartitioned(application.logger, application.diskManager),
		layer.NewFormatted(application.logger, application.diskManager),
		layer.NewMounted(application.logger, application.diskManager),
		layer.NewGlusterd(application.logger, application.glusterManager),
	}

	state := layer.State{
		Discovered: &layer.DiscoveredState{
			Device: device,
		},
	}

	return &DeviceStack{
		Current: 0,
		Layers:  layers,
		State:   state,
	}
}

// Transition executes the action for the next state and updates the FSM.
//
// NOTE: This method acquires a lock for the entire transition. All transition methods (onReady, onPartitioned, etc.)
// are called within this lock. Therefore, transitions MUST occur asynchronously and must not block for long periods.
// Any long-running operations should be performed outside the lock or in a separate goroutine.
func (stack *DeviceStack) MoveUp(ctx context.Context) error {
	if stack.Current == len(stack.Layers)-1 {
		return ErrInvalidTransition
	}
	layer := stack.Layers[stack.Current+1]
	err := layer.Up(ctx, &stack.State)
	if err != nil {
		return err
	}
	stack.Current++
	return nil
}

func (stack *DeviceStack) MoveDown(ctx context.Context) error {
	if stack.Current == 0 {
		return ErrInvalidTransition
	}
	layer := stack.Layers[stack.Current-1]
	err := layer.Down(ctx, &stack.State)
	if err != nil {
		return err
	}
	stack.Current--
	return nil
}

func (stack *DeviceStack) CurrentName() string {
	return stack.Layers[stack.Current].String()
}

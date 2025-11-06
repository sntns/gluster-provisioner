package http

import (
	"context"
	"encoding/json"
	"log"
	"net/http"
	"sync"

	"github.com/sntns/gluster-provisioner/pkg/capability"
	"github.com/sntns/gluster-provisioner/pkg/device"
	"github.com/sntns/gluster-provisioner/pkg/model"
)

type Controller struct {
	sync.Mutex
	logger   capability.Logger
	config   Configuration
	listener model.DeviceListener
	devices  []model.DeviceInfo
}

type Configuration struct {
	Filters device.Filters
}

func NewController(logger capability.Logger, listener model.DeviceListener, config Configuration) *Controller {
	return &Controller{
		logger:   logger,
		listener: listener,
		config:   config,
		devices:  make([]model.DeviceInfo, 0),
	}
}

func (c *Controller) Start(ctx context.Context) error {
	c.logger.Info("Starting HTTP Device Controller", map[string]any{
		"port": 8081,
	})
	_ = c.listener.DeviceChan() // ensure listener is started
	go http.ListenAndServe(":8081", c)
	return nil
}

func (c *Controller) Stop(ctx context.Context) error {
	c.logger.Info("Stopping HTTP Device Controller", map[string]any{})
	return nil
}

func (c *Controller) AddDevice(device model.DeviceInfo) {
	c.Lock()
	defer c.Unlock()
	c.devices = append(c.devices, device)
	if c.match(device) {
		log.Printf("Device matched: %+v", device)
		ch := c.listener.DeviceChan()
		ch <- device
	}
}

// match checks if any filter matches one of the device's properties
func (c *Controller) match(info model.DeviceInfo) bool {
	return c.config.Filters.Match(info)
}

func (c *Controller) ListDevices() []model.DeviceInfo {
	c.Lock()
	defer c.Unlock()
	return append([]model.DeviceInfo{}, c.devices...)
}

// REST Handlers
func (c *Controller) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	log.Printf("Received %s request for %s", r.Method, r.URL.Path)
	switch r.Method {
	case http.MethodGet:
		c.handleList(w, r)
	case http.MethodPost:
		c.handleAdd(w, r)
	default:
		w.WriteHeader(http.StatusMethodNotAllowed)
	}
}

func (c *Controller) handleList(w http.ResponseWriter, r *http.Request) {
	devices := c.ListDevices()
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(devices)
}

func (c *Controller) handleAdd(w http.ResponseWriter, r *http.Request) {
	var device model.DeviceInfo
	if err := json.NewDecoder(r.Body).Decode(&device); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	c.AddDevice(device)
	w.WriteHeader(http.StatusCreated)
}

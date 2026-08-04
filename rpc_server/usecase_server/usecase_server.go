package usecase_server

import (
	"github.com/enbility/eebus-go/api"
	"github.com/enbility/eebus-grpc/rpc_server"
	"github.com/enbility/eebus-grpc/rpc_services/control_service"
	log "github.com/enbility/eebus-grpc/utils/logging"
	"github.com/enbility/eebus-grpc/utils/type_conversions"
	spineapi "github.com/enbility/spine-go/api"
)

// useCaseEventBufferSize is the channel buffer for use case events.
// A slow or missing subscriber will have events dropped rather than
// blocking the eebus-go internal goroutine.
const useCaseEventBufferSize = 64

type UseCaseServer interface {
	OnUseCaseEvent(ski string, device spineapi.DeviceRemoteInterface, entity spineapi.EntityRemoteInterface, event api.EventType)
	GetUseCase() api.UseCaseInterface
	GetEvents() chan *control_service.SubscribeUseCaseEventsResponse
	rpc_server.RPCServer
}

type BaseUseCaseServer struct {
	events chan *control_service.SubscribeUseCaseEventsResponse

	log.Logger
}

func NewBaseUseCaseServer() *BaseUseCaseServer {
	return &BaseUseCaseServer{
		events: make(chan *control_service.SubscribeUseCaseEventsResponse, useCaseEventBufferSize),
	}
}

func (h *BaseUseCaseServer) OnUseCaseEvent(ski string, device spineapi.DeviceRemoteInterface, entity spineapi.EntityRemoteInterface, event api.EventType) {
	// entity can be nil when the remote device disconnects entirely
	// (e.g. UseCaseSupportUpdate fired by removeDeviceFromAvailableEntityScenarios).
	// Skip events with no entity address rather than panic.
	if entity == nil {
		return
	}
	if h.events == nil {
		return
	}
	remoteEntityAddress := type_conversions.ConvertEEBUSEntityAddress(entity.Address().Entity)
	useCaseEvent := type_conversions.ConvertEEBUSEventType(event)
	useCaseEventResponse := control_service.SubscribeUseCaseEventsResponse{
		RemoteSki:           ski,
		RemoteEntityAddress: remoteEntityAddress,
		UseCaseEvent:        &useCaseEvent,
	}
	select {
	case h.events <- &useCaseEventResponse:
	default:
		log.Infof("UseCaseEvent channel full; dropping event for SKI %s", ski)
	}
}

func (h *BaseUseCaseServer) GetEvents() chan *control_service.SubscribeUseCaseEventsResponse {
	return h.events
}

// CloseEvents closes the events channel, unblocking any SubscribeUseCaseEvents
// stream that is ranging over it. Must be called from the concrete server's Stop().
func (h *BaseUseCaseServer) CloseEvents() {
	if h.events != nil {
		close(h.events)
		h.events = nil
	}
}

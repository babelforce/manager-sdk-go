package manager

import (
	"context"

	managerapi "github.com/babelforce/manager-sdk-go/gen/manager"
)

// EventsResource is the events namespace (/api/v2/events): event definitions and custom events.
type EventsResource struct {
	gc *managerapi.ClientWithResponses
}

// List returns the available event definitions.
func (r *EventsResource) List(ctx context.Context) ([]managerapi.Event, error) {
	resp, err := r.gc.ListEventsWithResponse(ctx, nil)
	if err != nil {
		return nil, err
	}
	data, err := result(resp.JSON200, resp.HTTPResponse, resp.Body)
	if err != nil {
		return nil, err
	}
	return data.Items, nil
}

// CreateCustom creates a custom event.
func (r *EventsResource) CreateCustom(ctx context.Context, body managerapi.CustomEventRequest) (*managerapi.EventItemResponse, error) {
	resp, err := r.gc.CreateCustomEventWithResponse(ctx, body)
	if err != nil {
		return nil, err
	}
	return result(resp.JSON200, resp.HTTPResponse, resp.Body)
}

// DeleteCustom deletes a custom event.
func (r *EventsResource) DeleteCustom(ctx context.Context, id string) error {
	resp, err := r.gc.DeleteCustomEventWithResponse(ctx, id)
	if err != nil {
		return err
	}
	return resultVoid(resp.HTTPResponse, resp.Body)
}

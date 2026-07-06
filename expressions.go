package manager

import (
	"context"

	managerapi "github.com/babelforce/manager-sdk-go/gen/manager"
)

// ExpressionsResource is the expressions namespace (/api/v2/expressions): catalog and evaluation.
type ExpressionsResource struct {
	gc *managerapi.ClientWithResponses
}

// List returns the available expressions.
func (r *ExpressionsResource) List(ctx context.Context) ([]managerapi.AvailableExpression, error) {
	resp, err := r.gc.ListExpressionsWithResponse(ctx)
	if err != nil {
		return nil, err
	}
	data, err := result(resp.JSON200, resp.HTTPResponse, resp.Body)
	if err != nil {
		return nil, err
	}
	return data.Items, nil
}

// Evaluate evaluates an expression against a sample context. async dispatches automations
// asynchronously.
func (r *ExpressionsResource) Evaluate(ctx context.Context, body managerapi.EvaluateExpression, async bool) (*managerapi.EvaluateExpressionResponse, error) {
	resp, err := r.gc.EvaluateExpressionWithResponse(ctx, &managerapi.EvaluateExpressionParams{Async: async}, body)
	if err != nil {
		return nil, err
	}
	return result(resp.JSON200, resp.HTTPResponse, resp.Body)
}

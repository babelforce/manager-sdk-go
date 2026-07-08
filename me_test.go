package manager

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestMeNamespace(t *testing.T) {
	var resetMethod string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		switch r.URL.Path {
		case "/api/v2/user/me":
			_, _ = w.Write([]byte(`{"item":{"id":"` + uuidA + `","username":"me@example.com","roles":["manager"]},"success":true}`))
		case "/api/v2/user/account":
			_, _ = w.Write([]byte(`{"item":{"user":{},"customer":{"company":"Acme"}},"success":true}`))
		case "/api/v2/user/accounts":
			_, _ = w.Write([]byte(`{"items":[{"id":"` + uuidA + `","name":"Acme"},{"id":"` + uuidB + `","name":"Beta"}],"pagination":{"pages":1,"current":1,"total":2,"max":50}}`))
		case "/api/v2/user/reset-password":
			resetMethod = r.Method
			_, _ = w.Write([]byte(`{"message":"sent","success":true}`))
		default:
			w.WriteHeader(http.StatusNotFound)
		}
	}))
	defer srv.Close()

	ctx := context.Background()
	mgr, _ := Connect(ctx, Options{BaseURL: srv.URL, Auth: Bearer("TEST")})

	me, err := mgr.Me.Get(ctx)
	if err != nil {
		t.Fatal(err)
	}
	if me.Item.Username != "me@example.com" || !me.Success {
		t.Fatalf("me.Get = %+v", me.Item)
	}

	cust, err := mgr.Me.Customer(ctx)
	if err != nil {
		t.Fatal(err)
	}
	if cust.Item.Customer.Company != "Acme" {
		t.Fatalf("me.Customer = %+v", cust.Item)
	}

	accounts, err := mgr.Me.Accounts(ctx)
	if err != nil {
		t.Fatal(err)
	}
	if len(accounts) != 2 || accounts[0].Name != "Acme" || accounts[1].Name != "Beta" {
		t.Fatalf("me.Accounts = %+v", accounts)
	}

	if err := mgr.Me.ResetPassword(ctx); err != nil {
		t.Fatal(err)
	}
	if resetMethod != http.MethodPost {
		t.Fatalf("reset method = %q, want POST", resetMethod)
	}
}

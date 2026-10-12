// Command example lists users against a babelforce API host.
//
//	Client creds:   MANAGER_CLIENT_ID=… MANAGER_CLIENT_SECRET=… go run ./example
//	Password grant: MANAGER_USER=… MANAGER_PASS=… go run ./example
//	Other host:     MANAGER_BASE_URL=https://acme.babelforce.com … go run ./example
package main

import (
	"context"
	"fmt"
	"log"
	"os"

	manager "github.com/babelforce/manager-sdk-go"
	managerapi "github.com/babelforce/manager-sdk-go/gen/manager"
)

func main() {
	ctx := context.Background()

	var auth manager.Auth
	if id := os.Getenv("MANAGER_CLIENT_ID"); id != "" {
		auth = manager.ClientCredentials(id, os.Getenv("MANAGER_CLIENT_SECRET"))
	} else {
		auth = manager.Password(os.Getenv("MANAGER_USER"), os.Getenv("MANAGER_PASS"))
	}

	opts := manager.Options{Auth: auth}
	if base := os.Getenv("MANAGER_BASE_URL"); base != "" {
		opts.BaseURL = base // e.g. a non-production or per-customer host
	}

	mgr, err := manager.Connect(ctx, opts)
	if err != nil {
		log.Fatal(err)
	}

	shown := 0
	for user, err := range mgr.Users.List(ctx, manager.ListUsersQuery{}) {
		if err != nil {
			log.Fatal(err)
		}
		fmt.Printf("%s  %s\n", user.Id, user.Email)
		if shown++; shown >= 20 {
			break
		}
	}
	fmt.Printf("\n(showed %d users)\n", shown)

	fmt.Println("\nagents:")
	shown = 0
	for agent, err := range mgr.Agents.List(ctx, manager.ListAgentsQuery{}) {
		if err != nil {
			log.Fatal(err)
		}
		fmt.Printf("%s  %s\n", agent.Id, agent.Name)
		if shown++; shown >= 20 {
			break
		}
	}
	fmt.Printf("(showed %d agents)\n", shown)

	fmt.Println("\ncalls (simple report):")
	shown = 0
	for call, err := range mgr.Calls.Reporting.Simple(ctx, managerapi.ListAllSimpleReportingCallsParams{}) {
		if err != nil {
			log.Fatal(err)
		}
		fmt.Printf("%s\n", call.Id)
		if shown++; shown >= 20 {
			break
		}
	}
	fmt.Printf("(showed %d calls)\n", shown)

	if ids, err := mgr.Metrics.ListIds(ctx); err != nil {
		log.Fatal(err)
	} else {
		fmt.Printf("\nmetric ids: %+v\n", ids)
	}

	fmt.Println("\napplications:")
	shown = 0
	for app, err := range mgr.Applications.List(ctx, manager.ListApplicationsQuery{}) {
		if err != nil {
			log.Fatal(err)
		}
		v, err := manager.ApplicationViewOf(app)
		if err != nil {
			log.Fatal(err)
		}
		fmt.Printf("%s  %s  (%s)\n", v.Id, v.Name, v.Module)
		if shown++; shown >= 20 {
			break
		}
	}
	fmt.Printf("(showed %d applications)\n", shown)

	if rec, err := mgr.Settings.Telephony.AgentRecording.Get(ctx); err != nil {
		log.Fatal(err)
	} else {
		fmt.Printf("\ntelephony.agent.recording: %+v\n", rec)
	}
}

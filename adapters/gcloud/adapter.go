package gcloud

import (
	"context"
	"flag"
	"fmt"
	"log"
	"os"

	"github.com/machado-br/helm-api/adapters/models"
	"golang.org/x/oauth2/google"

	container "google.golang.org/api/container/v1"
)

type adapter struct {
	projectID string
	zone      string
}

type Adapter interface {
	DescribeCluster() (models.Cluster, error)
}

var (
	projectID = flag.String("project", "test-2022-09", "Project ID")
	zone      = flag.String("zone", "us-east1", "Compute zone")
)

func NewAdapter(
	projectID string,
	zone string,
) (adapter, error) {

	return adapter{}, nil
}

func (a adapter) DescribeCluster() {
	flag.Parse()

	if *projectID == "" {
		fmt.Fprintln(os.Stderr, "missing -project flag")
		flag.Usage()
		os.Exit(2)
	}
	if *zone == "" {
		fmt.Fprintln(os.Stderr, "missing -zone flag")
		flag.Usage()
		os.Exit(2)
	}

	ctx := context.Background()

	// See https://cloud.google.com/docs/authentication/.
	// Use GOOGLE_APPLICATION_CREDENTIALS environment variable to specify
	// a service account key file to authenticate to the API.
	hc, err := google.DefaultClient(ctx, container.CloudPlatformScope)
	if err != nil {
		log.Fatalf("Could not get authenticated client: %v", err)
	}

	svc, err := container.New(hc)
	if err != nil {
		log.Fatalf("Could not initialize gke client: %v", err)
	}

	if err := listClusters(svc, *projectID, *zone); err != nil {
		log.Fatal(err)
	}
}

func listClusters(svc *container.Service, projectID, zone string) error {
	list, err := svc.Projects.Zones.Clusters.List(projectID, zone).Do()
	if err != nil {
		return fmt.Errorf("failed to list clusters: %v", err)
	}
	for _, v := range list.Clusters {
		fmt.Printf("Cluster %q (%s) master_version: v%s", v.Name, v.Status, v.CurrentMasterVersion)

		poolList, err := svc.Projects.Zones.Clusters.NodePools.List(projectID, zone, v.Name).Do()
		if err != nil {
			return fmt.Errorf("failed to list node pools for cluster %q: %v", v.Name, err)
		}
		for _, np := range poolList.NodePools {
			fmt.Printf("  -> Pool %q (%s) machineType=%s node_version=v%s autoscaling=%v", np.Name, np.Status,
				np.Config.MachineType, np.Version, np.Autoscaling != nil && np.Autoscaling.Enabled)
		}
	}
	return nil
}

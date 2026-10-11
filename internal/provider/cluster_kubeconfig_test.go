package provider

import "testing"

func TestKubeConfigOfPendingCluster(t *testing.T) {
	resourceModel, err := (&ClusterResource{}).getKubeConfigResourceModel("")
	if err != nil || resourceModel != nil {
		t.Fatalf("resource: got %v, %v; want no kubeconfig and no error", resourceModel, err)
	}

	dataSourceModel, err := (&ClusterDataSource{}).getKubeConfigResourceModel("")
	if err != nil || dataSourceModel != nil {
		t.Fatalf("data source: got %v, %v; want no kubeconfig and no error", dataSourceModel, err)
	}
}

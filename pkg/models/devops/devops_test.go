package devops

import (
	"kubesphere.io/kubesphere/pkg/simple/client/devops"
	"kubesphere.io/kubesphere/pkg/simple/client/devops/fake"
	"net/http"
	"testing"
)

const baseUrl = "http://127.0.0.1/kapis/devops.kubesphere.io/v1alpha2/"

func TestGetNodesDetail(t *testing.T) {
	fakeData := make(map[string]interface{})
	PipelineRunNodes := []devops.PipelineRunNodes{
		{
			DisplayName: "Deploy to Kubernetes",
			ID:          "1",
			Result:      "SUCCESS",
		},
		{
			DisplayName: "Deploy to Kubernetes",
			ID:          "2",
			Result:      "SUCCESS",
		},
		{
			DisplayName: "Deploy to Kubernetes",
			ID:          "3",
			Result:      "SUCCESS",
		},
	}

	NodeSteps := []devops.NodeSteps{
		{
			DisplayName: "Deploy to Kubernetes",
			ID:          "1",
			Result:      "SUCCESS",
		},
	}

	fakeData["project1-pipeline1-run1"] = PipelineRunNodes
	fakeData["project1-pipeline1-run1-1"] = NodeSteps
	fakeData["project1-pipeline1-run1-2"] = NodeSteps
	fakeData["project1-pipeline1-run1-3"] = NodeSteps

	devopsClient := fake.NewFakeDevops(fakeData)

	devopsOperator := NewDevopsOperator(devopsClient)

	httpReq, _ := http.NewRequest(http.MethodGet, baseUrl+"devops/project1/pipelines/pipeline1/runs/run1/nodesdetail/?limit=10000", nil)

	nodesDetails, err := devopsOperator.GetNodesDetail("project1", "pipeline1", "run1", httpReq)
	if err != nil || nodesDetails == nil {
		t.Fatalf("should not get error %+v", err)
	}

	for _, v := range nodesDetails {
		if v.Steps[0].ID != "1" {
			t.Fatalf("Can not get any step.")
		}
	}
}

func TestGetBranchNodesDetail(t *testing.T) {
	fakeData := make(map[string]interface{})

	BranchPipelineRunNodes := []devops.BranchPipelineRunNodes{
		{
			DisplayName: "Deploy to Kubernetes",
			ID:          "1",
			Result:      "SUCCESS",
		},
		{
			DisplayName: "Deploy to Kubernetes",
			ID:          "2",
			Result:      "SUCCESS",
		},
		{
			DisplayName: "Deploy to Kubernetes",
			ID:          "3",
			Result:      "SUCCESS",
		},
	}

	BranchNodeSteps := []devops.NodeSteps{
		{
			DisplayName: "Deploy to Kubernetes",
			ID:          "1",
			Result:      "SUCCESS",
		},
	}

	fakeData["project1-pipeline1-branch1-run1"] = BranchPipelineRunNodes
	fakeData["project1-pipeline1-branch1-run1-1"] = BranchNodeSteps
	fakeData["project1-pipeline1-branch1-run1-2"] = BranchNodeSteps
	fakeData["project1-pipeline1-branch1-run1-3"] = BranchNodeSteps

	devopsClient := fake.NewFakeDevops(fakeData)

	devopsOperator := NewDevopsOperator(devopsClient)

	httpReq, _ := http.NewRequest(http.MethodGet, baseUrl+"devops/project1/pipelines/pipeline1/branchs/branch1/runs/run1/nodesdetail/?limit=10000", nil)

	nodesDetails, err := devopsOperator.GetBranchNodesDetail("project1", "pipeline1", "branch1", "run1", httpReq)
	if err != nil || nodesDetails == nil {
		t.Fatalf("should not get error %+v", err)
	}

	for _, v := range nodesDetails {
		if v.Steps[0].ID != "1" {
			t.Fatalf("Can not get any step.")
		}
	}
}

package main

import (
	"context"
	"encoding/json"
	"fmt"
	"sync"
	"time"

	"github.com/flyteorg/flytepropeller/pkg/apis/flyteworkflow"
	"github.com/flyteorg/flytepropeller/pkg/apis/flyteworkflow/v1alpha1"
	clientset "github.com/flyteorg/flytepropeller/pkg/client/clientset/versioned"
	"github.com/flyteorg/flytepropeller/pkg/controller/config"
	"github.com/flyteorg/flytepropeller/pkg/utils"

	stdconfig "github.com/flyteorg/flytestdlib/config"

	apiextensionsclientset "k8s.io/apiextensions-apiserver/pkg/client/clientset/clientset"

	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	types "k8s.io/apimachinery/pkg/types"
)

const (
	namespace = "flytesnacks-development"
	threadCount = 150 // one workflow per thread
	nodeCount = 100
)

type NodeStatusPatch struct {
	Op    string               `json:"op"`
	Path  string               `json:"path"`
	Value *v1alpha1.NodeStatus `json:"value"`
}

func main() {
	ctx := context.Background()

	// initialize kube client 
	cfg := &config.Config{
		KubeConfigPath: "$HOME/.kube/config",
		KubeConfig: config.KubeClientConfig{
			QPS:     100,
			Burst:   25,
			Timeout: stdconfig.Duration{Duration: 30 * time.Second},
		},
	}

	_, kubecfg, err := utils.GetKubeConfig(ctx, cfg)
	if err != nil {
		fmt.Printf("failed to get kube config with err '%v'\n", err)
		return
	}

	flyteworkflowClient, err := clientset.NewForConfig(kubecfg)
	if err != nil {
		fmt.Printf("failed to initialize flyteworkflow clientset with err '%v'\n", err)
		return
	}

	// create flyteworkflow CRD if it does not exist
	apiextensionsClient, err := apiextensionsclientset.NewForConfig(kubecfg)
	if err != nil {
		fmt.Printf("failed to intiialize apiextensions clientset with err '%v'\n", err)
		return
	}

	_, err = apiextensionsClient.ApiextensionsV1().CustomResourceDefinitions().Create(ctx, &flyteworkflow.CRD, metav1.CreateOptions{})
	if err != nil {
		if !apierrors.IsAlreadyExists(err) {
			fmt.Printf("failed to create FlyteWorkflow CRD with err '%v'\n", err)
			return
		}
	}

	var wg sync.WaitGroup
	wg.Add(threadCount)
	for i:=0; i<threadCount; i++ {
		go func(i int){
			defer wg.Done()

			// create workflow
			nodeStatuses := make(map[v1alpha1.NodeID]*v1alpha1.NodeStatus)
			for j:=0; j<nodeCount; j++ {
				nodeStatuses[fmt.Sprintf("node-%d", j)] = &v1alpha1.NodeStatus{
					Phase: v1alpha1.NodePhaseNotYetStarted,
				}
			}

			workflow := &v1alpha1.FlyteWorkflow{
				TypeMeta: metav1.TypeMeta{
				},
				ObjectMeta: metav1.ObjectMeta{
					Name: fmt.Sprintf("benchmark-%d", i),
					Namespace: namespace,
				},
				Status: v1alpha1.WorkflowStatus{
					NodeStatus: nodeStatuses,
				},
			}

			workflow, err = flyteworkflowClient.FlyteworkflowV1alpha1().FlyteWorkflows(namespace).Create(ctx, workflow, metav1.CreateOptions{})
			if err != nil {
				fmt.Printf("failed to create FlyteWorkflow CRD with err '%v'\n", err)
				return
			}

			patchSingle(ctx, workflow, flyteworkflowClient)
			//patchAll(ctx, workflow, flyteworkflowClient)
			//updateSingle(ctx, workflow, flyteworkflowClient)
			//updateAll(ctx, workflow, flyteworkflowClient)
		}(i)
	}

	wg.Wait()
}

func patchSingle(ctx context.Context, workflow *v1alpha1.FlyteWorkflow, flyteworkflowClient *clientset.Clientset) {
	for i:=0; i<nodeCount; i++ {
		nodeId := fmt.Sprintf("node-%d", i)
		status := workflow.Status.NodeStatus[nodeId]
		status.Phase = v1alpha1.NodePhaseRunning
		
		nodeStatusPatch := []NodeStatusPatch{{
			Op:   "replace",
			Path: fmt.Sprintf("/status/nodeStatus/node-%d", i),
			Value: status,
		}}

		patchBytes, err := json.Marshal(nodeStatusPatch)
		if err != nil {
			fmt.Printf("failed to marshal FlyteWorkflow CRD patch err '%v'\n", err)
			continue
		}

		workflow, err = flyteworkflowClient.FlyteworkflowV1alpha1().FlyteWorkflows(namespace).Patch(ctx, workflow.ObjectMeta.Name, types.JSONPatchType, patchBytes, metav1.PatchOptions{})
		if err != nil {
			fmt.Printf("failed to patch FlyteWorkflow CRD with err '%v'\n", err)
			return
		}
	}
}

func patchAll(ctx context.Context, workflow *v1alpha1.FlyteWorkflow, flyteworkflowClient *clientset.Clientset) {
	var err error
	var nodeStatusPatch []NodeStatusPatch
	for i:=0; i<nodeCount; i++ {
		nodeId := fmt.Sprintf("node-%d", i)
		status := workflow.Status.NodeStatus[nodeId]
		status.Phase = v1alpha1.NodePhaseRunning
		
		nodeStatusPatch = append(nodeStatusPatch, NodeStatusPatch{
			Op:   "replace",
			Path: fmt.Sprintf("/status/nodeStatus/node-%d", i),
			Value: status,
		})
	}

	patchBytes, err := json.Marshal(nodeStatusPatch)
	if err != nil {
		fmt.Printf("failed to marshal FlyteWorkflow CRD patch err '%v'\n", err)
		return
	}

	workflow, err = flyteworkflowClient.FlyteworkflowV1alpha1().FlyteWorkflows(namespace).Patch(ctx, workflow.ObjectMeta.Name, types.JSONPatchType, patchBytes, metav1.PatchOptions{})
	if err != nil {
		fmt.Printf("failed to patch FlyteWorkflow CRD with err '%v'\n", err)
		return
	}
}

func updateSingle(ctx context.Context, workflow *v1alpha1.FlyteWorkflow, flyteworkflowClient *clientset.Clientset) {
	var err error
	for i:=0; i<nodeCount; i++ {
		nodeId := fmt.Sprintf("node-%d", i)
		status := workflow.Status.NodeStatus[nodeId]
		status.Phase = v1alpha1.NodePhaseRunning

		workflow, err = flyteworkflowClient.FlyteworkflowV1alpha1().FlyteWorkflows(namespace).Update(ctx, workflow, metav1.UpdateOptions{})
		if err != nil {
			fmt.Printf("failed to update FlyteWorkflow CRD with err '%v'\n", err)
			return
		}
	}
}

func updateAll(ctx context.Context, workflow *v1alpha1.FlyteWorkflow, flyteworkflowClient *clientset.Clientset) {
	var err error
	for i:=0; i<nodeCount; i++ {
		nodeId := fmt.Sprintf("node-%d", i)
		status := workflow.Status.NodeStatus[nodeId]
		status.Phase = v1alpha1.NodePhaseRunning
	}

	workflow, err = flyteworkflowClient.FlyteworkflowV1alpha1().FlyteWorkflows(namespace).Update(ctx, workflow, metav1.UpdateOptions{})
	if err != nil {
		fmt.Printf("failed to update FlyteWorkflow CRD with err '%v'\n", err)
		return
	}
}

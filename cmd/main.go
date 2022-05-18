package main

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/flyteorg/flytepropeller/pkg/apis/flyteworkflow"
	"github.com/flyteorg/flytepropeller/pkg/apis/flyteworkflow/v1alpha1"
	clientset "github.com/flyteorg/flytepropeller/pkg/client/clientset/versioned"
	//v1alpha12 "github.com/flyteorg/flytepropeller/pkg/client/clientset/versioned/typed/flyteworkflow/v1alpha1"
	"github.com/flyteorg/flytepropeller/pkg/controller/config"
	"github.com/flyteorg/flytepropeller/pkg/utils"

	apiextensionsclientset "k8s.io/apiextensions-apiserver/pkg/client/clientset/clientset"

	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	types "k8s.io/apimachinery/pkg/types"
)

const (
	namespace = "flytesnacks-development"
	nodeCount = 10
)

type NodeStatusPatch struct {
	Op    string               `json:"op"`
	Path  string               `json:"path"`
	Value *v1alpha1.NodeStatus `json:"value"`
}

/*type NodeStatus struct {
	MutableStruct
	Phase                NodePhase     `json:"phase,omitempty"`
	QueuedAt             *metav1.Time  `json:"queuedAt,omitempty"`
	StartedAt            *metav1.Time  `json:"startedAt,omitempty"`
	StoppedAt            *metav1.Time  `json:"stoppedAt,omitempty"`
	LastUpdatedAt        *metav1.Time  `json:"lastUpdatedAt,omitempty"`
	LastAttemptStartedAt *metav1.Time  `json:"laStartedAt,omitempty"`
	Message              string        `json:"message,omitempty"`
	DataDir              DataReference `json:"-"`
	OutputDir            DataReference `json:"-"`
	Attempts             uint32        `json:"attempts,omitempty"`
	SystemFailures       uint32        `json:"systemFailures,omitempty"`
	Cached               bool          `json:"cached,omitempty"`

	// This is useful only for branch nodes. If this is set, then it can be used to determine if execution can proceed
	ParentNode    *NodeID                  `json:"parentNode,omitempty"`
	ParentTask    *TaskExecutionIdentifier `json:"-"`
	BranchStatus  *BranchNodeStatus        `json:"branchStatus,omitempty"`
	SubNodeStatus map[NodeID]*NodeStatus   `json:"subNodeStatus,omitempty"`
	// We can store the outputs at this layer

	// TODO not used delete
	WorkflowNodeStatus *WorkflowNodeStatus `json:"workflowNodeStatus,omitempty"`

	TaskNodeStatus    *TaskNodeStatus    `json:",omitempty"`
	DynamicNodeStatus *DynamicNodeStatus `json:"dynamicNodeStatus,omitempty"`
	// In case of Failing/Failed Phase, an execution error can be optionally associated with the Node
	Error *ExecutionError `json:"error,omitempty"`

	// Not Persisted
	DataReferenceConstructor storage.ReferenceConstructor `json:"-"`
}*/

func main() {
	ctx := context.Background()

	// initialize kube client 
	cfg := &config.Config{
		KubeConfigPath: "$HOME/.kube/config",
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

	// create workflow
	nodeStatuses := make(map[v1alpha1.NodeID]*v1alpha1.NodeStatus)
	for i:=0; i<nodeCount; i++ {
		nodeStatuses[fmt.Sprintf("node-%d", i)] = &v1alpha1.NodeStatus{
			Phase: v1alpha1.NodePhaseNotYetStarted,
		}
	}
	//fmt.Printf("%+v\n", workflow)

	workflow := &v1alpha1.FlyteWorkflow{
		TypeMeta: metav1.TypeMeta{
		},
		ObjectMeta: metav1.ObjectMeta{
			Name: "foo",
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

	// patch workflow
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

// Copyright 2019 Kube Capacity Authors
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package capacity

import (
	"context"
	"fmt"
	"os"

	"k8s.io/client-go/kubernetes"
	metrics "k8s.io/metrics/pkg/client/clientset/versioned"

	"github.com/fnickels/kube-capacity/pkg/kube"
	appsv1 "k8s.io/api/apps/v1"
	autov1 "k8s.io/api/autoscaling/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	resourcehelper "k8s.io/kubectl/pkg/util/resource"
	v1beta1 "k8s.io/metrics/pkg/apis/metrics/v1beta1"
)

// FetchAndPrint gathers cluster resource data and outputs it
func FetchAndPrint(cr *DisplayCriteria) {

	clientset, err := kube.NewClientSet(cr.KubeContext, cr.KubeConfig)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error connecting to Kubernetes: %v\n", err)
		os.Exit(1)
	}

	podList, nodeList := getPodsAndNodes(clientset, &cr.Filters)

	deploymentList := getDeployments(clientset)
	hpaList := getHPAs(clientset)

	var pmList *v1beta1.PodMetricsList
	var nmList *v1beta1.NodeMetricsList

	// grab utilization data if either flag is set
	if cr.ShowUtil || cr.BinpackAnalysis {
		mClientset, err := kube.NewMetricsClientSet(cr.KubeContext, cr.KubeConfig)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error connecting to Metrics API: %v\n", err)
			os.Exit(4)
		}

		pmList = getPodMetrics(mClientset, cr.Filters.Namespace)
		if cr.Filters.Namespace == "" && cr.Filters.NamespaceLabels == "" {
			nmList = getNodeMetrics(mClientset, cr.Filters.NodeLabels)
		}
	}

	if cr.ShowDebug {
		debug(clientset, podList, pmList, deploymentList, hpaList)
	}

	cm := buildClusterMetric(podList, pmList, nodeList, nmList, cr)

	cm.analyzeCluster(cr)

	printList(&cm, cr)
}

func getPodsAndNodes(clientset kubernetes.Interface, filters *SelectionFilters) (*corev1.PodList, *corev1.NodeList) {

	nodeList, err := clientset.CoreV1().Nodes().List(context.TODO(), metav1.ListOptions{
		LabelSelector: filters.NodeLabels,
	})
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error listing Nodes: %v\n", err)
		os.Exit(2)
	}

	podList, err := clientset.CoreV1().Pods(filters.Namespace).List(context.TODO(), metav1.ListOptions{
		LabelSelector: filters.PodLabels,
	})
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error listing Pods: %v\n", err)
		os.Exit(3)
	}

	newPodItems := []corev1.Pod{}

	nodes := map[string]bool{}
	for _, node := range nodeList.Items {
		nodes[node.GetName()] = true
	}

	for _, pod := range podList.Items {
		if _, ok := nodes[pod.Spec.NodeName]; !ok {
			continue
		}

		newPodItems = append(newPodItems, pod)
	}

	podList.Items = newPodItems

	if filters.Namespace == "" && filters.NamespaceLabels != "" {
		namespaceList, err := clientset.CoreV1().Namespaces().List(context.TODO(), metav1.ListOptions{
			LabelSelector: filters.NamespaceLabels,
		})
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error listing Namespaces: %v\n", err)
			os.Exit(3)
		}

		namespaces := map[string]bool{}
		for _, ns := range namespaceList.Items {
			namespaces[ns.GetName()] = true
		}

		newPodItems := []corev1.Pod{}

		for _, pod := range podList.Items {
			if _, ok := namespaces[pod.GetNamespace()]; !ok {
				continue
			}

			newPodItems = append(newPodItems, pod)
		}

		podList.Items = newPodItems
	}

	return podList, nodeList
}

func getPodMetrics(mClientset *metrics.Clientset, namespace string) *v1beta1.PodMetricsList {

	pmList, err := mClientset.MetricsV1beta1().PodMetricses(namespace).List(context.TODO(), metav1.ListOptions{})

	if err != nil {
		fmt.Fprintf(os.Stderr, "Error getting Pod Metrics: %v\n", err)
		fmt.Fprintf(os.Stderr, "For this to work, metrics-server needs to be running in your cluster\n")
		os.Exit(6)
	}

	return pmList
}

func getNodeMetrics(mClientset *metrics.Clientset, nodeLabels string) *v1beta1.NodeMetricsList {

	nmList, err := mClientset.MetricsV1beta1().NodeMetricses().List(context.TODO(), metav1.ListOptions{
		LabelSelector: nodeLabels,
	})

	if err != nil {
		fmt.Fprintf(os.Stderr, "Error getting Node Metrics: %v\n", err)
		fmt.Fprintf(os.Stderr, "For this to work, metrics-server needs to be running in your cluster\n")
		os.Exit(7)
	}

	return nmList
}

func debug(clientset kubernetes.Interface,
	podList *corev1.PodList,
	pmList *v1beta1.PodMetricsList,
	deployList *appsv1.DeploymentList,
	hpaList *autov1.HorizontalPodAutoscalerList) {

	fmt.Fprintf(os.Stdout, "-------------------\n")

	if hpaList != nil {
		fmt.Fprintf(os.Stdout, "===================\n")
		for i, hpa := range hpaList.Items {
			fmt.Fprintf(os.Stdout, "HPA  %d: %v\n", i, hpa.GetName())
			fmt.Fprintf(os.Stdout, "      : %v\n", hpa.GetNamespace())
			fmt.Fprintf(os.Stdout, "      : %v\n", hpa.GetUID())
			fmt.Fprintf(os.Stdout, "      : %v\n", hpa.GetResourceVersion())
			fmt.Fprintf(os.Stdout, "      : %v\n", hpa.GetLabels())
			fmt.Fprintf(os.Stdout, "      : %v\n", hpa.GetCreationTimestamp())
			fmt.Fprintf(os.Stdout, "      : %v\n", hpa.GetGenerateName())
			fmt.Fprintf(os.Stdout, "      : %v\n", hpa.GetManagedFields())
		}
		fmt.Fprintf(os.Stdout, "===================\n")
	}

	fmt.Fprintf(os.Stdout, "-------------------\n")

	if deployList != nil {
		fmt.Fprintf(os.Stdout, "===================\n")
		for i, deploy := range deployList.Items {
			fmt.Fprintf(os.Stdout, "Deployment %d: %v\n", i, deploy.GetName())
			fmt.Fprintf(os.Stdout, "            : %v\n", deploy.GetNamespace())
			fmt.Fprintf(os.Stdout, "            : %v\n", deploy.GetUID())
			fmt.Fprintf(os.Stdout, "            : %v\n", deploy.GetResourceVersion())
			fmt.Fprintf(os.Stdout, "            : %v\n", deploy.GetLabels())
			fmt.Fprintf(os.Stdout, "            : %v\n", deploy.GetCreationTimestamp())
			fmt.Fprintf(os.Stdout, "            : %v\n", deploy.GetGenerateName())
			fmt.Fprintf(os.Stdout, "            : %v\n", deploy.GetManagedFields())
		}
		fmt.Fprintf(os.Stdout, "===================\n")
	}

	fmt.Fprintf(os.Stdout, "-------------------\n")
	for i, pod := range podList.Items {
		fmt.Fprintf(os.Stdout, "pod %d: %v\n", i, pod.GetName())
		fmt.Fprintf(os.Stdout, "      : %v\n", pod.GetNamespace())
		fmt.Fprintf(os.Stdout, "      : %v\n", pod.Status.Phase)
		fmt.Fprintf(os.Stdout, "Parent: %v\n", podOwnerIs(clientset, &pod))
		fmt.Fprintf(os.Stdout, "      : %v\n", pod.GetUID())
		fmt.Fprintf(os.Stdout, " owner: %v\n", pod.GetOwnerReferences())
		fmt.Fprintf(os.Stdout, "      : %v\n", pod.Kind)
		fmt.Fprintf(os.Stdout, "      : %v\n", pod.GetGeneration())
		fmt.Fprintf(os.Stdout, "      : %v\n", pod.GetLabels())
		fmt.Fprintf(os.Stdout, "      : %v\n", pod.GetCreationTimestamp())
		fmt.Fprintf(os.Stdout, "      : %v\n", pod.GetAnnotations())

		req, limit := resourcehelper.PodRequestsAndLimits(&pod)

		fmt.Fprintf(os.Stdout, "      : %v\n", req)
		fmt.Fprintf(os.Stdout, "      : %v\n", limit)
		fmt.Fprintf(os.Stdout, " init : %v\n", len(pod.Spec.InitContainers))
		fmt.Fprintf(os.Stdout, " cont : %v\n", len(pod.Spec.Containers))
		fmt.Fprintf(os.Stdout, " ephm : %v\n", len(pod.Spec.EphemeralContainers))

		for i, c := range pod.Spec.InitContainers {
			fmt.Fprintf(os.Stdout, " init %3d -> %v\n", i, c.Name)
			fmt.Fprintf(os.Stdout, "               Request  %v\n", c.Resources.Requests)
			fmt.Fprintf(os.Stdout, "               Limit    %v\n", c.Resources.Limits)
		}
		for i, c := range pod.Spec.Containers {
			fmt.Fprintf(os.Stdout, " cont %3d -> %v\n", i, c.Name)
			fmt.Fprintf(os.Stdout, "               Request  %v\n", c.Resources.Requests)
			fmt.Fprintf(os.Stdout, "               Limit    %v\n", c.Resources.Limits)
		}
		for i, c := range pod.Spec.EphemeralContainers {
			fmt.Fprintf(os.Stdout, " ephm %3d -> %v \n", i, c.Name)
			fmt.Fprintf(os.Stdout, "               Request  %v\n", c.Resources.Requests)
			fmt.Fprintf(os.Stdout, "               Limit    %v\n", c.Resources.Limits)
		}

		fmt.Fprintf(os.Stdout, "      : %v\n", pod)
	}
	fmt.Fprintf(os.Stdout, "-------------------\n")

	if pmList != nil {
		fmt.Fprintf(os.Stdout, "===================\n")
		for i, pod := range pmList.Items {
			fmt.Fprintf(os.Stdout, "pod %d: %v\n", i, pod.GetName())
			fmt.Fprintf(os.Stdout, "      : %v\n", pod.GetNamespace())
			fmt.Fprintf(os.Stdout, "      : %v\n", pod.GetUID())
			fmt.Fprintf(os.Stdout, " owner: %v\n", pod.GetOwnerReferences())
			fmt.Fprintf(os.Stdout, "      : %v\n", pod.GetLabels())
			fmt.Fprintf(os.Stdout, "      : %v\n", pod.GetCreationTimestamp())
			fmt.Fprintf(os.Stdout, " cont : %v\n", len(pod.Containers))

			for i, c := range pod.Containers {
				fmt.Fprintf(os.Stdout, "      %3d -> %v  ==> %v\n", i, c.Name, c.Usage)
			}
		}
		fmt.Fprintf(os.Stdout, "===================\n")
	}

}

func podOwnerIs(clientset kubernetes.Interface, pod *corev1.Pod) string {

	if len(pod.OwnerReferences) == 0 {
		return "<<has no owner>>"
	}

	var ownerName, ownerKind string

	switch pod.OwnerReferences[0].Kind {
	case "ReplicaSet":
		fmt.Fprintf(os.Stdout, "get replica set for %s in %s\n", pod.OwnerReferences[0].Name, pod.Namespace)
		replica, repErr := clientset.AppsV1().ReplicaSets(pod.Namespace).Get(context.TODO(), pod.OwnerReferences[0].Name, metav1.GetOptions{})
		if repErr != nil {
			fmt.Fprintf(os.Stderr, "Error getting Pod Replica Set: %v\n", repErr)
			os.Exit(1)
		}

		ownerName = replica.OwnerReferences[0].Name
		ownerKind = "Deployment"

	case "DaemonSet", "StatefulSet":
		ownerName = pod.OwnerReferences[0].Name
		ownerKind = pod.OwnerReferences[0].Kind

	default:
		return fmt.Sprintf("unknown type [%s]", pod.OwnerReferences[0].Kind)
	}

	return fmt.Sprintf("%s -> [%s]", ownerKind, ownerName)
}

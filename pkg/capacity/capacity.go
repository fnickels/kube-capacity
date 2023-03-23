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

	"github.com/robscott/kube-capacity/pkg/kube"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	resourcehelper "k8s.io/kubectl/pkg/util/resource"
	v1beta1 "k8s.io/metrics/pkg/apis/metrics/v1beta1"
)

// FetchAndPrint gathers cluster resource data and outputs it
func FetchAndPrint(
	showContainers, showPods, showUtil, showPodCount, showAllNodeLabels,
	availableFormat, binpackAnalysis, showPodSummary, showDebug bool,
	podLabels, nodeLabels, displayNodeLabels, groupByNodeLabels,
	namespaceLabels, namespace,
	kubeContext, kubeConfig, output, sortBy string) {

	clientset, err := kube.NewClientSet(kubeContext, kubeConfig)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error connecting to Kubernetes: %v\n", err)
		os.Exit(1)
	}

	podList, nodeList := getPodsAndNodes(clientset, podLabels, nodeLabels, namespaceLabels, namespace)
	var pmList *v1beta1.PodMetricsList
	var nmList *v1beta1.NodeMetricsList

	// grab utilization data if either flag is set
	if showUtil || binpackAnalysis {
		mClientset, err := kube.NewMetricsClientSet(kubeContext, kubeConfig)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error connecting to Metrics API: %v\n", err)
			os.Exit(4)
		}

		pmList = getPodMetrics(mClientset, namespace)
		if namespace == "" && namespaceLabels == "" {
			nmList = getNodeMetrics(mClientset, nodeLabels)
		}
	}

	if showDebug {
		fmt.Fprintf(os.Stdout, "-------------------\n")
		for i, pod := range podList.Items {
			fmt.Fprintf(os.Stdout, "pod %d: %v\n", i, pod.GetName())
			fmt.Fprintf(os.Stdout, "      : %v\n", pod.GetNamespace())
			fmt.Fprintf(os.Stdout, "      : %v\n", pod.Status.Phase)
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

			if i > 5 {
				break
			}
		}
		fmt.Fprintf(os.Stdout, "-------------------\n")

		if pmList != nil {
			fmt.Fprintf(os.Stdout, "===================\n")
			for i, pod := range pmList.Items {
				fmt.Fprintf(os.Stdout, "pod %d: %v\n", i, pod.GetName())
				fmt.Fprintf(os.Stdout, "      : %v\n", pod.GetNamespace())
				fmt.Fprintf(os.Stdout, "      : %v\n", pod.GetLabels())
				fmt.Fprintf(os.Stdout, "      : %v\n", pod.GetCreationTimestamp())
				fmt.Fprintf(os.Stdout, " cont : %v\n", len(pod.Containers))

				for i, c := range pod.Containers {
					fmt.Fprintf(os.Stdout, "      %3d -> %v  ==> %v\n", i, c.Name, c.Usage)
				}

				if i > 5 {
					break
				}
			}
			fmt.Fprintf(os.Stdout, "===================\n")
		}
	}

	cm := buildClusterMetric(podList, pmList, nodeList, nmList)

	showNamespace := namespace == ""

	printList(&cm,
		showContainers, showPods, showUtil, showPodCount, showNamespace, showAllNodeLabels, showDebug,
		displayNodeLabels, groupByNodeLabels,
		output, sortBy, availableFormat, binpackAnalysis, showPodSummary)
}

func getPodsAndNodes(clientset kubernetes.Interface, podLabels, nodeLabels, namespaceLabels, namespace string) (*corev1.PodList, *corev1.NodeList) {
	nodeList, err := clientset.CoreV1().Nodes().List(context.TODO(), metav1.ListOptions{
		LabelSelector: nodeLabels,
	})
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error listing Nodes: %v\n", err)
		os.Exit(2)
	}

	podList, err := clientset.CoreV1().Pods(namespace).List(context.TODO(), metav1.ListOptions{
		LabelSelector: podLabels,
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
		if !nodes[pod.Spec.NodeName] {
			continue
		}

		newPodItems = append(newPodItems, pod)
	}

	podList.Items = newPodItems

	if namespace == "" && namespaceLabels != "" {
		namespaceList, err := clientset.CoreV1().Namespaces().List(context.TODO(), metav1.ListOptions{
			LabelSelector: namespaceLabels,
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
			if !namespaces[pod.GetNamespace()] {
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

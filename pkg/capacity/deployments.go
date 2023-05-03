package capacity

import (
	"context"
	"fmt"
	"os"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"

	appsv1 "k8s.io/api/apps/v1"
	autov1 "k8s.io/api/autoscaling/v1"
	corev1 "k8s.io/api/core/v1"
)

func getDeployments(clientset kubernetes.Interface) *appsv1.DeploymentList {

	deploymentList, err := clientset.AppsV1().Deployments(corev1.NamespaceAll).List(context.TODO(), metav1.ListOptions{})

	if err != nil {
		fmt.Fprintf(os.Stderr, "Error listing Deployments: %v\n", err)
		os.Exit(2)
	}

	return deploymentList
}

func getHPAs(clientset kubernetes.Interface) *autov1.HorizontalPodAutoscalerList {

	hpaList, err := clientset.AutoscalingV1().HorizontalPodAutoscalers(corev1.NamespaceAll).List(context.TODO(), metav1.ListOptions{})

	if err != nil {
		fmt.Fprintf(os.Stderr, "Error listing HPAs: %v\n", err)
		os.Exit(2)
	}

	return hpaList
}

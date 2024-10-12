// This module is responsible for authenticating the application to the kubernetes API
// and then providing some functionality to interact with the cluster

package kubernetes

import (
	"context"
	"log"
	"os"

	"github.com/google/uuid"

	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/clientcmd"

	core "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// Client is a struct that holds the kubernetes client
type Client struct {
	client *kubernetes.Clientset
}

// NewClient creates a new kubernetes client
func NewClient() (*Client, error) {
	// Get the kubeconfig file
	kubeconfigPath := os.Getenv("KUBECONFIG")

	// Load the kubeconfig file
	config, err := clientcmd.BuildConfigFromFlags("", kubeconfigPath)
	if err != nil {
		log.Println("Error loading kubeconfig", "ERROR", err)
		return nil, err
	}

	// Create the client
	client, err := kubernetes.NewForConfig(config)
	if err != nil {
		log.Println("Error creating kubernetes client", "ERROR", err)
		return nil, err
	}

	// if namespace "rovers" does not exist, create it
	_, err = client.CoreV1().Namespaces().Get(context.TODO(), "rovers", metav1.GetOptions{})
	if err != nil {
		_, err = client.CoreV1().Namespaces().Create(context.TODO(), &core.Namespace{
			ObjectMeta: metav1.ObjectMeta{
				Name: "rovers",
			},
		}, metav1.CreateOptions{})
		if err != nil {
			log.Println("Error creating namespace", "namespace", "rovers", "ERROR", err)
			return nil, err
		}
	}

	log.Println("Kubernetes client created", "kubeconfigPath", kubeconfigPath)

	return &Client{client: client}, nil
}

// GetPods returns a list of pods in the cluster given a namespace and label
func (c *Client) GetPods(namespace string) ([]string, error) {
	// Get the pods
	pods, err := c.client.CoreV1().Pods(namespace).List(context.TODO(), metav1.ListOptions{})
	if err != nil {
		return nil, err
	}

	// Get the pod names
	var podNames []string
	for _, pod := range pods.Items {
		podNames = append(podNames, pod.Name)
	}

	return podNames, nil
}

// GetPodStatus returns the status of a pod given a namespace and pod name
func (c *Client) GetPodStatus(namespace, podName string) (string, error) {
	// Get the pod
	pod, err := c.client.CoreV1().Pods(namespace).Get(context.TODO(), podName, metav1.GetOptions{})
	if err != nil {
		return "", err
	}

	return string(pod.Status.Phase), nil
}

// GetPodIP returns the IP of a pod given a namespace and pod name
func (c *Client) GetPodIP(namespace, podName string) (string, error) {
	// Get the pod
	pod, err := c.client.CoreV1().Pods(namespace).Get(context.TODO(), podName, metav1.GetOptions{})
	if err != nil {
		return "", err
	}

	return pod.Status.PodIP, nil
}

// Find pod details by name
func (c *Client) GetPod(namespace, podName string) (*core.Pod, error) {
	// Get the pod
	pod, err := c.client.CoreV1().Pods(namespace).Get(context.TODO(), podName, metav1.GetOptions{})
	if err != nil {
		return nil, err
	}

	return pod, nil
}

// DeletePod deletes a pod given a namespace and pod name
func (c *Client) DeletePod(namespace, podName string) error {
	// Delete the pod
	err := c.client.CoreV1().Pods(namespace).Delete(context.TODO(), podName, metav1.DeleteOptions{})
	if err != nil {
		return err
	}

	return nil
}

// CreatePod creates a pod given a namespace and optional pod configuration
func (c *Client) CreatePod(namespace string, podConfig *core.Pod) error {
	// Set a default pod configuration if podConfig is nil
	if podConfig == nil {
		podConfig = &core.Pod{
			ObjectMeta: metav1.ObjectMeta{
				Name: uuid.New().String(),
			},
			Spec: core.PodSpec{
				SecurityContext: &core.PodSecurityContext{
					RunAsUser:  func(i int64) *int64 { return &i }(0),
					RunAsGroup: func(i int64) *int64 { return &i }(0),
				},
				Containers: []core.Container{
					{
						Name:  "rover",
						Image: os.Getenv("ROVER_IMAGE"),
						Args: []string{
							"-rover",
						},
						Ports: []core.ContainerPort{
							{
								ContainerPort: 80,
							},
						},
						VolumeMounts: []core.VolumeMount{
							{
								Name:      "talosconfig",
								MountPath: "/etc/talos/config",
							},
						},
					},
				},
				// probably necessary to have the rover phone home
				Volumes: []core.Volume{
					{
						Name: "talosconfig",
						VolumeSource: core.VolumeSource{
							Secret: &core.SecretVolumeSource{
								SecretName: "talosconfig",
							},
						},
					},
				},
			},
		}
	}

	// Create the pod
	_, err := c.client.CoreV1().Pods(namespace).Create(context.TODO(), podConfig, metav1.CreateOptions{})
	if err != nil {
		log.Println("Error creating pod", "namespace", namespace, "podConfig", podConfig, "ERROR", err)
		return err
	}

	return nil
}

package k8s

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"strings"

	log "github.com/go-logr/logr"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
)

// StatefulSet the StatefulSet service that knows how to interact with k8s to manage them
type StatefulSet interface {
	GetStatefulSet(namespace, name string) (*appsv1.StatefulSet, error)
	GetStatefulSetPods(namespace, name string) (*corev1.PodList, error)
	CreateStatefulSet(namespace string, statefulSet *appsv1.StatefulSet) error
	UpdateStatefulSet(namespace string, statefulSet *appsv1.StatefulSet) error
	CreateOrUpdateStatefulSet(namespace string, statefulSet *appsv1.StatefulSet) error
	DeleteStatefulSet(namespace string, name string) error
	ListStatefulSets(namespace string) (*appsv1.StatefulSetList, error)
	GetStatefulSetReadPod(namespace, name string) ([]corev1.Pod, error)
}

// StatefulSetService is the service account service implementation using API calls to kubernetes.
type StatefulSetService struct {
	kubeClient kubernetes.Interface
	logger     log.Logger
}

// NewStatefulSetService returns a new StatefulSet KubeService.
func NewStatefulSetService(kubeClient kubernetes.Interface, logger log.Logger) *StatefulSetService {
	logger = logger.WithValues("service", "k8s.statefulSet")
	return &StatefulSetService{
		kubeClient: kubeClient,
		logger:     logger,
	}
}

// GetStatefulSet will retrieve the requested statefulset based on namespace and name
func (s *StatefulSetService) GetStatefulSet(namespace, name string) (*appsv1.StatefulSet, error) {
	statefulSet, err := s.kubeClient.AppsV1().StatefulSets(namespace).Get(context.TODO(), name, metav1.GetOptions{})
	if err != nil {
		return nil, err
	}
	return statefulSet, err
}

// GetStatefulSetPods will give a list of pods that are managed by the statefulset
func (s *StatefulSetService) GetStatefulSetPods(namespace, name string) (*corev1.PodList, error) {
	statefulSet, err := s.GetStatefulSet(namespace, name)
	if err != nil {
		return nil, err
	}
	labels := []string{}
	for k, v := range statefulSet.Spec.Selector.MatchLabels {
		labels = append(labels, fmt.Sprintf("%s=%s", k, v))
	}
	selector := strings.Join(labels, ",")
	return s.kubeClient.CoreV1().Pods(namespace).List(context.TODO(), metav1.ListOptions{LabelSelector: selector})
}

func (s *StatefulSetService) GetStatefulSetReadPod(namespace, name string) ([]corev1.Pod, error) {
	var podlist []corev1.Pod
	podList, err := s.GetStatefulSetPods(namespace, name)
	if err != nil {
		return podlist, err
	}
	num := 0
	for _, pod := range podList.Items {

		for _, c := range pod.Status.Conditions {
			if c.Type == "Ready" && c.Status == "True" {
				num = num + 1
				podlist = append(podlist, pod)
			}
		}

		//if len(pod.Status.Conditions) != 4 {
		//	continue
		//}
		//if pod.Status.Conditions[1].Type == "Ready" &&
		//	pod.Status.Conditions[1].Status == "True" {
		//	num = num + 1
		//	podlist = append(podlist, pod)
		//}
	}
	return podlist, nil
}

// CreateStatefulSet will create the given statefulset
func (s *StatefulSetService) CreateStatefulSet(namespace string, statefulSet *appsv1.StatefulSet) error {
	_, err := s.kubeClient.AppsV1().StatefulSets(namespace).Create(context.TODO(), statefulSet, metav1.CreateOptions{})
	if err != nil {
		return err
	}
	s.logger.WithValues("namespace", namespace).WithValues("statefulSet", statefulSet.ObjectMeta.Name).Info("statefulSet created")
	return err
}

// UpdateStatefulSet will update the given statefulset
func (s *StatefulSetService) UpdateStatefulSet(namespace string, statefulSet *appsv1.StatefulSet) error {
	_, err := s.kubeClient.AppsV1().StatefulSets(namespace).Update(context.TODO(), statefulSet, metav1.UpdateOptions{})
	if err != nil {
		return err
	}
	s.logger.WithValues("namespace", namespace).WithValues("statefulSet", statefulSet.ObjectMeta.Name).Info("statefulSet updated")
	return err
}

// CreateOrUpdateStatefulSet will update the statefulset or create it if does not exist
func (s *StatefulSetService) CreateOrUpdateStatefulSet(namespace string, statefulSet *appsv1.StatefulSet) error {
	storedStatefulSet, err := s.GetStatefulSet(namespace, statefulSet.Name)
	if err != nil {
		// If no resource we need to create.
		if errors.IsNotFound(err) {
			return s.CreateStatefulSet(namespace, statefulSet)
		}
		return err
	}

	// Already exists, need to Update.
	// Set the correct resource version to ensure we are on the latest version. This way the only valid
	// namespace is our spec(https://github.com/kubernetes/community/blob/master/contributors/devel/api-conventions.md#concurrency-control-and-consistency),
	// we will replace the current namespace state.

	switch checkSts(storedStatefulSet, statefulSet) {
	case Update:
		statefulSet.ResourceVersion = storedStatefulSet.ResourceVersion
		return s.UpdateStatefulSet(namespace, statefulSet)
	//updates to statefulset spec for fields other than 'replicas', 'template', and 'updateStrategy' are forbidden
	case Delete:
		return s.DeleteStatefulSet(namespace, storedStatefulSet.Name)
	}
	return nil
}

// DeleteStatefulSet will delete the statefulset
func (s *StatefulSetService) DeleteStatefulSet(namespace, name string) error {
	propagation := metav1.DeletePropagationForeground
	return s.kubeClient.AppsV1().StatefulSets(namespace).Delete(context.TODO(), name, metav1.DeleteOptions{PropagationPolicy: &propagation})
}

// ListStatefulSets will retrieve a list of statefulset in the given namespace
func (s *StatefulSetService) ListStatefulSets(namespace string) (*appsv1.StatefulSetList, error) {
	return s.kubeClient.AppsV1().StatefulSets(namespace).List(context.TODO(), metav1.ListOptions{})
}

type operator int

const (
	None operator = iota
	Update
	Delete
)

// check whether delete sts
func checkSts(old *appsv1.StatefulSet, new *appsv1.StatefulSet) operator {

	rsA, _ := json.Marshal(old.Spec.Template.Spec.Containers[0].Resources)
	rsB, _ := json.Marshal(new.Spec.Template.Spec.Containers[0].Resources)

	envA, _ := json.Marshal(old.Spec.Template.Spec.Containers[0].Env)
	envB, _ := json.Marshal(new.Spec.Template.Spec.Containers[0].Env)

	if checkVolumeClaimTemplates(old, new) {
		return Delete
	}

	if !bytes.Equal(rsA, rsB) || *old.Spec.Replicas != *new.Spec.Replicas || !bytes.Equal(envA, envB) {
		return Update
	}

	return None
}

// check whether delete sts
func checkVolumeClaimTemplates(old *appsv1.StatefulSet, new *appsv1.StatefulSet) bool {
	ov := old.Spec.VolumeClaimTemplates
	nv := new.Spec.VolumeClaimTemplates
	if len(ov) == 0 && len(nv) == 0 {
		return false
	}

	if len(ov) != len(nv) {
		return true
	}

	if len(ov) > 0 && len(nv) > 0 {
		vmA, _ := json.Marshal(old.Spec.VolumeClaimTemplates[0].Spec.Resources)
		vmB, _ := json.Marshal(new.Spec.VolumeClaimTemplates[0].Spec.Resources)

		oscn := old.Spec.VolumeClaimTemplates[0].Spec.StorageClassName
		nscn := new.Spec.VolumeClaimTemplates[0].Spec.StorageClassName
		return !bytes.Equal(vmA, vmB) || !strings.EqualFold(*oscn, *nscn)
	}

	return false
}

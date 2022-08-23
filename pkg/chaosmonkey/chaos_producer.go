package chaosmonkey

import (
	"context"
	"errors"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"log"
	"math/rand"
	"time"
)

var randSource = rand.NewSource(time.Now().UnixNano())
var randGen = rand.New(randSource)

type Workload struct {
	Namespace    string
	Schedule     time.Duration
	SelectConfig SelectConfig
	client       kubernetes.Interface
	SelfPodName  string
}

type SelectConfig struct {
	Labels string
}

func NewWorkload(namespace, schedule, selfPodName string, config SelectConfig, client kubernetes.Interface) (Workload, error) {

	parsedSchedule, err := time.ParseDuration(schedule)
	if err != nil {
		return Workload{}, err
	}
	return Workload{
		Namespace:    namespace,
		Schedule:     parsedSchedule,
		SelfPodName:  selfPodName,
		SelectConfig: config,
		client:       client,
	}, nil
}

func (wl *Workload) Start(ctx context.Context) <-chan error {
	ticker := time.NewTicker(wl.Schedule)
	errChan := make(chan error, 1)
	go func() {
		for {
			select {
			case <-ticker.C:
				log.Println("getting pod list...")
				podList, err := wl.getPodsList(ctx)
				if err != nil {
					errChan <- err
					break
				}
				podToTerminate, err := wl.getCandidatePodList(podList)
				if err != nil {
					errChan <- err
					break
				}
				index := randGen.Intn(len(podToTerminate))
				log.Println("index:", index)
				err = wl.deletePod(ctx, podToTerminate[index].Name)
				if err != nil {
					errChan <- err
				}
			case <-ctx.Done():
				ticker.Stop()
				close(errChan)
				return
			}
		}
	}()

	return errChan
}

func (wl *Workload) getPodsList(ctx context.Context) (*v1.PodList, error) {
	options := metav1.ListOptions{}
	if wl.SelectConfig.Labels != "" {
		options.LabelSelector = wl.SelectConfig.Labels
	}
	pods, err := wl.client.CoreV1().Pods(wl.Namespace).List(ctx, options)
	if err != nil {
		return nil, err
	}
	return pods, nil
}

func (wl *Workload) deletePod(ctx context.Context, name string) error {
	log.Println("deleting pod:", name)
	return wl.client.CoreV1().Pods(wl.Namespace).Delete(ctx, name, metav1.DeleteOptions{})
}

func (wl *Workload) getCandidatePodList(list *v1.PodList) ([]v1.Pod, error) {
	var candidateList []v1.Pod
	for _, pod := range list.Items {
		// skip our programme pod
		if pod.Name == wl.SelfPodName {
			continue
		}
		for _, cond := range pod.Status.Conditions {
			if cond.Type == v1.PodReady && cond.Status == v1.ConditionTrue {
				candidateList = append(candidateList, pod)
			}
		}
	}

	if len(candidateList) == 0 {
		return nil, errors.New("no matching pod found")
	}
	return candidateList, nil
}

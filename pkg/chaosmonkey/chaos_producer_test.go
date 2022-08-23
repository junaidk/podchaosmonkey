package chaosmonkey

import (
	"context"
	"github.com/stretchr/testify/assert"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/client-go/kubernetes/fake"
	"testing"
	"time"
)

func TestGetPods(t *testing.T) {
	podList := []runtime.Object{
		&corev1.Pod{
			ObjectMeta: metav1.ObjectMeta{
				Name:      "pod1",
				Namespace: "namespace1",
				Labels: map[string]string{
					"label1": "value1",
				},
			},
		},
		&corev1.Pod{
			ObjectMeta: metav1.ObjectMeta{
				Name:      "pod2",
				Namespace: "namespace1",
				Labels: map[string]string{
					"label1": "value2",
				},
			},
		},
		&corev1.Pod{
			ObjectMeta: metav1.ObjectMeta{
				Name:      "pod3",
				Namespace: "namespace1",
				Labels:    map[string]string{},
			},
		},
	}

	testCases := []struct {
		name             string
		pods             []runtime.Object
		targetNamespace  string
		targetPod        string
		targetLabel      string
		expectedPodCount int
	}{
		{
			name:             "get all pods",
			pods:             podList,
			targetNamespace:  "namespace1",
			targetLabel:      "",
			expectedPodCount: 3,
		},
		{
			name:             "get all pods with one label",
			pods:             podList,
			targetNamespace:  "namespace1",
			targetLabel:      "label1=value1",
			expectedPodCount: 1,
		},
		{
			name:             "get all pods with multiple labels",
			pods:             podList,
			targetNamespace:  "namespace1",
			targetLabel:      "label1 in (value1,value2)",
			expectedPodCount: 2,
		},
		{
			name:             "get no pod with no matching label",
			pods:             podList,
			targetNamespace:  "namespace1",
			targetLabel:      "label1 in (wrong_label)",
			expectedPodCount: 0,
		},
		{
			name:             "get no pod in wrong namespace",
			pods:             podList,
			targetNamespace:  "namespace2",
			targetLabel:      "label1=value1",
			expectedPodCount: 0,
		},
	}

	for _, test := range testCases {
		t.Run(test.name, func(t *testing.T) {
			fakeClientset := fake.NewSimpleClientset(test.pods...)
			wl, err := NewWorkload(test.targetNamespace, "1s", "", SelectConfig{
				Labels: test.targetLabel,
			}, fakeClientset)
			assert.NoError(t, err)
			podList, err := wl.getPodsList(context.Background())
			assert.NoError(t, err)
			assert.Equal(t, test.expectedPodCount, len(podList.Items))

		})
	}
}

func TestGetCandidatePod(t *testing.T) {
	podList := corev1.PodList{
		Items: []corev1.Pod{
			{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "pod1",
					Namespace: "namespace1",
					Labels: map[string]string{
						"label1": "value1",
					},
				},
				Status: corev1.PodStatus{
					Conditions: []corev1.PodCondition{
						{
							Type:   corev1.PodReady,
							Status: corev1.ConditionFalse,
						},
					},
				},
			},
			{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "pod2",
					Namespace: "namespace1",
					Labels: map[string]string{
						"label1": "value2",
					},
				},
				Status: corev1.PodStatus{
					Conditions: []corev1.PodCondition{
						{
							Type:   corev1.PodReady,
							Status: corev1.ConditionTrue,
						},
					},
				},
			},
			{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "self-pod",
					Namespace: "namespace1",
					Labels: map[string]string{
						"label1": "value2",
					},
				},
				Status: corev1.PodStatus{
					Conditions: []corev1.PodCondition{
						{
							Type:   corev1.PodReady,
							Status: corev1.ConditionTrue,
						},
					},
				},
			},
		},
	}

	wl, err := NewWorkload("namespace1", "1s", "self-pod", SelectConfig{}, nil)
	assert.NoError(t, err)
	candidatePods, err := wl.getCandidatePodList(&podList)
	assert.Equal(t, 1, len(candidatePods))

	candidatePods, err = wl.getCandidatePodList(&corev1.PodList{})
	assert.Error(t, err)
}

func TestWorkload_Start(t *testing.T) {
	namespace := "namespace1"
	podList := []runtime.Object{
		&corev1.Pod{
			ObjectMeta: metav1.ObjectMeta{
				Name:      "pod1",
				Namespace: namespace,
				Labels: map[string]string{
					"label1": "value1",
				},
			},
			Status: corev1.PodStatus{
				Conditions: []corev1.PodCondition{
					{
						Type:   corev1.PodReady,
						Status: corev1.ConditionTrue,
					},
				},
			},
		},
		&corev1.Pod{
			ObjectMeta: metav1.ObjectMeta{
				Name:      "pod2",
				Namespace: namespace,
				Labels: map[string]string{
					"label1": "value2",
				},
			},
			Status: corev1.PodStatus{
				Conditions: []corev1.PodCondition{
					{
						Type:   corev1.PodReady,
						Status: corev1.ConditionTrue,
					},
				},
			},
		},
		&corev1.Pod{
			ObjectMeta: metav1.ObjectMeta{
				Name:      "pod3",
				Namespace: namespace,
				Labels:    map[string]string{},
			},
			Status: corev1.PodStatus{
				Conditions: []corev1.PodCondition{
					{
						Type:   corev1.PodReady,
						Status: corev1.ConditionTrue,
					},
				},
			},
		},
		&corev1.Pod{
			ObjectMeta: metav1.ObjectMeta{
				Name:      "pod4",
				Namespace: namespace,
				Labels: map[string]string{
					"label1": "value2",
				},
			},
			Status: corev1.PodStatus{
				Conditions: []corev1.PodCondition{
					{
						Type:   corev1.PodReady,
						Status: corev1.ConditionTrue,
					},
				},
			},
		},
		&corev1.Pod{
			ObjectMeta: metav1.ObjectMeta{
				Name:      "pod5",
				Namespace: namespace,
				Labels: map[string]string{
					"label1": "value2",
				},
			},
			Status: corev1.PodStatus{
				Conditions: []corev1.PodCondition{
					{
						Type:   corev1.PodReady,
						Status: corev1.ConditionFalse,
					},
				},
			},
		},
	}

	fakeClientset := fake.NewSimpleClientset(podList...)
	wl, err := NewWorkload("namespace1", "1ms", "pod4", SelectConfig{
		Labels: "label1 in (value2,value1)",
	}, fakeClientset)
	assert.NoError(t, err)

	ctx, cancel := context.WithCancel(context.Background())
	errChan := wl.Start(ctx)

	go func() {
		for {
			select {
			case <-errChan:

			}
		}
	}()

	assert.Eventually(t, func() bool {
		pods, err := fakeClientset.CoreV1().Pods(namespace).List(ctx, metav1.ListOptions{})
		if err != nil {
			return false
		}
		if len(pods.Items) == 3 {
			return true
		}
		return false
	}, 1*time.Second, 100*time.Millisecond)
	cancel()

}

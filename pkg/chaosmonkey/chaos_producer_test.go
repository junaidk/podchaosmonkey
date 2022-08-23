package chaosmonkey

import (
	"context"
	"github.com/stretchr/testify/assert"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/client-go/kubernetes/fake"
	"testing"
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

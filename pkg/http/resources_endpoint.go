package http

import (
	"net/http"

	"github.com/datagravity-ai/keel/internal/k8s"
	"github.com/datagravity-ai/keel/internal/policy"
  
  "github.com/datagravity-ai/keel/provider/kubernetes"

)

type resource struct {
	Provider    string            `json:"provider"`
	Identifier  string            `json:"identifier"`
	Name        string            `json:"name"`
	Namespace   string            `json:"namespace"`
	Kind        string            `json:"kind"`
	Policy      string            `json:"policy"`
	Images      []string          `json:"images"`
	Labels      map[string]string `json:"labels"`
	Annotations map[string]string `json:"annotations"`
	Status      k8s.Status        `json:"status"`
}

func (s *TriggerServer) resourcesHandler(resp http.ResponseWriter, req *http.Request) {

	vals := s.grc.Values()

	var res []resource

	for _, v := range vals {

		p := policy.GetPolicyFromLabelsOrAnnotations(v.GetLabels(), v.GetAnnotations())
		filterFunc := kubernetes.GetMonitorContainersFromMeta(v.GetLabels(), v.GetAnnotations())

		res = append(res, resource{
			Provider:    "kubernetes",
			Identifier:  v.Identifier,
			Name:        v.Name,
			Namespace:   v.Namespace,
			Kind:        v.Kind(),
			Policy:      p.Name(),
			Labels:      v.GetLabels(),
			Annotations: v.GetAnnotations(),
			Images:      v.GetImages(filterFunc),
			Status:      v.GetStatus(),
		})
	}

	response(res, 200, nil, resp, req)
}

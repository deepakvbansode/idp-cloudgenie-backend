package k8swatcher

import (
	"context"
	"encoding/json"
	"time"

	"github.com/deepakvbansode/idp-cloudgenie-backend/internal/adaptors/driven/mongo"
	"github.com/deepakvbansode/idp-cloudgenie-backend/internal/common/k8s"
	"github.com/deepakvbansode/idp-cloudgenie-backend/internal/core/entities"
	"github.com/deepakvbansode/idp-cloudgenie-backend/internal/core/ports"
	"github.com/google/uuid"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/apimachinery/pkg/watch"
	"k8s.io/client-go/dynamic"
)

type ResourceWatcher struct {
	logger      ports.Logger
	repoAdaptor *mongo.RepositoryAdaptor
}

func NewResourceWatcher(logger ports.Logger, repoAdaptor *mongo.RepositoryAdaptor) *ResourceWatcher {
	return &ResourceWatcher{
		logger:      logger,
		repoAdaptor: repoAdaptor,
	}
}
// WatchXRDInstances watches all XRD instances with label blueprint-name and logs their status with trace UUID
func (r *ResourceWatcher) WatchXRDInstances(ctx context.Context) error {
       defer func(r *ResourceWatcher) {
	       if err := recover(); err != nil {
		       r.logger.Error("Recovered from panic in WatchXRDInstances: ", err)
	       }
       }(r)
       traceID := uuid.New().String()
       r.logger = r.logger.WithField("trace_id", traceID)
       config, err := k8s.GetKubeConfig()
       if err != nil {
	       r.logger.Error("failed to get kubeconfig: ", err)
	       return err
       }
       dynClient, err := dynamic.NewForConfig(config)
       if err != nil {
	       r.logger.Error("failed to create dynamic client: ", err)
	       return err
       }

       // 1. List all XRDs (CompositeResourceDefinitions)
       crdGVR := schema.GroupVersionResource{
	       Group:    "apiextensions.crossplane.io",
	       Version:  "v1",
	       Resource: "compositeresourcedefinitions",
       }
       crdList, err := dynClient.Resource(crdGVR).List(ctx, metav1.ListOptions{
			LabelSelector: "blueprint-name",	
	   })
       if err != nil {
	       r.logger.Error("failed to list XRDs: ", err)
	       return err
       }

       // 2. For each XRD, get GVK and watch its instances
       for _, xrd := range crdList.Items {
	       spec, found, _ := unstructured.NestedMap(xrd.Object, "spec")
	       if !found {
		       continue
	       }
	       group, _, _ := unstructured.NestedString(spec, "group")
	       names, _, _ := unstructured.NestedMap(spec, "names")
	       plural, _, _ := unstructured.NestedString(names, "plural")
	       scope, _, _ := unstructured.NestedString(spec, "scope")

	       // Find the version with referenceable=true
	       var version string
	       if versions, ok, _ := unstructured.NestedSlice(spec, "versions"); ok {
		       for _, v := range versions {
			       vmap, ok := v.(map[string]interface{})
			       if !ok {
				       continue
			       }
			       ref, ok := vmap["referenceable"].(bool)
			       if ok && ref {
				       if name, ok := vmap["name"].(string); ok {
					       version = name
					       break
				       }
			       }
		       }
		       
	       }

	       if group == "" || plural == "" || version == "" {
		       continue
	       }

	       gvr := schema.GroupVersionResource{
		       Group:    group,
		       Version:  version,
		       Resource: plural,
	       }
	       go r.watchCompositeResource(ctx, dynClient, gvr, scope)
       }
       // Block forever (or until context is cancelled)
       <-ctx.Done()
       return nil
}

func (r *ResourceWatcher) watchCompositeResource(ctx context.Context, dynClient dynamic.Interface, gvr schema.GroupVersionResource, scope string) {
       defer func(r *ResourceWatcher) {
	       if err := recover(); err != nil {
		       r.logger.Error("Recovered from panic in watchCompositeResource: ", err)
	       }
       }(r)
      
       for {
	       select {
	       case <-ctx.Done():
		       r.logger.Info("Context cancelled, stopping watcher for ", gvr.String())
		       return
	       default:
	       }
	       var watcher watch.Interface
	       var err error
	       if scope == "Namespaced" {
		       watcher, err = dynClient.Resource(gvr).Namespace("").Watch(ctx, metav1.ListOptions{})
	       } else {
		       watcher, err = dynClient.Resource(gvr).Watch(ctx, metav1.ListOptions{})
	       }
	       if err != nil {
		       r.logger.Error("Failed to watch ", gvr.String(), ": ", err)
		       time.Sleep(10 * time.Second)
		       continue
	       }
	       r.logger.Info("Watching ", gvr.String(), " for resources with label blueprint-name...")
	       for {
		       select {
		       case <-ctx.Done():
			       r.logger.Info("Context cancelled, stopping event loop for ", gvr.String())
			       watcher.Stop()
			       return
		       case event, ok := <-watcher.ResultChan():
			       if !ok {
				       r.logger.Info("Watcher channel closed for ", gvr.String())
				       return
			       }
			       u, ok := event.Object.(*unstructured.Unstructured)
			       if !ok {
				       continue
			       }
		       status, found, _ := unstructured.NestedFieldNoCopy(u.Object, "status")
		       if found {
			       statusJSON, _ := json.MarshalIndent(status, "", "  ")
			       r.logger.Info(gvr.Resource, " ", u.GetName(), " status: ", string(statusJSON))
			       // Unmarshal status into entities.ResourceStatus struct
			       var resourceStatus entities.ResourceStatus
			       if err := json.Unmarshal(statusJSON, &resourceStatus); err != nil {
				       r.logger.Error("Failed to unmarshal status for ", u.GetName(), ": ", err)
				       continue
			       }
			       if err := r.repoAdaptor.UpdateResourceStatus(ctx, u.GetName(), resourceStatus); err != nil {
				       r.logger.Error("Failed to update resource status in MongoDB for ", u.GetName(), ": ", err)
			       }
		       } else {
			       r.logger.Info(gvr.Resource, " ", u.GetName(), " has no status yet")
		       }
		       }
	       }
	       
       }
}

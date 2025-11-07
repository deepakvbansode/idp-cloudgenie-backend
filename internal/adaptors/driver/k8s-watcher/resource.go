package k8swatcher

import (
	"context"
	"encoding/json"
	"time"

	"github.com/deepakvbansode/idp-cloudgenie-backend/internal/common/k8s"
	"github.com/deepakvbansode/idp-cloudgenie-backend/internal/core/ports"
	"github.com/google/uuid"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/apimachinery/pkg/watch"
	"k8s.io/client-go/dynamic"
)

// WatchXRDInstances watches all XRD instances with label blueprint-name and logs their status with trace UUID
func WatchXRDInstances(ctx context.Context, logger ports.Logger) error {
       defer func() {
	       if r := recover(); r != nil {
		       logger.Error("Recovered from panic in WatchXRDInstances: ", r)
	       }
       }()
       traceID := uuid.New().String()
       logger = logger.WithField("trace_id", traceID)
       config, err := k8s.GetKubeConfig()
       if err != nil {
	       logger.Error("failed to get kubeconfig: ", err)
	       return err
       }
       dynClient, err := dynamic.NewForConfig(config)
       if err != nil {
	       logger.Error("failed to create dynamic client: ", err)
	       return err
       }

       // 1. List all XRDs (CompositeResourceDefinitions)
       crdGVR := schema.GroupVersionResource{
	       Group:    "apiextensions.crossplane.io",
	       Version:  "v1",
	       Resource: "compositeresourcedefinitions",
       }
       crdList, err := dynClient.Resource(crdGVR).List(ctx, metav1.ListOptions{})
       if err != nil {
	       logger.Error("failed to list XRDs: ", err)
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
	       version, _, _ := unstructured.NestedString(spec, "versions", "0", "name")
	       scope, _, _ := unstructured.NestedString(spec, "scope")
	       if group == "" || plural == "" || version == "" {
		       continue
	       }
	       gvr := schema.GroupVersionResource{
		       Group:    group,
		       Version:  version,
		       Resource: plural,
	       }
	       go watchCompositeResource(ctx, dynClient, gvr, scope, logger)
       }
       // Block forever (or until context is cancelled)
       <-ctx.Done()
       return nil
}

func watchCompositeResource(ctx context.Context, dynClient dynamic.Interface, gvr schema.GroupVersionResource, scope string, logger ports.Logger) {
       defer func() {
	       if r := recover(); r != nil {
		       logger.Error("Recovered from panic in watchCompositeResource: ", r)
	       }
       }()
       labelSelector := "blueprint-name"
       for {
	       select {
	       case <-ctx.Done():
		       logger.Info("Context cancelled, stopping watcher for ", gvr.String())
		       return
	       default:
	       }
	       var watcher watch.Interface
	       var err error
	       if scope == "Namespaced" {
		       watcher, err = dynClient.Resource(gvr).Namespace("").Watch(ctx, metav1.ListOptions{LabelSelector: labelSelector})
	       } else {
		       watcher, err = dynClient.Resource(gvr).Watch(ctx, metav1.ListOptions{LabelSelector: labelSelector})
	       }
	       if err != nil {
		       logger.Error("Failed to watch ", gvr.String(), ": ", err)
		       time.Sleep(10 * time.Second)
		       continue
	       }
	       logger.Info("Watching ", gvr.String(), " for resources with label blueprint-name...")
	       for {
		       select {
		       case <-ctx.Done():
			       logger.Info("Context cancelled, stopping event loop for ", gvr.String())
			       watcher.Stop()
			       return
		       case event, ok := <-watcher.ResultChan():
			       if !ok {
				       logger.Info("Watcher channel closed for ", gvr.String())
				       return
			       }
			       u, ok := event.Object.(*unstructured.Unstructured)
			       if !ok {
				       continue
			       }
			       status, found, _ := unstructured.NestedFieldNoCopy(u.Object, "status")
			       if found {
				       statusJSON, _ := json.MarshalIndent(status, "", "  ")
				       logger.Info(gvr.Resource, " ", u.GetName(), " status: ", string(statusJSON))
			       } else {
				       logger.Info(gvr.Resource, " ", u.GetName(), " has no status yet")
			       }
		       }
	       }
	       // If the watcher channel closes, restart the watch
	       time.Sleep(2 * time.Second)
       }
}

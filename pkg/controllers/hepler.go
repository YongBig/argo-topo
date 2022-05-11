/**
 * @Author: dy
 * @Description:
 * @File: hepler
 * @Date: 2022/5/7 17:09
 */
package controllers

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/YongBig/argo-topo/pkg/resourcetree"
	"github.com/argoproj/argo-cd/v2/pkg/apis/application"
	"github.com/argoproj/argo-cd/v2/pkg/apis/application/v1alpha1"
	clustercache "github.com/argoproj/gitops-engine/pkg/cache"
	"github.com/argoproj/gitops-engine/pkg/utils/kube"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/klog/v2"
	ctrlruntimeclient "sigs.k8s.io/controller-runtime/pkg/client"
)

type Helper struct {
	K8sClient       ctrlruntimeclient.Client
	K8sClusterCache clustercache.ClusterCache
}

func NewHelper(k8sClient ctrlruntimeclient.Client, k8sClusterCache clustercache.ClusterCache) *Helper {
	return &Helper{
		K8sClient:       k8sClient,
		K8sClusterCache: k8sClusterCache,
	}
}

func (argo *Helper) GetManagedResources(name string, namespace string) (*v1alpha1.Application, []*unstructured.Unstructured, error) {
	app, err := argo.getApplication(name, namespace)
	if err != nil {
		return nil, nil, err
	}

	managedResources, err := argo.parseSpecManifests(app)
	if err != nil {
		return nil, nil, err
	}

	return app, managedResources, nil
}

func (argo *Helper) getApplication(name string, namespace string) (*v1alpha1.Application, error) {
	var app v1alpha1.Application
	err := argo.K8sClient.Get(context.TODO(), ctrlruntimeclient.ObjectKey{
		Namespace: namespace,
		Name:      name,
	}, &app)
	app.TypeMeta = metav1.TypeMeta{
		Kind:       application.ApplicationKind,
		APIVersion: v1alpha1.ApplicationSchemaGroupVersionKind.GroupVersion().String(),
	}
	return &app, err
}
func (argo *Helper) parseSpecManifests(app *v1alpha1.Application) ([]*unstructured.Unstructured, error) {
	var managedManifests []*unstructured.Unstructured
	if len(app.Status.Resources) == 0 {
		return managedManifests, nil
	}
	// argo engin cluster cache filter resouce query criteria
	// Query criteria do not match. Resource not found
	for _, objRef := range app.Status.Resources {
		var managedManifest unstructured.Unstructured
		managedManifest.SetNamespace(objRef.Namespace)
		managedManifest.SetName(objRef.Name)
		managedManifest.SetKind(objRef.Kind)
		managedManifest.SetAPIVersion(objRef.GroupVersionKind().GroupVersion().String())
		managedManifests = append(managedManifests, &managedManifest)
	}

	return managedManifests, nil
}

func (argo *Helper) CacheKey(app *v1alpha1.Application) string {
	return fmt.Sprintf("%s/%s", app.Namespace, app.Name)
}

func (argo *Helper) BuildResourceTreeForApp(app *v1alpha1.Application) (*resourcetree.ResourceTree, error) {
	managedResources, err := argo.parseSpecManifests(app)
	if err != nil {
		return nil, err
	}
	return argo.buildResourceTree(app, managedResources)

}

func (argo *Helper) buildResourceTree(app *v1alpha1.Application, managedResources []*unstructured.Unstructured) (*resourcetree.ResourceTree, error) {
	a, _ := json.Marshal(managedResources)
	fmt.Println(string(a))
	gvk := app.GroupVersionKind()
	root := resourcetree.NewTreeNode(&resourcetree.ResourceRef{
		Group:     gvk.Group,
		Version:   gvk.Version,
		Kind:      gvk.Kind,
		Namespace: app.GetNamespace(),
		Name:      app.GetName(),
		Uid:       string(app.GetUID()),
	})
	tree := resourcetree.NewResourceTree(root)
	desiredObjects, err := argo.K8sClusterCache.GetManagedLiveObjs(managedResources, func(_ *clustercache.Resource) bool {

		return true
	})
	fmt.Println(desiredObjects)
	if err != nil {
		return nil, err
	}
	for i := range managedResources {
		managedResource := managedResources[i]
		desiredObject, ok := desiredObjects[kube.ResourceKey{
			Group:     managedResource.GroupVersionKind().Group,
			Kind:      managedResource.GroupVersionKind().Kind,
			Namespace: managedResource.GetNamespace(),
			Name:      managedResource.GetName(),
		}]
		if !ok {
			klog.Warningf("spec %s %s manifests missing desired manifests",
				managedResource.GroupVersionKind().String(), managedResource.GetName())
			continue
		}

		tree.Root.AddNode(resourcetree.NewTreeNode(resourcetree.CreateResourceRefFromUnstructured(desiredObject)))
	}

	// dfs 深度优先
	// todo: bfs广度优先
	for i := 0; i < len(tree.Root.Children); i++ {
		argo.dfs(tree.Root.Children[i])
	}
	return tree, nil
}

func (argo *Helper) dfs(node *resourcetree.TreeNode) {
	resourceMap := make(map[string]*clustercache.Resource, 0)
	// note 迭代资源数 for argocd engine
	argo.K8sClusterCache.IterateHierarchy(kube.ResourceKey{
		Group:     node.ResourceRef.Group,
		Kind:      node.ResourceRef.Kind,
		Namespace: node.ResourceRef.Namespace,
		Name:      node.ResourceRef.Name,
	}, func(resource *clustercache.Resource, children map[kube.ResourceKey]*clustercache.Resource) bool {
		for _, child := range children {
			if isParentOf(resource, child) {
				resourceMap[string(resource.Ref.UID)] = child
			}
		}
		return false
	})

	for uid, child := range resourceMap {
		fmt.Println(child.Ref.Namespace, "/", child.Ref.Name)
		node.AddNode(resourcetree.NewTreeNode(&resourcetree.ResourceRef{
			Group:     child.Ref.GroupVersionKind().Group,
			Version:   child.Ref.GroupVersionKind().Version,
			Kind:      child.Ref.Kind,
			Namespace: child.Ref.Namespace,
			Name:      child.Ref.Name,
			Uid:       uid,
		}))
	}

	for i := 0; i < len(node.Children); i++ {
		argo.dfs(node.Children[i])
	}
}

func isParentOf(r *clustercache.Resource, child *clustercache.Resource) bool {
	for i, ownerRef := range child.OwnerRefs {

		// backfill UID of inferred owner child references
		if ownerRef.UID == "" && r.Ref.Kind == ownerRef.Kind && r.Ref.APIVersion == ownerRef.APIVersion && r.Ref.Name == ownerRef.Name {
			ownerRef.UID = r.Ref.UID
			child.OwnerRefs[i] = ownerRef
			return true
		}

		if r.Ref.UID == ownerRef.UID {
			return true
		}
	}

	return false
}

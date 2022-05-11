/**
 * @Author: dy
 * @Description:
 * @File: config
 * @Date: 2022/5/7 11:31
 */
package Config

import (
	"fmt"
	"github.com/YongBig/argo-topo/pkg/scheme"
	ArgoEngineCache "github.com/argoproj/gitops-engine/pkg/cache"
	"k8s.io/client-go/rest"
	ctrl "sigs.k8s.io/controller-runtime"
	manager "sigs.k8s.io/controller-runtime"
)

type Config struct {
	Manager                manager.Manager
	ArgoEngineClusterCache ArgoEngineCache.ClusterCache
}

func NewConfig(Rconfig *rest.Config) *Config {
	c := &Config{}
	managerI, err := ctrl.NewManager(Rconfig, ctrl.Options{
		Scheme: scheme.Scheme, //argo v1alpha1 scheme add
	})
	if err != nil {
		panic(err)
	}
	k8sClusterCache := ArgoEngineCache.NewClusterCache(Rconfig)
	a := k8sClusterCache.GetAPIResources()
	fmt.Println(a)
	c.Manager = managerI
	c.ArgoEngineClusterCache = k8sClusterCache
	return c
}

/**
 * @Author: dy
 * @Description:
 * @File: App
 * @Date: 2022/5/7 16:29
 */
package controllers

import (
	"context"
	"github.com/argoproj/argo-cd/v2/pkg/apis/application/v1alpha1"
	clustercache "github.com/argoproj/gitops-engine/pkg/cache"
	"k8s.io/apimachinery/pkg/api/errors"
	ctrlruntimeclient "sigs.k8s.io/controller-runtime/pkg/client"
)

// new App Controller
func NewArgoAppController(
	config Config,
	kubeClient ctrlruntimeclient.Client,
	clusterCache clustercache.ClusterCache,
) *Controller {
	return &Controller{
		Config:       config,
		ClusterCache: clusterCache,
		Helper: func() *Helper {
			return NewHelper(kubeClient, clusterCache)
		}(),
	}
}

func (c *Controller) GetApp(ctx context.Context, name, namespace string) (*v1alpha1.Application, error) {
	app := &v1alpha1.Application{}
	err := c.Helper.K8sClient.Get(ctx, ctrlruntimeclient.ObjectKey{Name: name, Namespace: namespace}, app)
	if errors.IsNotFound(err) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return app, nil
}

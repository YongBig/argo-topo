/**
 * @Author: dy
 * @Description:
 * @File: controller
 * @Date: 2022/5/7 16:31
 */
package controllers

import (
	"context"
	"github.com/YongBig/argo-topo/pkg/DB"
	"github.com/argoproj/argo-cd/v2/pkg/apis/application/v1alpha1"
	clustercache "github.com/argoproj/gitops-engine/pkg/cache"
	"k8s.io/klog/v2"
	ctrlruntime "sigs.k8s.io/controller-runtime"
	ctrlcontroller "sigs.k8s.io/controller-runtime/pkg/controller"
)

type Controller struct {
	Config       Config
	ClusterCache clustercache.ClusterCache
	Helper       *Helper
}

func (c *Controller) SetupManager(mgr ctrlruntime.Manager) error {
	return ctrlruntime.NewControllerManagedBy(mgr).
		Named("argo_topo").
		WithOptions(ctrlcontroller.Options{
			MaxConcurrentReconciles: c.Config.MaxConcurrentReconciles,
			CacheSyncTimeout:        c.Config.CacheSyncTimeout,
			RecoverPanic:            true,
		}).
		For(&v1alpha1.Application{}).
		Complete(c)
}

func (c *Controller) Reconcile(ctx context.Context, request ctrlruntime.Request) (ctrlruntime.Result, error) {
	//sync catch for argo
	if err := c.ClusterCache.EnsureSynced(); err != nil {
		return ctrlruntime.Result{}, err
	}
	app, err := c.GetApp(ctx, request.Name, request.Namespace)
	if app == nil && err == nil {
		return ctrlruntime.Result{}, nil
	}

	if err != nil {
		return ctrlruntime.Result{}, err
	}

	if !app.DeletionTimestamp.IsZero() {
		// 尝试清理缓存
		cacheKey := c.Helper.CacheKey(app)

		ok, err := DB.RDB.Del(cacheKey)
		if ok && err != nil {
			klog.Errorf("delete %s toplogy cache error: %v", cacheKey, err)
			// TODO: 对 application 设置 finalizer, 直到 cache 清理成功后在去除 finalizer
			return ctrlruntime.Result{}, nil
		}

		return ctrlruntime.Result{}, nil
	}

	appTopology, err := c.Helper.BuildResourceTreeForApp(app)

	if err != nil {
		klog.Errorf("build %s application toplogy error: %v", request.NamespacedName, err)
		return ctrlruntime.Result{}, err
	}

	klog.Infof("build %s application toplogy success", request.NamespacedName)
	// TODO: 应该 diff? trade-off diff 和 direct-write-cache, 当前总是 direct-write-cache
	cacheKey := c.Helper.CacheKey(app)
	err = DB.RDB.Set(cacheKey, appTopology.Marshal())
	if err != nil {
		klog.Errorf("set %s toplogy cache error: %v", cacheKey, err)
		return ctrlruntime.Result{}, err
	}
	klog.Infof("set %s application toplogy cache success", cacheKey)
	return ctrlruntime.Result{}, nil
}

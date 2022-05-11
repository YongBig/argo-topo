/**
 * @Author: dy
 * @Description:
 * @File: scheme
 * @Date: 2022/5/10 16:40
 */
package scheme

import (
	"github.com/argoproj/argo-cd/v2/pkg/apis/application/v1alpha1"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime/schema"
	utilruntime "k8s.io/apimachinery/pkg/util/runtime"

	"k8s.io/apimachinery/pkg/runtime"
)

var localSchemeBuilder = runtime.SchemeBuilder{
	v1alpha1.AddToScheme,
}
var AddToScheme = localSchemeBuilder.AddToScheme

var Scheme = runtime.NewScheme()

func init() {
	v1.AddToGroupVersion(Scheme, schema.GroupVersion{Version: "v1"})
	utilruntime.Must(AddToScheme(Scheme))

}

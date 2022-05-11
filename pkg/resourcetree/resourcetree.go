/**
 * @Author: dy
 * @Description:
 * @File: resourcetree
 * @Date: 2022/5/10 15:51
 */
package resourcetree

import (
	"encoding/json"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/klog/v2"
)

type ResourceTree struct {
	Root *TreeNode `json:"root,omitempty"`
}
type TreeNode struct {
	ResourceRef *ResourceRef ` json:"resourceRef,omitempty"`
	Children    []*TreeNode  `json:"children,omitempty"`
}

type ResourceRef struct {
	Group     string `json:"group,omitempty"`
	Version   string `json:"version,omitempty"`
	Kind      string `json:"kind,omitempty"`
	Namespace string `json:"namespace,omitempty"`
	Name      string `json:"name,omitempty"`
	Uid       string `json:"uid,omitempty"`
}

// NewTreeNode create tree node with given resourceRef
func NewTreeNode(resourceRef *ResourceRef) *TreeNode {
	return &TreeNode{ResourceRef: resourceRef, Children: make([]*TreeNode, 0)}
}

func NewResourceTree(root *TreeNode) *ResourceTree {
	return &ResourceTree{Root: root}
}
func (t *TreeNode) AddNode(node *TreeNode) {
	t.Children = append(t.Children, node)
}

func CreateResourceRefFromUnstructured(un *unstructured.Unstructured) *ResourceRef {
	gvk := un.GroupVersionKind()
	return &ResourceRef{
		Group:     gvk.Group,
		Version:   gvk.Version,
		Kind:      gvk.Kind,
		Namespace: un.GetNamespace(),
		Name:      un.GetName(),
	}
}

func (r *ResourceTree) Marshal() []byte {
	s, e := json.Marshal(r)
	if e != nil {
		klog.Errorf("ResourceTree Marshal is err : %v", e)
		return []byte("")
	}
	return s
}

package k8s

import (
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/apimachinery/pkg/selection"
	v1 "k8s.io/client-go/listers/apps/v1"
)

type FunctionList struct {
	deployLister      v1.DeploymentLister
	namespace         string
	functionsSelector labels.Selector
}

func NewFunctionList(namespace string, deployLister v1.DeploymentLister) *FunctionList {

	sel := labels.NewSelector()
	req, _ := labels.NewRequirement("faas_function", selection.Exists, []string{})
	onlyFunctions := sel.Add(*req)

	return &FunctionList{
		deployLister:      deployLister,
		namespace:         namespace,
		functionsSelector: onlyFunctions,
	}
}

func (f *FunctionList) Count() (int, error) {
	list, err := f.deployLister.Deployments(f.namespace).List(f.functionsSelector)
	if err != nil {
		return 0, err
	}

	return len(list), nil
}

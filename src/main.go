// Copyright (c) Microsoft Corporation.
// Licensed under the MIT License.

package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"

	"net/http"

	wfclientset "github.com/argoproj/argo/pkg/client/clientset/versioned"
	corev1 "k8s.io/api/core/v1"
	apis "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/runtime"
	"k8s.io/client-go/informers"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/cache"
)

const SUBSCRUBERS_FOLDER = "/subscribers"

func main() {
	// use the current context in kubeconfig
	config, err := rest.InClusterConfig()
	panicErr(err)

	// create the clientset
	clientset, err := kubernetes.NewForConfig(config)
	panicErr(err)

	namespace := os.Getenv("POD_NAMESPACE")

	//create a workflow client
	workflowclientset := wfclientset.NewForConfigOrDie(config).ArgoprojV1alpha1()
	workflowClient := workflowclientset.Workflows(namespace)
	panicErr(err)

	factory := informers.NewSharedInformerFactoryWithOptions(clientset, 0, informers.WithNamespace(namespace))
	informer := factory.Core().V1().Events().Informer()

	stopper := make(chan struct{})
	defer close(stopper)
	defer runtime.HandleCrash()

	informer.AddEventHandler(cache.ResourceEventHandlerFuncs{
		AddFunc: func(obj interface{}) {
			event := obj.(*corev1.Event)
			if event.Source.Component == "workflow-controller" {
				workflow, err := workflowClient.Get(event.InvolvedObject.Name, apis.GetOptions{})
				if err == nil {
					workflowEvent := WorkflowEvent{event, workflow}
					fmt.Println(workflowEvent.getEventMessage())
					sendToSubscribers(workflowEvent.getEventMessage())
				} else {
					fmt.Println(err)
				}

			}
		}})

	go informer.Run(stopper)
	if !cache.WaitForCacheSync(stopper, informer.HasSynced) {
		runtime.HandleError(fmt.Errorf("Timed out waiting for caches to sync"))
		return
	}
	<-stopper
}

func sendMessage(uri string, eventMessage *EventMessage) {
	eventMessageStruct := *eventMessage
	reqBody, err := json.Marshal(eventMessageStruct)
	if err == nil {
		resp, _ := http.Post(uri, "application/json", bytes.NewBuffer(reqBody))
		fmt.Println("HTTP Response Status:", uri, resp.StatusCode, http.StatusText(resp.StatusCode))
	} else {
		fmt.Println(err)
	}

}

func sendToSubscribers(eventMessage *EventMessage) {
	files, err := ioutil.ReadDir(SUBSCRUBERS_FOLDER)
	if err == nil {
		for _, file := range files {
			if !file.IsDir() {
				fmt.Println(file.Name())
				dat, err := ioutil.ReadFile(filepath.Join(SUBSCRUBERS_FOLDER, file.Name()))
				if err == nil {
					uri := string(dat)
					fmt.Println(uri)
					sendMessage(uri, eventMessage)
				} else {
					fmt.Println(err)
				}
			}
		}
	} else {
		fmt.Println(err)
	}

}

func panicErr(e error) {
	if e != nil {
		panic(e)
	}
}

// Copyright (c) Microsoft Corporation.
// Licensed under the MIT License.

package main

import (
	"strings"

	wfv1 "github.com/argoproj/argo/pkg/apis/workflow/v1alpha1"
	corev1 "k8s.io/api/core/v1"
)

type EventMessage struct {
	KfpRunId     string
	WorkflowName string
	NodeName     string
	EventType    string
	Message      string
}

type WorkflowEvent struct {
	event    *corev1.Event
	workflow *wfv1.Workflow
}

func (workflowEvent *WorkflowEvent) getNodeName() string {
	words := strings.Split(workflowEvent.event.Message, workflowEvent.workflow.Name+".")
	wordsLen := len(words)
	if wordsLen > 1 {
		dotSepar := strings.Split(words[wordsLen-1], ".")
		return dotSepar[len(dotSepar)-1]
	}
	return ""
}

func (workflowEvent *WorkflowEvent) getKfpRunID() string {
	return workflowEvent.workflow.ObjectMeta.Labels["pipeline/runid"]
}

func (workflowEvent *WorkflowEvent) getEventMessage() *EventMessage {
	eventMessage := EventMessage{
		workflowEvent.getKfpRunID(),
		workflowEvent.workflow.Name,
		workflowEvent.getNodeName(),
		workflowEvent.event.Reason,
		workflowEvent.event.Message}
	return &eventMessage
}

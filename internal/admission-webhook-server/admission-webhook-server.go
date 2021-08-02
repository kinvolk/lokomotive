// Copyright 2021 The Lokomotive Authors
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

// Package admissionwebhookserver contains code for admission webhook.
package admissionwebhookserver

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"

	"github.com/golang/glog"
	v1 "k8s.io/api/admission/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes/scheme"
)

// ServeMutateServiceAccount is a service to mutate default service account.
func ServeMutateServiceAccount(w http.ResponseWriter, r *http.Request) {
	serve(w, r, mutateServiceAccount)
}

type patchOperation struct {
	Op    string      `json:"op"`
	Path  string      `json:"path"`
	Value interface{} `json:"value,omitempty"`
}

type admitFunc func(v1.AdmissionReview) *v1.AdmissionResponse

func toAdmissionResponse(err error) *v1.AdmissionResponse {
	return &v1.AdmissionResponse{
		Result: &metav1.Status{
			Message: err.Error(),
		},
	}
}

func readRequest(r *http.Request) ([]byte, error) {
	body := []byte{}

	if r.Body == nil {
		return body, fmt.Errorf("empty body")
	}

	return ioutil.ReadAll(r.Body)
}

func readAndValidateRequest(r *http.Request) ([]byte, error) {
	body, err := readRequest(r)
	if err != nil {
		return nil, err
	}

	// Verify the content type is accurate.
	contentType := r.Header.Get("Content-Type")
	if contentType != "application/json" {
		return nil, fmt.Errorf("Content-Type=%s, expected application/json", contentType)
	}

	return body, nil
}

func serve(w http.ResponseWriter, r *http.Request, admit admitFunc) {
	body, err := readAndValidateRequest(r)
	if err != nil {
		glog.Error(err.Error())
		http.Error(w, err.Error(), http.StatusBadRequest)
	}

	// The AdmissionReview that was sent to the webhook.
	requestedAdmissionReview := v1.AdmissionReview{}

	// The AdmissionReview that will be returned.
	responseAdmissionReview := v1.AdmissionReview{}

	deserializer := scheme.Codecs.UniversalDeserializer()

	if _, _, err := deserializer.Decode(body, nil, &requestedAdmissionReview); err != nil {
		glog.Error(err)
		responseAdmissionReview.Response = toAdmissionResponse(err)
	} else {
		responseAdmissionReview.Response = admit(requestedAdmissionReview)
	}

	responseAdmissionReview.Response.UID = requestedAdmissionReview.Request.UID
	responseAdmissionReview.APIVersion = "admission.k8s.io/v1"
	responseAdmissionReview.Kind = "AdmissionReview"

	respBytes, err := json.Marshal(responseAdmissionReview)
	if err != nil {
		glog.Errorf("marshaling json data: %v", err)
	}

	if _, err := w.Write(respBytes); err != nil {
		glog.Errorf("writing response data: %v", err)
	}
}

func mutateServiceAccount(ar v1.AdmissionReview) *v1.AdmissionResponse {
	req := ar.Request

	glog.Infof("AdmissionReview for Kind=%v, Namespace=%v Name=%v UID=%v patchOperation=%v UserInfo=%v",
		req.Kind, req.Namespace, req.Name, req.UID, req.Operation, req.UserInfo)

	if req.Kind.Kind != "ServiceAccount" || req.Name != "default" {
		glog.Infof("Skipping mutation for Kind=%v Name=%v", req.Kind, req.Name)

		return &v1.AdmissionResponse{
			Allowed: true,
		}
	}

	reviewResponse := v1.AdmissionResponse{}
	reviewResponse.Allowed = true

	patch := []patchOperation{
		{
			Op:    "add",
			Path:  "/automountServiceAccountToken",
			Value: false,
		},
	}

	patchFinal, err := json.Marshal(patch)
	if err != nil {
		glog.Fatalf("marshaling patch data %v:", err)
	}

	reviewResponse.Patch = patchFinal
	pt := v1.PatchTypeJSONPatch
	reviewResponse.PatchType = &pt

	glog.Infof("AdmissionResponse: patch=%v\n", string(patchFinal))

	return &reviewResponse
}

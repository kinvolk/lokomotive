/*
Copyright 2020 The Flux authors

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package v1beta1

const SourceFinalizer = "finalizers.fluxcd.io"

const (
	// URLInvalidReason represents the fact that a given source has an invalid URL.
	URLInvalidReason string = "URLInvalid"

	// StorageOperationFailedReason signals a failure caused by a storage operation.
	StorageOperationFailedReason string = "StorageOperationFailed"

	// AuthenticationFailedReason represents the fact that a given secret does not
	// have the required fields or the provided credentials do not match.
	AuthenticationFailedReason string = "AuthenticationFailed"

	// VerificationFailedReason represents the fact that the cryptographic
	// provenance verification for the source failed.
	VerificationFailedReason string = "VerificationFailed"
)

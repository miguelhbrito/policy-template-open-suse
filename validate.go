package main

import (
	"fmt"

	onelog "github.com/francoispqt/onelog"
	corev1 "github.com/kubewarden/k8s-objects/api/core/v1"
	kubewarden "github.com/kubewarden/policy-sdk-go"
	kubewarden_protocol "github.com/kubewarden/policy-sdk-go/protocol"
	"github.com/mailru/easyjson"
)

func validate(payload []byte) ([]byte, error) {
	// Create a ValidationRequest instance from the incoming payload
	validationRequest := kubewarden_protocol.ValidationRequest{}
	err := easyjson.Unmarshal(payload, &validationRequest)
	if err != nil {
		return kubewarden.RejectRequest(
			kubewarden.Message(err.Error()),
			kubewarden.Code(400))
	}

	// Create a Settings instance from the ValidationRequest object
	settings, err := NewSettingsFromValidationReq(&validationRequest)
	if err != nil {
		return kubewarden.RejectRequest(
			kubewarden.Message(err.Error()),
			kubewarden.Code(400))
	}

	// Access the **raw** JSON that describes the object
	podJSON := validationRequest.Request.Object

	// Try to create a Pod instance using the RAW JSON we got from the
	// ValidationRequest.
	pod := &corev1.Pod{}
	if err := easyjson.Unmarshal([]byte(podJSON), pod); err != nil {
		return kubewarden.RejectRequest(
			kubewarden.Message(
				fmt.Sprintf("Cannot decode Pod object: %s", err.Error())),
			kubewarden.Code(400))
	}

	logger.Debug("validating labels")
	// Validates every label key to check if is a palindrome
	for key := range pod.Metadata.Labels {
		if settings.IsPalindrome(key) {
			logger.InfoWithFields("rejecting label", func(e onelog.Entry) {
				e.String("label key", key)
			})

			return kubewarden.RejectRequest(
				kubewarden.Message(
					fmt.Sprintf("The label key '%s' is a palindrome", key)),
				kubewarden.NoCode)
		}
	}

	return kubewarden.AcceptRequest()
}

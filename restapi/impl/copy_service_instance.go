package impl

import (
	"github.com/go-openapi/runtime/middleware"
	"ocopea/k8sdsb/models"
	"ocopea/k8sdsb/restapi/operations/dsb_web"
	k8sClient "ocopea/kubernetes/client"
	"log"
)

func CopyServiceInstance(
k8s k8sClient.ClientInterface,
params dsb_web.CopyServiceInstanceParams) middleware.Responder {

	log.Println("Faking copy yey!")
	return dsb_web.NewCopyServiceInstanceOK().WithPayload(&models.CopyServiceInstanceResponse{
		CopyID:        *params.CopyDetails.CopyID,
		Status:        0,
		StatusMessage: "yey",
	})
}


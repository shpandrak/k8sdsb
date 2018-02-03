package impl

import (
	"ocopea/k8sdsb/models"
	"ocopea/k8sdsb/restapi/operations/dsb_web"
	"github.com/go-openapi/runtime/middleware"
)

func DsbInfoResponse() middleware.Responder {

	return dsb_web.NewGetDSBInfoOK().WithPayload(
		&models.DsbInfo{
			Name: "k8s-dsb",
			Description: "Docker containers DSB for kubernetes",
			Type:  "datasource",
			Plans: []*models.DsbPlan{
				{
					ID: "mysql",
					Name: "Mysql",
					Description: "Single container mysql deployment",
					DsbSettings: map[string]string{
						"imageName":"mysql",
						"imageVersion":"5.6",
						"containerPort":"3306",
						"dataFolder": "/var/lib/mysql",
					},
					CopyProtocols: []*models.DsbSupportedCopyProtocol{
						{
							CopyProtocol: "ShpanRest",
							CopyProtocolVersion: "1.0",
						},
					},
					Protocols: []*models.DsbSupportedProtocol{
						{
							Protocol: "mysql",
						},
					},
				},
			},
		})

}

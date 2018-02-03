package impl

import (
	"ocopea/k8sdsb/models"
	"ocopea/k8sdsb/restapi/operations/dsb_web"
	"github.com/go-openapi/runtime/middleware"
	k8sClient "ocopea/kubernetes/client"
	"fmt"
	"ocopea/kubernetes/client/v1"
	"log"
)

type K8sDsbBindings struct {
	Host     string `json:"host"`
	Password string `json:"password"`
	Port     string `json:"port"`
}

func DsbGetServiceInstancesResponse(
k8s *k8sClient.Client,
params dsb_web.GetServiceInstanceParams) middleware.Responder {

	d, err := getBindingInfoForInstance(k8s, params.InstanceID)
	if (err != nil) {
		return getError(dsb_web.NewGetServiceInstanceDefault(500), err, 500)
	}

	return dsb_web.NewGetServiceInstanceOK().WithPayload(d)

}

func getBindingInfoForInstance(k8s *k8sClient.Client, instanceId string) (*models.ServiceInstanceDetails, error) {
	serviceName := getServiceNameFromInstanceId(instanceId)
	log.Printf("testing service %s ", serviceName)
	isReady, svc, err := k8s.TestService(serviceName)
	if (err != nil) {
		return nil, err
	}

	// Still creating
	if (!isReady) {
		return &models.ServiceInstanceDetails{
			InstanceID: instanceId,
			State: "CREATING",
		}, nil
	}



	port := svc.Spec.Ports[0].Port
	var host string
	bindingInfo := make(map[string]string)
	bindingInfo["port"] = fmt.Sprintf("%d", port)

	//todo: use secrets!
//	bindingInfo["username"] = "nazuser"
//	bindingInfo["password"] = "nazpassword"
	bindingInfo["username"] = "root"
	bindingInfo["password"] = "nazgul123"
	//bindingInfo["password"] = ""
	if (svc.Spec.Type == v1.ServiceTypeClusterIP) {
		host = svc.Spec.ClusterIP
		bindingInfo["host"] = host
		//bindingInfo["host"] = serviceName
	} else {
		return nil, fmt.Errorf("Unsupported k8s service type %s for service %s", svc.Spec.Type, serviceName);
	}

	// Verify mongo is alive
	addr := fmt.Sprintf("%s:%d", host, port)
	log.Printf("TESTING for %s to start accepting connections %s", serviceName, addr)

	// todo: verify container started and listening on port
	// either by trying the port or by testing the pod state on k8s or by enabling some sort of custom verify command per-plan

	p := "tcp"
	int32Port := int32(port)
	return &models.ServiceInstanceDetails{
		InstanceID: instanceId,
		State: "RUNNING",
		Binding: bindingInfo,
		BindingPorts:[]*models.BindingPort{
			{
				Protocol:&p,
				Destination:&host,
				Port: &int32Port},
		},
		Size:500,
		StorageType: "Kubernetes Temp Volume",
	}, nil
}
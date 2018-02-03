package impl

import (
	"ocopea/k8sdsb/models"
	"ocopea/k8sdsb/restapi/operations/dsb_web"
	"github.com/go-openapi/runtime/middleware"
	"log"
	"ocopea/kubernetes/client/v1"
	"ocopea/kubernetes/client/types"
	k8sClient "ocopea/kubernetes/client"
	"strconv"
	"fmt"
)

func CreateInstanceResponse(
k8s *k8sClient.Client,
params *dsb_web.CreateServiceInstanceParams) middleware.Responder {

	err := createDsbInstance(k8s, params.ServiceSettings)
	if (err != nil) {
		return getError(dsb_web.NewCreateServiceInstanceDefault(500), err, 500)
	}

	return dsb_web.NewCreateServiceInstanceOK().WithPayload(
		&models.ServiceInstance{
			InstanceID: *params.ServiceSettings.InstanceID,
		})
}

func getError(failedResponse ErrorResponse, err error, errCode int) middleware.Responder {
	log.Printf("error occured %d - %s", errCode, err.Error())
	modelErrorInt := int32(errCode)
	errStr := err.Error()
	modelsError := models.Error{Code: &modelErrorInt, Message: &errStr}
	failedResponse.SetPayload(&modelsError)
	return failedResponse

}

//todo:better
func getServiceNameFromInstanceId(instanceId string) string {
	l := len(instanceId)
	if l > 22 {
		return "m-" + instanceId[l - 22:l]
	} else {
		return "m-" + instanceId
	}
}

func createDsbInstance(k8sClient *k8sClient.Client, serviceInstanceInfo *models.CreateServiceInstance) error {
	plan := serviceInstanceInfo.InstanceSettings["plan"];
	if (len(plan) == 0) {
		return fmt.Errorf("plan was not defined when trying to create service %s", *serviceInstanceInfo.InstanceID)
	}

	log.Printf("Creating DSB %s with plan %s", *serviceInstanceInfo.InstanceID, plan);
	return createDockerServiceInstance(k8sClient, serviceInstanceInfo, plan)
}

func createDockerServiceInstance(k8sClient *k8sClient.Client, createServiceInstanceParams *models.CreateServiceInstance, plan string) error {
	appUniqueName := getServiceNameFromInstanceId(*createServiceInstanceParams.InstanceID)

	var replicas int = 1;
	// Building rc spec

	imageName := createServiceInstanceParams.InstanceSettings["imageName"];
	imageVersion := createServiceInstanceParams.InstanceSettings["imageVersion"];
	if len(imageVersion) > 0 {
		imageName += ":" + imageVersion
	}

	containerPort, err := strconv.Atoi(createServiceInstanceParams.InstanceSettings["containerPort"]);
	if err != nil {
		return fmt.Errorf("Failed parsing container port for plan %s when trying to create %s - %s",
			plan, createServiceInstanceParams.InstanceID, err.Error())
	}

	spec := v1.ReplicationControllerSpec{}
	spec.Replicas = &replicas
	spec.Selector = make(map[string]string)
	spec.Selector["app"] = appUniqueName
	spec.Template = &v1.PodTemplateSpec{}
	spec.Template.ObjectMeta = v1.ObjectMeta{}
	spec.Template.ObjectMeta.Labels = make(map[string]string)
	spec.Template.ObjectMeta.Labels["k8s-docker-dsb"] = "yeah"
	spec.Template.ObjectMeta.Labels["app"] = appUniqueName
	spec.Template.ObjectMeta.Labels["nazKind"] = "dsbInstance"
	containerSpec := v1.Container{}
	containerSpec.Name = appUniqueName
	containerSpec.Image = imageName
	containerSpec.Ports = []v1.ContainerPort{{ContainerPort:containerPort}}

	//todo: 1) make this secret
	//todo: 2) make this parametrized in the plan rather constant for mysql!
	containerSpec.Env = []v1.EnvVar{
		{
			Name: "MYSQL_ROOT_PASSWORD",
			Value: "nazgul123",
		},
	}

	containers := []v1.Container{containerSpec}

	spec.Template.Spec = v1.PodSpec{}
	spec.Template.Spec.Containers = containers



	// Create a replicationController object for running the app
	rc := &v1.ReplicationController{}
	rc.Name = appUniqueName;
	rc.Labels = make(map[string]string)
	rc.Labels["nazKind"] = "dsbInstance"
	rc.Labels["k8s-docker-dsb"] = "yeah"
	rc.Spec = spec

	_, err = k8sClient.CreateReplicationController(rc, false)
	if err != nil {
		return err
	}

	svc := &v1.Service{}
	svc.Name = appUniqueName
	svc.Spec.Type = v1.ServiceTypeClusterIP;
	svc.Spec.Ports = []v1.ServicePort{{
		Port:containerPort,
		TargetPort: types.NewIntOrStringFromInt(containerPort),
		Protocol:v1.ProtocolTCP,
		Name:"tcp",
	}}
	svc.Spec.Selector = map[string]string{"app": appUniqueName}
	svc, err = k8sClient.CreateService(svc, false)

	return err

}




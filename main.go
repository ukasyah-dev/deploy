package main

import (
	"context"
	"fmt"
	"time"

	"github.com/appleboy/graceful"
	"github.com/caitlinelfring/go-env-default"
	"github.com/gofiber/fiber/v2"
	"github.com/ukasyah-dev/common/log"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
)

var httpPort = env.GetIntDefault("HTTP_PORT", 3000)
var inCluster = env.GetBoolDefault("IN_CLUSTER", false)
var kubeconfig = env.GetDefault("KUBECONFIG", "~/.kube/config")

func main() {
	var config *rest.Config
	var err error

	if inCluster {
		config, err = rest.InClusterConfig()
		if err != nil {
			panic(err)
		}
	} else {
		config, err = clientcmd.BuildConfigFromFlags("", kubeconfig)
		if err != nil {
			panic(err)
		}
	}

	clientset, err := kubernetes.NewForConfig(config)
	if err != nil {
		panic(err)
	}

	app := fiber.New()

	app.Get("/", func(c *fiber.Ctx) error {
		return c.JSON(fiber.Map{"status": "ok"})
	})

	app.Post("/:name", func(c *fiber.Ctx) error {
		deploymentName := c.Params("name")

		client := clientset.AppsV1().Deployments("default")
		data := fmt.Sprintf(
			`{"spec": {"template": {"metadata": {"annotations": {"kubectl.kubernetes.io/restartedAt": "%s"}}}}}`,
			time.Now().Format("20060102150405"),
		)
		deployment, err := client.Patch(c.Context(), deploymentName, types.StrategicMergePatchType, []byte(data), v1.PatchOptions{})
		if err != nil {
			log.Errorf("Failed to patch deployment: %s", err)
			return c.JSON(fiber.Map{"message": "Failed to patch deployment"})
		}

		return c.JSON(fiber.Map{"message": "Deployment patched", "deployment": deployment})
	})

	m := graceful.NewManager()

	m.AddRunningJob(func(ctx context.Context) error {
		return app.Listen(fmt.Sprintf(":%d", httpPort))
	})

	m.AddShutdownJob(func() error {
		return app.Shutdown()
	})
}

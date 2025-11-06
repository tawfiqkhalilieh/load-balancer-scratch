package main

import "github.com/gin-gonic/gin"
import "fmt"
import "net/http"

func healthChecks(address string, servicesStatus map[string]bool) {
	resp, err := http.Get(address)
	if err != nil || resp.StatusCode != 200 {

		servicesStatus[address] = false 
	}
	if resp.Body != nil {
		defer resp.Body.Close()
	}
	servicesStatus[address] = true
}

func setupServices(serviceURL string, deployedServices int) []string {
	services := []string{}
	for i := 1; i <= deployedServices; i++ {
		serviceURL := fmt.Sprintf("http://%s:%d", serviceURL, 8000+i)
		services = append(services, serviceURL)
	}
	return services
}

func main() {
	deployedServices := setupServices("localhost", 3)
  servicesStatus := make(map[string]bool)

	for _, service := range deployedServices {
		go healthChecks(service, servicesStatus)
	}

	router := gin.Default()

	router.GET("/", func(c *gin.Context) {
		c.JSON(200, gin.H{
			"message": "This is a load balancer example",
		})
	})

	router.GET("/health", func(c *gin.Context) {
		c.JSON(200, gin.H{
			"health": "ok",
		})
	})

	router.GET("/message", func(c *gin.Context) {

	})

	router.Run()
}

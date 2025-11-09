package main

import "github.com/gin-gonic/gin"
import "fmt"
import "net/http"
import "io"
import "log"

func callTask(c *gin.Context, service string) {
	resp, err := http.Post(fmt.Sprintf("%s/task", service),
		"application/json", nil,
	)

	if err != nil {
		c.JSON(500, gin.H{
			"error": err.Error(),
		})
		return
	}

	defer resp.Body.Close()
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		c.JSON(500, gin.H{
			"error": "service unavailable",
		})
		return
	}
	c.Data(resp.StatusCode, "application/json", body)
}

func roundRobin(services []string, counter *int) string {
	service := services[*counter]
	*counter++

	if *counter >= len(services) {
		*counter = 0
	}

	return service
}

func healthChecks(address string, servicesStatus *map[string]bool, conns *map[string]int) {
	objectValue := *servicesStatus
	connsValue := *conns
	resp, err := http.Get(address)
	if err != nil || resp.StatusCode != 200 {
		objectValue[address] = false
	} else {
		objectValue[address] = true
		connsValue[address] = 0
	}
}

func setupServices(serviceURL string, deployedServices int) []string {
	services := []string{}
	for i := 1; i <= deployedServices; i++ {
		serviceURL := fmt.Sprintf("http://%s:%d", serviceURL, 4000+i)
		services = append(services, serviceURL)
	}
	return services
}

func getAvailabilityStatus(service string) bool {
	res, err := http.Get(service + "/health")
	if err != nil || res.StatusCode != 200 {
		return false
	}
	defer res.Body.Close()
	if res.StatusCode == http.StatusOK {
		res, err := http.Get(service + "/busy")
		if err != nil || res.StatusCode != 200 {
			return false
		}

		defer res.Body.Close()

		if res.StatusCode == http.StatusOK {
			bodyBytes, err := io.ReadAll(res.Body)
			if err != nil {
				log.Fatal(err)
			}
			bodyString := string(bodyBytes)
			fmt.Println("Service", service, "busy status:", bodyString)
			return bodyString == "\"no\""
		}
	}
	return false
}

func getLeastCons(conns *map[string]int) string {
	connsValue := *conns
	minConns := 0
	minConnsService := ""

	fmt.Println("Current connections:", connsValue)
	for name, v := range connsValue {
		if minConnsService != "" {
			if v < minConns {
				minConns = v
				minConnsService = name
			}
		} else {
			minConns = v
			minConnsService = name
		}
	}

	return minConnsService
}

func main() {
	counter := 0
	conns := make(map[string]int)

	patterns := make(map[string]bool)
	patterns["roundrobin"] = false
	patterns["leastconn"] = true

	deployedServices := setupServices("localhost", 2)
	servicesStatus := make(map[string]bool)

	for _, service := range deployedServices {
		go healthChecks(service, &servicesStatus, &conns)
	}

	fmt.Println("********************")
	fmt.Println("Deployed services:", conns)

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

	router.GET("/admin/services", func(c *gin.Context) {
		// if c.Query("key") != "admin" {
		// 	c.JSON(401, gin.H{
		// 		"error": "unauthorized",
		// 	})
		// 	return
		// }

		c.JSON(200, servicesStatus)
	})

	router.GET("/admin/conns", func(c *gin.Context) {
		// if c.Query("key") != "admin" {
		// 	c.JSON(401, gin.H{
		// 		"error": "unauthorized",
		// 	})
		// 	return
		// }

		c.JSON(200, conns)
	})

	router.GET("/message", func(c *gin.Context) {
		if patterns["roundrobin"] {
			service := roundRobin(deployedServices, &counter)
			for !getAvailabilityStatus(service) {
				service = roundRobin(deployedServices, &counter)
			}
			callTask(c, service)
		} else if patterns["leastconn"] {
			// if len(conns) == 0 {
			// 	c.JSON(500, gin.H{
			// 		"error": "no available service",
			// 	})
			// 	return
			// }

			service := getLeastCons(&conns)

			fmt.Println("Routing to service:", service, "with current connections:", conns[service])

			callTask(c, service)

		}
	})

	router.Run()
}

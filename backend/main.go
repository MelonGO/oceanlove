package main

import (
	"context"
	"fmt"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/redis/go-redis/v9"
)

var (
	server      *gin.Engine
	ctx         context.Context
	redisclient *redis.Client
)

func init() {
	// ? Create a context
	ctx = context.TODO()

	// ? Connect to Redis
	redisclient = redis.NewClient(&redis.Options{
		Addr: "redis:6379",
	})

	if _, err := redisclient.Ping(ctx).Result(); err != nil {
		fmt.Println(err.Error())
	}

	_, err := redisclient.Get(ctx, "gift").Result()

	if err == redis.Nil {
		err := redisclient.Set(ctx, "gift", 0, 0).Err()
		if err != nil {
			fmt.Println(err.Error())
		}
	} else if err != nil {
		fmt.Println(err.Error())
	}

	fmt.Println("Redis client connected successfully...")

	// ? Create the Gin Engine instance
	server = gin.Default()
}

func serveHome(c *gin.Context) {
	if c.Request.URL.Path != "/" {
		c.String(http.StatusNotFound, "Not Found")
		return
	}
	if c.Request.Method != http.MethodGet {
		c.String(http.StatusMethodNotAllowed, "Not Allowed")
		return
	}

	value, err := redisclient.Get(ctx, "gift").Result()

	if err == redis.Nil {
		fmt.Println("key: test does not exist")
	} else if err != nil {
		fmt.Println(err.Error())
	}

	c.HTML(http.StatusOK, "index.html", gin.H{"gift": value})
}

func main() {
	hub := newHub()
	go hub.run()

	server.Static("/textures", "./textures")
	server.Static("/node_modules", "./node_modules")
	server.Static("/static", "./static")
	server.Static("/models", "./models")

	server.LoadHTMLGlob("templates/*")

	server.GET("/", serveHome)
	server.GET("/ws", func(c *gin.Context) {
		serveWs(hub, c.Writer, c.Request)
	})

	server.GET("/sendGift", func(c *gin.Context) {
		value, err := redisclient.Get(ctx, "gift").Result()

		if err != nil {
			fmt.Println(err.Error())
		} else {
			num, _ := strconv.Atoi(value)
			num += 1
			err = redisclient.Set(ctx, "gift", num, 0).Err()
			if err != nil {
				fmt.Println(err.Error())
			}
			c.String(http.StatusOK, strconv.Itoa(num))
		}
	})

	server.GET("/fetchStatus", func(c *gin.Context) {
		gift, err := redisclient.Get(ctx, "gift").Result()
		if err != nil {
			fmt.Println(err.Error())
		}

		msg := gift
		c.String(http.StatusOK, msg)
	})

	server.Run(":8080")
}

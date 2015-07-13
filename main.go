package main

import (
	"crypto/tls"
	"github.com/BurntSushi/toml"
	"github.com/garyburd/redigo/redis"
	"github.com/gin-gonic/gin"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"time"
)

var (
	pool       *redis.Pool
	execDir    = "./"
	stopChan   = make(chan bool)
	conf       Config
	tls_config = tls.Config{
		MinVersion:               tls.VersionTLS12,
		PreferServerCipherSuites: true,
		CipherSuites: []uint16{tls.TLS_ECDHE_ECDSA_WITH_AES_128_GCM_SHA256,
			tls.TLS_ECDHE_RSA_WITH_AES_128_GCM_SHA256,
			tls.TLS_ECDHE_RSA_WITH_AES_256_CBC_SHA,
			tls.TLS_ECDHE_RSA_WITH_AES_128_CBC_SHA,
			tls.TLS_ECDHE_ECDSA_WITH_AES_256_CBC_SHA,
			tls.TLS_ECDHE_ECDSA_WITH_AES_128_CBC_SHA,
		},
	}
)

type Config struct {
	ListenHost    string
	ListenHostTLS string
	SSLCert       string
	SSLPrivateKey string
	Debug         bool
	RedisServer   string
	RedisPass     string
}

func newPool(server, password string) *redis.Pool {
	return &redis.Pool{
		MaxIdle:     3,
		IdleTimeout: 240 * time.Second,
		Dial: func() (redis.Conn, error) {
			c, err := redis.Dial("tcp", server)
			if err != nil {
				return nil, err
			}
			if _, err := c.Do("AUTH", password); err != nil {
				c.Close()
				return nil, err
			}
			return c, err
		},
		TestOnBorrow: func(c redis.Conn, t time.Time) error {
			_, err := c.Do("PING")
			return err
		},
	}
}

func blogHandler(ctx *gin.Context) {
	ctx.String(200, "hi")
}

func makeRouter() *gin.Engine {
	if !conf.Debug {
		gin.SetMode(gin.ReleaseMode)
	}
	router := gin.Default()
	router.GET("/", blogHandler)
	return router
}

func main() {
	router := makeRouter()

	go http.ListenAndServe(conf.ListenHost, router)
	srv := http.Server{TLSConfig: &tls_config, Addr: conf.ListenHostTLS, Handler: router}

	srv.ListenAndServeTLS(conf.SSLCert, conf.SSLPrivateKey)
	for {
		select {
		case <-stopChan:
			log.Print("Exiting!")
			return
		}
	}
}

func init() {
	var err error
	execDir, err = filepath.Abs(filepath.Dir(os.Args[0]))
	if err != nil {
		log.Fatal(err)
	}

	config_file := os.Getenv("CONFIG")
	if config_file == "" {
		config_file = "./config.toml"
	}
	config_data, err := ioutil.ReadFile(config_file)
	if err != nil {
		log.Fatal("Could not read config file: ", err)
	}
	if _, err := toml.Decode(string(config_data), &conf); err != nil {
		log.Println("Failed to decode config file")
	}

	pool = newPool(conf.RedisServer, conf.RedisPass)

}

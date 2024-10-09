package main

import (
	"context"
	"crypto/tls"
	"crypto/x509"
	"errors"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/stdlib"
	"github.com/jmoiron/sqlx"
	"github.com/minio/minio-go/v7"
	"github.com/minio/minio-go/v7/pkg/credentials"
	"github.com/redis/go-redis/v9"
)

func NewTLSConfig() (*tls.Config, error) {
	cert, err := tls.LoadX509KeyPair("./config/cert/client.crt", "./config/cert/client.key")
	if err != nil {
		return nil, err
	}

	caCert, err := os.ReadFile("./config/cert/server.crt")
	if err != nil {
		return nil, err
	}
	caCertPool := x509.NewCertPool()
	caCertPool.AppendCertsFromPEM(caCert)
	return &tls.Config{
		MinVersion:         tls.VersionTLS12,
		InsecureSkipVerify: true,
		RootCAs:            caCertPool,
		Certificates:       []tls.Certificate{cert},
	}, nil
}

func NewRedis(tlsConfig *tls.Config) (*redis.Client, error) {
	url, ok := os.LookupEnv("REDIS_URL")
	if !ok {
		return nil, errors.New("REDIS_URL is missing")
	}

	opts, err := redis.ParseURL(url)
	if err != nil {
		return nil, err
	}

	client := redis.NewClient(&redis.Options{
		Addr:      opts.Addr,
		Username:  opts.Username,
		Password:  opts.Password,
		TLSConfig: tlsConfig,
	})

	if err := client.Ping(context.Background()).Err(); err != nil {
		return nil, err
	}

	return client, nil
}

func NewMinio(tlsConfig *tls.Config) (*minio.Client, error) {
	url, ok := os.LookupEnv("MINIO_URL")
	if !ok {
		return nil, errors.New("MINIO_URL is missing")
	}

	accessKey, ok := os.LookupEnv("MINIO_ACCESS_KEY")
	if !ok {
		return nil, errors.New("MINIO_ACCESS_KEY is missing")
	}

	secretKey, ok := os.LookupEnv("MINIO_SECRET_KEY")
	if !ok {
		return nil, errors.New("MINIO_SECRET_KEY is missing")
	}

	client, err := minio.New(url, &minio.Options{
		Creds:  credentials.NewStaticV4(accessKey, secretKey, ""),
		Secure: true,
		Transport: &http.Transport{
			TLSClientConfig: tlsConfig,
		},
	})
	if err != nil {
		return nil, err
	}

	if _, err := client.HealthCheck(time.Second); err != nil {
		return nil, err
	}

	return client, nil
}

func NewPostgres(tlsConfig *tls.Config) (*sqlx.DB, error) {
	url, ok := os.LookupEnv("POSTGRES_URL")
	if !ok {
		return nil, errors.New("POSTGRES_URL is missing")
	}

	connConfig, _ := pgx.ParseConfig(url)
	connConfig.TLSConfig = tlsConfig
	connStr := stdlib.RegisterConnConfig(connConfig)

	client, err := sqlx.Connect("pgx", connStr)
	if err != nil {
		return nil, err
	}

	if err := client.Ping(); err != nil {
		return nil, err
	}

	return client, nil
}

func main() {
	tlsConfig, err := NewTLSConfig()
	if err != nil {
		log.Fatalln("tls config", err.Error())
	}
	if _, err := NewMinio(tlsConfig); err != nil {
		log.Fatalln("minio", err.Error())
	}

	if _, err := NewPostgres(tlsConfig); err != nil {
		log.Fatalln("postgres", err.Error())
	}
	if _, err := NewRedis(tlsConfig); err != nil {
		log.Fatalln("redis", err.Error())
	}
}

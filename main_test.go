package main

import (
	"bytes"
	"fmt"
	"log"
	"net/http"
	"os"
	"strings"
	"testing"

	"github.com/ory/dockertest/v3"
	"github.com/ory/dockertest/v3/docker"
)

var pool *dockertest.Pool

func TestMain(m *testing.M) {
	var err error
	pool, err = dockertest.NewPool("")
	if err != nil {
		log.Fatalf("could not construct pool: %s", err)
	}

	if err = pool.Client.Ping(); err != nil {
		log.Fatalf("could not connect to Docker: %s", err)
	}

	m.Run()
}

func TestStdout(t *testing.T) {

	resource, err := pool.Run("hello-world", "latest", []string{})
	if err != nil {
		t.Fatalf("could not start resource: %s", err)
	}
	defer pool.Purge(resource)

	// ログを取得
	var logBuffer bytes.Buffer
	err = pool.Client.Logs(docker.LogsOptions{
		Container:    resource.Container.ID,
		OutputStream: &logBuffer,
		ErrorStream:  &logBuffer,
		Stdout:       true,
		Stderr:       true,
	})
	if err != nil {
		t.Fatalf("could not get logs: %s", err)
	}

	output := logBuffer.String()
	// "Hello from Docker!" が含まれるか確認
	if !strings.Contains(output, "Hello from Docker!") {
		t.Fatalf("not contained 'Hello from Docker!' in output: %s", output)
	}
}

func TestNginx(t *testing.T) {

	// nginxコンテナを起動
	resource, err := pool.RunWithOptions(&dockertest.RunOptions{
		Repository:   "nginx",
		Tag:          "latest",
		ExposedPorts: []string{"80/tcp"},
	})
	if err != nil {
		t.Fatalf("could not start resource: %s", err)
	}
	defer pool.Purge(resource)

	// 動的に割り当てられたポートを取得
	hostPort := resource.GetPort("80/tcp")

	// Docker Compose内で実行する場合はhost.docker.internalを使用
	dockerHost := os.Getenv("DOCKER_HOST_ADDR")
	if dockerHost == "" {
		dockerHost = "localhost"
	}
	url := fmt.Sprintf("http://%s:%s", dockerHost, hostPort)

	// nginxが起動するまで待機してリトライ
	if err := pool.Retry(func() error {
		resp, err := http.Get(url)
		if err != nil {
			return err
		}
		defer resp.Body.Close()

		if resp.StatusCode != http.StatusOK {
			return fmt.Errorf("expected status 200, got %d", resp.StatusCode)
		}
		return nil
	}); err != nil {
		t.Fatalf("could not connect to nginx: %s", err)
	}

	// GET / リクエストしてステータス200を確認
	resp, err := http.Get(url)
	if err != nil {
		t.Fatalf("could not make request: %s", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		t.Fatalf("expected status 200, got %d", resp.StatusCode)
	}
}

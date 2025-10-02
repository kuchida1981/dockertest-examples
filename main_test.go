package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"path/filepath"
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

func TestMultipleServices(t *testing.T) {
	// サービスBを起動
	pwd, err := os.Getwd()
	if err != nil {
		t.Fatalf("failed to get working directory: %s", err)
	}

	serviceBResource, err := pool.BuildAndRunWithOptions(
		filepath.Join(pwd, "service-b", "Dockerfile"),
		&dockertest.RunOptions{
			Name:         "service-b",
			ExposedPorts: []string{"8080/tcp"},
		},
		func(config *docker.HostConfig) {
			config.AutoRemove = true
		},
	)
	if err != nil {
		t.Fatalf("could not start service B: %s", err)
	}
	defer pool.Purge(serviceBResource)

	serviceBPort := serviceBResource.GetPort("8080/tcp")

	// Docker Compose内で実行する場合はhost.docker.internalを使用
	dockerHost := os.Getenv("DOCKER_HOST_ADDR")
	if dockerHost == "" {
		dockerHost = "localhost"
	}

	serviceBURL := fmt.Sprintf("http://%s:%s", dockerHost, serviceBPort)

	// サービスBが起動するまで待機
	if err := pool.Retry(func() error {
		resp, err := http.Get(serviceBURL)
		if err != nil {
			return err
		}
		defer resp.Body.Close()
		return nil
	}); err != nil {
		t.Fatalf("could not connect to service B: %s", err)
	}

	// サービスAを起動（サービスBのURLを環境変数で渡す）
	serviceAResource, err := pool.BuildAndRunWithOptions(
		filepath.Join(pwd, "service-a", "Dockerfile"),
		&dockertest.RunOptions{
			Name:         "service-a",
			ExposedPorts: []string{"8080/tcp"},
			Env: []string{
				fmt.Sprintf("SERVICE_B_URL=%s", serviceBURL),
			},
		},
		func(config *docker.HostConfig) {
			config.AutoRemove = true
		},
	)
	if err != nil {
		t.Fatalf("could not start service A: %s", err)
	}
	defer pool.Purge(serviceAResource)

	serviceAPort := serviceAResource.GetPort("8080/tcp")
	serviceAURL := fmt.Sprintf("http://%s:%s", dockerHost, serviceAPort)

	// サービスAが起動するまで待機
	if err := pool.Retry(func() error {
		resp, err := http.Get(serviceAURL)
		if err != nil {
			return err
		}
		defer resp.Body.Close()
		return nil
	}); err != nil {
		t.Fatalf("could not connect to service A: %s", err)
	}

	// クライアントからサービスAにリクエスト
	client := &http.Client{}
	resp, err := client.Get(serviceAURL)
	if err != nil {
		t.Fatalf("could not make request to service A: %s", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		t.Fatalf("expected status 200, got %d", resp.StatusCode)
	}

	// レスポンスを確認
	var result map[string]interface{}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		t.Fatalf("could not decode response: %s", err)
	}

	// サービスAからの応答に、サービスBの情報が含まれていることを確認
	if result["service"] != "A" {
		t.Fatalf("expected service A, got %v", result["service"])
	}

	serviceBMsg, ok := result["service_b_msg"].(string)
	if !ok {
		t.Fatalf("service_b_msg not found in response")
	}

	if !strings.Contains(serviceBMsg, "service B") {
		t.Fatalf("expected service B message, got %s", serviceBMsg)
	}
}

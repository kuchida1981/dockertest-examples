package main

import (
	"bytes"
	"fmt"
	"log"
	"net/http"
	"strings"
	"testing"

	"github.com/ory/dockertest/v3"
	"github.com/ory/dockertest/v3/docker"
)

func TestStdout(t *testing.T) {
	pool, err := dockertest.NewPool("")
	if err != nil {
		log.Fatalf("could not construct pool: %s", err)
	}

	pool.Client.Ping()
	if err != nil {
		log.Fatalf("could not connect to Docker: %s", err)
	}

	resource, err := pool.Run("hello-world", "latest", []string{})
	if err != nil {
		log.Fatalf("could not start resource: %s", err)
	}

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
		log.Fatalf("could not get logs: %s", err)
	}

	output := logBuffer.String()
	// "Hello from Docker!" が含まれるか確認
	if !strings.Contains(output, "Hello from Docker!") {
		log.Fatalf("not contained 'Hello from Docker!' in output: %s", output)
	}

	// クリーンアップ
	if err := pool.Purge(resource); err != nil {
		log.Fatalf("could not purge resource: %s", err)
	}

}

func TestNginx(t *testing.T) {
	pool, err := dockertest.NewPool("")
	if err != nil {
		t.Fatalf("could not construct pool: %s", err)
	}

	if err := pool.Client.Ping(); err != nil {
		t.Fatalf("could not connect to Docker: %s", err)
	}

	// nginxコンテナを起動
	resource, err := pool.RunWithOptions(&dockertest.RunOptions{
		Repository: "nginx",
		Tag:        "latest",
	}, func(config *docker.HostConfig) {
		config.PortBindings = map[docker.Port][]docker.PortBinding{
			"80/tcp": {{HostPort: "8080"}},
		}
	})
	if err != nil {
		t.Fatalf("could not start resource: %s", err)
	}
	defer pool.Purge(resource)

	// nginxが起動するまで待機してリトライ
	if err := pool.Retry(func() error {
		resp, err := http.Get("http://localhost:8080")
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
	resp, err := http.Get("http://localhost:8080")
	if err != nil {
		t.Fatalf("could not make request: %s", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		t.Fatalf("expected status 200, got %d", resp.StatusCode)
	}
}

package tests

import (
    "bytes"
    "context"
    "errors"
    "fmt"
    "net/http"
    "os"
    "testing"
    "time"

    "github.com/EventStore/EventStore-Client-Go/v4/esdb"
    "github.com/testcontainers/testcontainers-go"
    "github.com/testcontainers/testcontainers-go/wait"
    "github.com/walletera/message-processor/eventstoredb"
    "github.com/walletera/message-processor/rabbitmq"
    "github.com/walletera/payments/internal/app"
)

const (
    mockserverPort                     = "2090"
    eventStoreDBHost                   = "127.0.0.1"
    eventStoreDBPort                   = "2113"
    eventStoreByCategoryProjectionPath = "/projection/$by_category/query?emit=yes"
    containersStartTimeout             = 30 * time.Second
    containersTerminationTimeout       = 10 * time.Second
)

func TestMain(m *testing.M) {
    ctx, _ := context.WithTimeout(context.Background(), containersStartTimeout)

    terminateEventSToreDBContainer, err := startEventStoreDBContainer(ctx)
    if err != nil {
        panic("error starting esdb container: " + err.Error())
    }

    err = setESDBByCategoryProjectionSeparator(ctx)
    if err != nil {
        panic(err.Error())
    }

    err = enableESDBByCategoryProjection(ctx)
    if err != nil {
        panic(err.Error())
    }

    err = createEventstoreDBPersistentSubscriptionForCategory(ctx, app.ESDB_ByCategoryProjection_Payments)
    if err != nil {
        panic(err.Error())
    }

    terminateRabbitMQContainer, err := startRabbitMQContainer(ctx)
    if err != nil {
        panic("error starting rabbitmq container: " + err.Error())
    }

    terminateMockserverContainer, err := startMockserverContainer(ctx)
    if err != nil {
        panic("error starting mockserver container:" + err.Error())
    }

    status := m.Run()

    err = terminateEventSToreDBContainer()
    if err != nil {
        panic("error terminating esdb container: " + err.Error())
    }

    err = terminateRabbitMQContainer()
    if err != nil {
        panic("error terminating rabbitmq container: " + err.Error())
    }

    err = terminateMockserverContainer()
    if err != nil {
        panic("error terminating mockserver container: " + err.Error())
    }

    os.Exit(status)
}

func startEventStoreDBContainer(ctx context.Context) (func() error, error) {
    req := testcontainers.ContainerRequest{
        Image: "eventstore/eventstore:21.10.7-buster-slim",
        Name:  "esdb-node",
        Cmd:   []string{"--insecure", "--run-projections=All"},
        ExposedPorts: []string{
            fmt.Sprintf("%s:%s", eventStoreDBPort, eventStoreDBPort),
        },
        WaitingFor: wait.
            ForHTTP("/health/live").
            WithPort("2113/tcp").
            WithStartupTimeout(10 * time.Second).
            WithStatusCodeMatcher(func(status int) bool {
                return status == http.StatusNoContent
            }),
        LogConsumerCfg: &testcontainers.LogConsumerConfig{
            Consumers: []testcontainers.LogConsumer{NewContainerLogConsumer("esdb")},
        },
    }
    esdbContainer, err := testcontainers.GenericContainer(ctx, testcontainers.GenericContainerRequest{
        ContainerRequest: req,
        Started:          true,
    })
    if err != nil {
        return nil, fmt.Errorf("error creating esdb container: %w", err)
    }

    return func() error {
        terminationCtx, terminationCtxCancel := context.WithTimeout(context.Background(), containersTerminationTimeout)
        defer terminationCtxCancel()
        terminationErr := esdbContainer.Terminate(terminationCtx)
        if terminationErr != nil {
            fmt.Errorf("failed terminating esdb container: %w", err)
        }
        return nil
    }, nil
}

func setESDBByCategoryProjectionSeparator(ctx context.Context) error {
    url := fmt.Sprintf("http://%s:%s%s", eventStoreDBHost, eventStoreDBPort, eventStoreByCategoryProjectionPath)
    body := []byte("last\n.")
    req, err := http.NewRequestWithContext(ctx, http.MethodPut, url, bytes.NewReader(body))
    if err != nil {
        return fmt.Errorf("failed creating request to update byCategory projection separator: %w", err)
    }
    req.Header.Set("Content-Type", "application/json; charset=utf-8")
    resp, err := http.DefaultClient.Do(req)
    if err != nil {
        return fmt.Errorf("failed creating request to update byCategory projection separator: %w", err)
    }
    if resp.StatusCode != http.StatusOK {
        return fmt.Errorf("failed creating request to update byCategory projection separator - response status code %d", resp.StatusCode)
    }
    return nil
}

func enableESDBByCategoryProjection(ctx context.Context) error {
    req, err := http.NewRequestWithContext(
        ctx,
        http.MethodPost,
        fmt.Sprintf("http://127.0.0.1:%s/projection/$by_category/command/enable", eventStoreDBPort),
        nil,
    )
    if err != nil {
        return fmt.Errorf("failed creating request for enabling $by_category projection: %w", err)
    }
    req.Header.Add("Accept", "application/json")
    req.Header.Add("Content-Length", "0")
    _, err = http.DefaultClient.Do(req)
    if err != nil {
        return fmt.Errorf("failed enabling $by_category projection: %w", err)
    }
    return nil
}

func createEventstoreDBPersistentSubscriptionForCategory(ctx context.Context, categoryName string) error {
    subscriptionSettings := esdb.SubscriptionSettingsDefault()
    subscriptionSettings.ResolveLinkTos = true
    subscriptionSettings.MaxRetryCount = 3

    esdbClient, err := eventstoredb.GetESDBClient(eventStoreDBUrl)
    if err != nil {
        return err
    }

    err = esdbClient.CreatePersistentSubscription(
        context.Background(),
        categoryName,
        app.ESDB_SubscriptionGroupName,
        esdb.PersistentStreamSubscriptionOptions{
            Settings: &subscriptionSettings,
        },
    )
    // FIXME: delete persistent subscription on the After hook
    if err != nil {
        var esdbError *esdb.Error
        ok := errors.As(err, &esdbError)
        if !ok || !esdbError.IsErrorCode(esdb.ErrorCodeResourceAlreadyExists) {
            return fmt.Errorf("CreatePersistentSubscription failed: %w", err)
        }
    }

    return nil
}

func startRabbitMQContainer(ctx context.Context) (func() error, error) {
    req := testcontainers.ContainerRequest{
        Image: "rabbitmq:3.8.0-management",
        Name:  "rabbitmq",
        User:  "rabbitmq",
        ExposedPorts: []string{
            fmt.Sprintf("%d:%d", rabbitmq.DefaultPort, rabbitmq.DefaultPort),
            fmt.Sprintf("%d:%d", rabbitmq.ManagementUIPort, rabbitmq.ManagementUIPort),
        },
        WaitingFor: wait.NewExecStrategy([]string{"rabbitmqadmin", "list", "queues"}).WithStartupTimeout(20 * time.Second),
        LogConsumerCfg: &testcontainers.LogConsumerConfig{
            Consumers: []testcontainers.LogConsumer{NewContainerLogConsumer("rabbitmq")},
        },
    }
    rabbitmqC, err := testcontainers.GenericContainer(ctx, testcontainers.GenericContainerRequest{
        ContainerRequest: req,
        Started:          true,
    })
    if err != nil {
        return nil, fmt.Errorf("error creating rabbitmq container: %w", err)
    }

    return func() error {
        terminationCtx, terminationCtxCancel := context.WithTimeout(context.Background(), containersTerminationTimeout)
        defer terminationCtxCancel()
        terminationErr := rabbitmqC.Terminate(terminationCtx)
        if terminationErr != nil {
            fmt.Errorf("failed terminating rabbitmq container: %w", err)
        }
        return nil
    }, nil
}

func startMockserverContainer(ctx context.Context) (func() error, error) {
    req := testcontainers.ContainerRequest{
        Image: "mockserver/mockserver",
        Name:  "mockserver",
        Env: map[string]string{
            "MOCKSERVER_SERVER_PORT": mockserverPort,
            "MOCKSERVER_LOG_LEVEL":   "DEBUG",
        },
        ExposedPorts: []string{fmt.Sprintf("%s:%s", mockserverPort, mockserverPort)},
        WaitingFor:   wait.ForHTTP("/mockserver/status").WithMethod(http.MethodPut).WithPort(mockserverPort),
        LogConsumerCfg: &testcontainers.LogConsumerConfig{
            Consumers: []testcontainers.LogConsumer{NewContainerLogConsumer("mockserver")},
        },
    }
    mockserverC, err := testcontainers.GenericContainer(ctx, testcontainers.GenericContainerRequest{
        ContainerRequest: req,
        Started:          true,
    })
    if err != nil {
        return nil, fmt.Errorf("error creating mockserver container: %w", err)
    }

    return func() error {
        terminationCtx, terminationCtxCancel := context.WithTimeout(context.Background(), containersTerminationTimeout)
        defer terminationCtxCancel()
        terminationErr := mockserverC.Terminate(terminationCtx)
        if terminationErr != nil {
            fmt.Errorf("failed terminating mockserver container: %w", err)
        }
        return nil
    }, nil
}

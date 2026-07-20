package client

import (
	"context"
	"crypto/tls"
	"time"

	kublingv1 "github.com/kubling-community/kubling-grpc/sdk-go/kubling/v1"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
	"google.golang.org/grpc/credentials/insecure"
)

type Client struct {
	conn          *grpc.ClientConn
	sessionClient kublingv1.SessionServiceClient
	queryClient   kublingv1.QueryServiceClient
	token         string
}

type Options struct {
	Address  string
	Username string
	Password string
	VDBName  string
}

func NewClient(options Options) (*Client, error) {

	conn, sessionClient, queryClient, err := newConnection(options.Address)
	if err != nil {
		return nil, err
	}

	loginCtx, cancel :=
		context.WithTimeout(
			context.Background(),
			30*time.Second,
		)
	defer cancel()

	login, err :=
		sessionClient.Login(
			loginCtx,
			&kublingv1.LoginRequest{
				Username: options.Username,
				Password: options.Password,
				VdbName:  options.VDBName,
			},
		)
	if err != nil {
		_ = conn.Close()
		return nil, err
	}

	return &Client{
		conn:          conn,
		sessionClient: sessionClient,
		queryClient:   queryClient,
		token:         login.GetExpiringToken(),
	}, nil

}

func newConnection(
	address string,
) (
	*grpc.ClientConn,
	kublingv1.SessionServiceClient,
	kublingv1.QueryServiceClient,
	error,
) {

	conn, err :=
		grpc.NewClient(
			address,
			grpc.WithTransportCredentials(
				credentials.NewTLS(
					&tls.Config{
						InsecureSkipVerify: true,
					},
				),
			),
		)
	if err != nil {
		return nil, nil, nil, err
	}

	sessionClient := kublingv1.NewSessionServiceClient(conn)

	if err := ping(sessionClient); err != nil {

		_ = conn.Close()

		conn, err =
			grpc.NewClient(
				address,
				grpc.WithTransportCredentials(
					insecure.NewCredentials(),
				),
			)
		if err != nil {
			return nil, nil, nil, err
		}

		sessionClient = kublingv1.NewSessionServiceClient(conn)

		if err := ping(sessionClient); err != nil {
			_ = conn.Close()
			return nil, nil, nil, err
		}

	}

	return conn,
		sessionClient,
		kublingv1.NewQueryServiceClient(conn),
		nil

}

func ping(
	sessionClient kublingv1.SessionServiceClient,
) error {

	ctx, cancel :=
		context.WithTimeout(
			context.Background(),
			3*time.Second,
		)
	defer cancel()

	_, err :=
		sessionClient.Ping(
			ctx,
			&kublingv1.PingRequest{},
		)

	return err

}

func (c *Client) Close() error {
	return c.conn.Close()
}

func (c *Client) Token() string {
	return c.token
}

func (c *Client) Query() kublingv1.QueryServiceClient {
	return c.queryClient
}

func (c *Client) Logout() error {

	ctx, cancel :=
		context.WithTimeout(
			context.Background(),
			30*time.Second,
		)
	defer cancel()

	_, err :=
		c.sessionClient.Logout(
			ctx,
			&kublingv1.LogoutRequest{
				ExpiringToken: c.token,
			},
		)

	return err

}

func (c *Client) IsInTransaction(
	ctx context.Context,
) (*kublingv1.IsInTransactionResponse, error) {

	return c.queryClient.IsInTransaction(
		ctx,
		&kublingv1.IsInTransactionRequest{
			ExpiringToken: c.token,
		},
	)

}

func NewClientWithToken(
	options Options,
	token string,
) (*Client, error) {

	conn, sessionClient, queryClient, err :=
		newConnection(
			options.Address,
		)
	if err != nil {
		return nil, err
	}

	return &Client{
		conn:          conn,
		sessionClient: sessionClient,
		queryClient:   queryClient,
		token:         token,
	}, nil

}

func (c *Client) QueryService() kublingv1.QueryServiceClient {
	return c.queryClient
}

func (c *Client) SessionService() kublingv1.SessionServiceClient {
	return c.sessionClient
}

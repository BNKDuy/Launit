package engine

import (
	"context"
	"errors"
	"fmt"
	"log"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/lambda"
	"github.com/aws/aws-sdk-go-v2/service/lambda/types"
)

type instanceSize string

const (
	XS instanceSize = "xs"
	S  instanceSize = "s"
	M  instanceSize = "m"
	L  instanceSize = "l"
)

func (s instanceSize) getSize() (int32, bool) {
	switch s {
	case XS:
		return 128, true
	case S:
		return 256, true
	case M:
		return 512, true
	case L:
		return 1024, true
	default:
		return 0, false
	}
}

const (
	MAX_MEMORY int32 = 10 << 10
	MIN_MEMORY int32 = 128
)

type LambdaService struct {
	client        *lambda.Client
	bucket        string
	executionRole string
}

func NewLambdaService(client *lambda.Client, bucket string, role string) *LambdaService {
	return &LambdaService{
		client:        client,
		bucket:        bucket,
		executionRole: role,
	}
}

var _ Engine = (*LambdaService)(nil)

func (l *LambdaService) Create(ctx context.Context, name string, size string, runtime string, timeout int32, binaryURI string) (string, error) {
	memory, valid := instanceSize(size).getSize()
	if !valid {
		return "", errors.New("invalid instance size")
	}

	function, err := l.client.CreateFunction(ctx, &lambda.CreateFunctionInput{
		FunctionName:  aws.String(name),
		Architectures: []types.Architecture{types.ArchitectureArm64},

		Role: aws.String(l.executionRole),

		Runtime:    types.RuntimeProvidedal2023,
		Timeout:    aws.Int32(timeout),
		MemorySize: aws.Int32(memory),
		Handler:    aws.String("bootstrap"),

		Code: &types.FunctionCode{
			S3Bucket: aws.String(l.bucket),
			S3Key:    aws.String(binaryURI),
		},
	})

	if err != nil {
		log.Println("CreateFunction failed: " + err.Error())
		return "", fmt.Errorf("Failed to create function")
	}

	// Wait up to 1 min for the function to create
	waiter := lambda.NewFunctionActiveV2Waiter(l.client)
	_, err = waiter.WaitForOutput(ctx, &lambda.GetFunctionInput{
		FunctionName: aws.String(name)}, 1*time.Minute)

	if err != nil {
		log.Printf("Couldn't wait for function %v to be active. Here's why: %v\n", name, err)
		return "", fmt.Errorf("Failed to create function")
	}

	// Create a Lambda for URL
	out, err := l.client.CreateFunctionUrlConfig(ctx, &lambda.CreateFunctionUrlConfigInput{
		AuthType:     types.FunctionUrlAuthTypeNone,
		FunctionName: function.FunctionName,
	})

	if err != nil {
		log.Println("Cannot create a lambda url" + err.Error())
		return "", fmt.Errorf("Failed to create function")
	}

	// Execution permission
	// Anyone can access
	_, err = l.client.AddPermission(ctx, &lambda.AddPermissionInput{
		FunctionName:        aws.String(name),
		StatementId:         aws.String("allow-public-url"),
		Action:              aws.String("lambda:InvokeFunctionUrl"),
		Principal:           aws.String("*"),
		FunctionUrlAuthType: types.FunctionUrlAuthTypeNone,
	})

	if err != nil {
		log.Println("Cannot set policy:" + err.Error())
		return "", fmt.Errorf("Failed to create function")
	}

	return *out.FunctionUrl, nil
}
func (l *LambdaService) Update(ctx context.Context, name string, binaryURI string) error { return nil }

func (l *LambdaService) Delete(ctx context.Context, name string) error {
	_, err := l.client.DeleteFunction(ctx, &lambda.DeleteFunctionInput{
		FunctionName: aws.String(name),
	})

	if err != nil {
		log.Println("Failed to delete function: " + err.Error())
		return fmt.Errorf("Failed to delete function")
	}

	return nil
}

func (l *LambdaService) List(ctx context.Context) ([]Function, error) {
	output, err := l.client.ListFunctions(ctx, &lambda.ListFunctionsInput{})
	if err != nil {
		log.Println("failed to list functions: ", err)
		return nil, fmt.Errorf("Internal Server Error")
	}

	out := make([]Function, len(output.Functions))

	for i, fn := range output.Functions {
		var name string
		if fn.FunctionName != nil {
			name = *fn.FunctionName
		}

		var memorySize int32
		if fn.MemorySize != nil {
			memorySize = *fn.MemorySize
		}

		var runtimeStr string
		runtimeStr = string(fn.Runtime)

		out[i] = Function{
			Name:       name,
			Runtime:    runtimeStr,
			MemorySize: memorySize,
			Status:     string(fn.LastUpdateStatus), // active, successful, or failed status
			Url:        "",
		}
	}

	return out, nil
}

@echo off
echo ================================================
echo  Generating proto files for both services
echo ================================================

REM Create output directories
mkdir order-service\internal\pb 2>nul
mkdir order-service\internal\orderpb 2>nul
mkdir payment-service\internal\pb 2>nul

REM Generate payment proto -> order-service (used as gRPC client)
echo Generating payment.proto for order-service...
protoc ^
  --go_out=order-service/internal ^
  --go_opt=paths=source_relative ^
  --go-grpc_out=order-service/internal ^
  --go-grpc_opt=paths=source_relative ^
  proto/payment.proto

REM Generate payment proto -> payment-service (used as gRPC server)
echo Generating payment.proto for payment-service...
protoc ^
  --go_out=payment-service/internal ^
  --go_opt=paths=source_relative ^
  --go-grpc_out=payment-service/internal ^
  --go-grpc_opt=paths=source_relative ^
  proto/payment.proto

REM Generate order proto -> order-service (streaming server)
echo Generating order.proto for order-service...
protoc ^
  --go_out=order-service/internal ^
  --go_opt=paths=source_relative ^
  --go-grpc_out=order-service/internal ^
  --go-grpc_opt=paths=source_relative ^
  proto/order.proto

echo ================================================
echo  Done! Generated files:
echo  - order-service/internal/pb/
echo  - order-service/internal/orderpb/
echo  - payment-service/internal/pb/
echo ================================================
